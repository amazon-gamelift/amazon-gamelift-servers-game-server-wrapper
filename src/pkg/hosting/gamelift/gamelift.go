/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package gamelift

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/internal"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/constants"
	"github.com/amazon-gamelift/amazon-gamelift-servers-go-server-sdk/model"
	"github.com/amazon-gamelift/amazon-gamelift-servers-go-server-sdk/server"

	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/config"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/hosting"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/hosting/gamelift/initialiser"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/hosting/gamelift/platform"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/hosting/gamelift/sdk"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/observability"
	"github.com/Engine-Room-VR/amazon-gamelift-servers-game-server-wrapper/pkg/types/events"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

type gamelift struct {
	logger           *slog.Logger
	spanner          observability.Spanner
	sdk              sdk.GameLiftSdk
	ctx              context.Context
	cfg              *Config
	ec               chan error
	logDir           string
	runId            uuid.UUID
	gameServerLogDir string

	onHealthCheck      func(ctx context.Context) events.GameStatus
	onHostingStart     func(ctx context.Context, h *events.HostingStart, end <-chan error) error
	onHostingTerminate func(ctx context.Context, h *events.HostingTerminate) error
	onError            func(err error)
	init               initialiser.Service
}

type Config struct {
	GamePort               int
	Anywhere               config.Anywhere // Contains configuration for GameLift Anywhere fleet
	LogDirectory           string          // Specifies the directory for general logging
	GameServerLogDirectory string          // Specifies the directory for game server specific logs
}

// Init initializes the Amazon GameLift SDK with the provided configuration.
//
// Parameters:
//   - ctx: Context for the initialization process
//   - args: Initialization arguments
//
// Returns:
//   - *hosting.InitMeta: Metadata about the initialized hosting option
//   - error: Any error that occurred during initialization
func (gameLift *gamelift) Init(ctx context.Context, args *hosting.InitArgs) (*hosting.InitMeta, error) {
	gameLift.ctx = ctx

	ctx, span, _ := gameLift.spanner.NewSpan(ctx, "Amazon GameLift init", nil)
	defer span.End()

	gameLift.logDir = ctx.Value(constants.ContextKeyRunLogDir).(string)
	if len(gameLift.cfg.GameServerLogDirectory) != 0 {
		gameLift.gameServerLogDir = gameLift.cfg.GameServerLogDirectory
	}

	gameLift.runId = args.RunId

	if err := gameLift.init.InitSdk(ctx); err != nil {
		return nil, errors.Wrap(err, "failed to initialize Amazon GameLift")
	}

	meta := &hosting.InitMeta{
		InstanceWorkingDirectory: platform.InstancePath(),
	}

	return meta, nil
}

// Run starts the server process and establishes communication with the Amazon GameLift service.
//
// Parameters:
//   - ctx: Context for managing the server process lifecycle
//
// Returns:
//   - error: Any error that occurred during the server process execution
func (gameLift *gamelift) Run(ctx context.Context) error {
	gameLift.ctx = ctx

	ctx, span, _ := gameLift.spanner.NewSpan(ctx, "Amazon GameLift run", nil)
	defer span.End()

	logPaths := make([]string, 0)
	if len(gameLift.logDir) != 0 {
		gameLift.logger.DebugContext(ctx, "configuring logging directory", "dir", gameLift.logDir)
		logPaths = append(logPaths, gameLift.logDir)
	} else {
		gameLift.logger.WarnContext(ctx, "no log directory specified - no logs will be saved to Amazon GameLift")
	}

	// Set SDK tool name and version before calling ProcessReady
	err := os.Setenv(constants.EnvironmentKeySDKToolName, internal.AppName())
	if err != nil {
		return errors.Wrapf(err, "unable to set SDKToolName environment variable")
	}

	err = os.Setenv(constants.EnvironmentKeySDKToolVersion, internal.SemVer())
	if err != nil {
		return errors.Wrapf(err, "unable to set SDKToolVersion environment variable")
	}

	err = gameLift.sdk.ProcessReady(ctx, server.ProcessParameters{
		Port: gameLift.cfg.GamePort,
		LogParameters: server.LogParameters{
			LogPaths: logPaths,
		},
		OnHealthCheck:       gameLift.glHealthcheck,
		OnProcessTerminate:  gameLift.glOnProcessTerminate,
		OnStartGameSession:  gameLift.glOnStartGameSession,
		OnUpdateGameSession: gameLift.glOnUpdateGameSession,
	})

	if err != nil {
		return errors.Wrapf(err, "failed to call process ready")
	}

	select {
	case <-ctx.Done():
		return nil
	case err := <-gameLift.ec:
		gameLift.logger.WarnContext(ctx, "error returned from Amazon GameLift hosting", "err", err)
		return err
	}
}

// SetOnHostingStart registers a callback function that will be invoked when a hosting
// start event occurs. The callback receives the hosting start event and an error channel.
func (gameLift *gamelift) SetOnHostingStart(f func(ctx context.Context, h *events.HostingStart, end <-chan error) error) {
	gameLift.onHostingStart = f
}

// SetOnHostingTerminate registers a callback function that will be invoked when a hosting
// termination event occurs.
func (gameLift *gamelift) SetOnHostingTerminate(f func(ctx context.Context, h *events.HostingTerminate) error) {
	gameLift.onHostingTerminate = f
}

// SetOnHealthCheck registers a callback function that will be invoked to check the health
// status of the game server. The callback should return the current game status.
func (gameLift *gamelift) SetOnHealthCheck(f func(ctx context.Context) events.GameStatus) {
	gameLift.onHealthCheck = f
}

// Close performs a shutdown of the server process. The cleanup includes copying game
// server logs to the designated log directory, notifying GameLift SDK that the process
// is ending, destroying the SDK resources, and closing the initializer service.
func (gameLift *gamelift) Close(ctx context.Context) error {
	gameLift.logger.InfoContext(ctx, "cleaning up Amazon GameLift resources")

	copyErr := copyGameServerLogs(gameLift.logger, gameLift.gameServerLogDir, gameLift.logDir)

	if copyErr != nil {
		gameLift.ec <- copyErr
	}

	var err error

	if e := gameLift.sdk.ProcessEnding(ctx); e != nil {
		gameLift.logger.ErrorContext(ctx, "failed to call process ending", "err", e)
		err = e
	}

	if e := gameLift.sdk.Destroy(ctx); e != nil {
		gameLift.logger.ErrorContext(ctx, "failed to call destroy", "err", e)
		if err != nil {
			err = errors.Wrapf(err, e.Error())
		} else {
			err = e
		}
	}

	return err
}

func (gameLift *gamelift) glHealthcheck() bool {
	res := gameLift.onHealthCheck(gameLift.ctx)
	switch res {
	case events.GameStatusWaiting:
		fallthrough
	case events.GameStatusRunning:
		return true

	case events.GameStatusErrored:
		fallthrough
	case events.GameStatusFinished:
		return false

	default:
		return false
	}
}

func (gameLift *gamelift) glOnProcessTerminate() {
	err := gameLift.onHostingTerminate(gameLift.ctx, &events.HostingTerminate{
		Reason: events.HostingTerminateReasonHostingShutdown,
	})

	if err != nil {
		gameLift.ec <- err
	}
}

func (gameLift *gamelift) glOnUpdateGameSession(gs model.UpdateGameSession) {
	gameLift.logger.DebugContext(gameLift.ctx, "update game session called, but no action is taken by the Game Server Wrapper")
}

func (gameLift *gamelift) glOnError(err error) {
	gameLift.logger.Error("Amazon GameLift error", "err", err)
	err = gameLift.Close(gameLift.ctx)
	if err != nil {
		gameLift.logger.Error("Error closing game after Amazon GameLift error.", "err", err)
		return
	}
}

func (gameLift *gamelift) glOnStartGameSession(gs model.GameSession) {

	ctx, span, _ := gameLift.spanner.NewSpan(gameLift.ctx, "Amazon GameLift OnStartGameSession", nil)
	gameLift.ctx = ctx
	defer span.End()

	gameLift.logger.DebugContext(gameLift.ctx, "start game sessions called", "gs", gs)

	gameLift.logger.DebugContext(gameLift.ctx, "manager onHostingStart", "event", gs)
	cliArgs := make([]config.CliArg, 0)

	hse := &events.HostingStart{
		DNSName:                   gs.DNSName,
		CliArgs:                   cliArgs,
		FleetId:                   gs.FleetID,
		GamePort:                  gs.Port,
		GameProperties:            gs.GameProperties,
		GameSessionData:           gs.GameSessionData,
		GameSessionId:             gs.GameSessionID,
		GameSessionName:           gs.Name,
		IpAddress:                 gs.IPAddress,
		LogDirectory:              gameLift.logDir,
		MatchmakerData:            gs.MatchmakerData,
		MaximumPlayerSessionCount: gs.MaximumPlayerSessionCount,
		Provider:                  config.ProviderGameLift,
	}

	if strings.HasPrefix(hse.FleetId, "containerfleet-") {
		hse.ContainerPort = gameLift.cfg.GamePort
	}

	if err := gameLift.sdk.ActivateGameSession(gameLift.ctx); err != nil {
		gameLift.ec <- err
		return
	}

	gameLift.logger.DebugContext(gameLift.ctx, "calling onHostingStart")
	if err := gameLift.onHostingStart(gameLift.ctx, hse, nil); err != nil {
		gameLift.ec <- err
	}

}

type InitialiserServiceFactory interface {
	GetService(ctx context.Context, anywhere config.Anywhere, gameLiftSdk sdk.GameLiftSdk, logger *slog.Logger) (initialiser.Service, error)
}

func New(ctx context.Context, cfg *Config, logger *slog.Logger, spanner observability.Spanner, initialiserServiceFactory InitialiserServiceFactory, gameLiftSdk sdk.GameLiftSdk) (*gamelift, error) {

	init, err := initialiserServiceFactory.GetService(ctx, cfg.Anywhere, gameLiftSdk, logger)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create Amazon GameLift initialiser")
	}

	// the upper bound excludes the max port in case query port is not set
	if cfg.GamePort <= 0 || cfg.GamePort >= 65535 {
		return nil, errors.Errorf("game port needs to be a valid port: '%d'", cfg.GamePort)
	}

	g := &gamelift{
		cfg:     cfg,
		ec:      make(chan error),
		sdk:     gameLiftSdk,
		init:    init,
		logger:  logger,
		spanner: spanner,
	}

	return g, nil
}

func copyGameServerLogs(logger *slog.Logger, gameServerLogDir string, logDir string) error {
	if gameServerLogDir == "" {
		logger.Info("gameServerLogDir empty, not copying game server log")
		return nil
	}
	// Check if source directory exists
	if _, err := os.Stat(gameServerLogDir); os.IsNotExist(err) {
		logger.Info("gameServerLogDir does not exist, skipping copy", "dir", gameServerLogDir)
		return nil
	}

	var cpString string
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cpString = fmt.Sprintf("robocopy %v %v /E /NFL /NDL /NJH /NJS /NC /NS", gameServerLogDir, logDir)
		cmd = exec.Command("cmd", "/c", cpString)
	default:
		cpString = fmt.Sprintf("cp -Rf %v %v", gameServerLogDir, logDir)
		cmd = exec.Command("sh", "-c", cpString)
	}

	var stderr bytes.Buffer
	cmd.Stdout = os.Stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		return nil
	}
	stderrStr := strings.TrimSpace(stderr.String())
	// If on Windows, allow robocopy warnings (exit codes < 8)
	if runtime.GOOS == "windows" {
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() < 8 {
			return nil
		}
	}
	return fmt.Errorf("error copying source '%v' into destination '%v', error: %v, stderr: %v", gameServerLogDir, logDir, err, stderrStr)
}
