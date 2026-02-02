package orchestration

import (
	"context"
	"encoding/json"
	"fmt"
	"interserverauth"
	"log/slog"
	"strings"
	"time"

	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/internal/config"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/helpers"
	"github.com/amazon-gamelift/amazon-gamelift-servers-go-server-sdk/v5/model"
	"github.com/google/uuid"
	"github.com/pkg/errors"
)

// Values set as per: https://docs.aws.amazon.com/gameliftservers/latest/developerguide/queue-events.html
const gameLiftVersion = "0"
const gameLiftQueuePlacementEvent = "GameLift Queue Placement Event"
const gameLiftSource = "aws.gamelift"
const authHeaderKey = "Authorization"

type Sender struct {
	requestHelper requestHelper
	cfg           *config.Orchestration
	region        string
	logger        *slog.Logger
	gameSession   *model.GameSession
	auth          *interserverauth.Handler
}

func NewSender(logger *slog.Logger, requestHelper requestHelper, cfg *config.Orchestration, region string) *Sender {
	return &Sender{
		requestHelper: requestHelper,
		cfg:           cfg,
		region:        region,
		logger:        logger,
		auth:          interserverauth.New(cfg.AuthHeaderPrefix, cfg.ClientId, cfg.ClientSecret, cfg.GetTokenUrl, requestHelper, logger),
	}
}

func (s *Sender) OnHostingTerminate(ctx context.Context) error {
	if !s.cfg.EmitCustomEvents {
		s.logger.DebugContext(ctx, "emitting events to the orchestration layer is disabled")
		return nil
	}

	s.logger.DebugContext(ctx, "emitting hosting terminate event to the orchestration layer")

	event := s.newEvent()
	event.Id = uuid.New().String()
	event.Detail = Detail{
		Type:           PlacementTerminated.String(),
		GameSessionArn: s.gameSession.GameSessionID,
		PlacementId:    s.gameSession.GameSessionID[strings.LastIndex(s.gameSession.GameSessionID, "/")+1:],
		StartTime:      time.Now(),
	}
	err := s.emitEvent(context.Background(), event)
	if err != nil {
		msg := fmt.Sprintf("Error emitting OnHostingTerminate event: %v", err)
		s.logger.Error(msg)
		return errors.Wrap(err, msg)
	}

	return nil
}

func (s *Sender) OnHealthCheck(ctx context.Context) error {
	if !s.cfg.EmitCustomEvents {
		s.logger.DebugContext(ctx, "emitting events to the orchestration layer is disabled")
		return nil
	}

	// This method is a placeholder.  It can be expanded if a requirement arises for messaging the game server on a health check.
	return nil
}

func (s *Sender) OnStartGameSession(ctx context.Context, gs model.GameSession) error {
	if !s.cfg.EmitCustomEvents {
		s.logger.DebugContext(ctx, "emitting events to the orchestration layer is disabled")
		return nil
	}

	s.logger.DebugContext(ctx, fmt.Sprintf("emitting start game session event to the orchestration layer (session=%s,fleet=%s)", gs.GameSessionID, gs.FleetID))

	s.gameSession = &gs
	event := s.newEvent()
	event.Id = gs.GameSessionID
	event.Detail = Detail{
		Type:              PlacementActive.String(),
		PlacementId:       gs.GameSessionID[strings.LastIndex(gs.GameSessionID, "/")+1:],
		Port:              gs.Port,
		GameSessionArn:    gs.GameSessionID,
		IpAddress:         gs.IPAddress,
		DnsName:           gs.DNSName,
		StartTime:         time.Now(),
		GameSessionRegion: gs.Location,
	}
	err := s.emitEvent(ctx, event)
	if err != nil {
		msg := fmt.Sprintf("Error emitting OnStartGameSession event: %v", err)
		s.logger.ErrorContext(ctx, msg)
		return errors.Wrap(err, msg)
	}

	return nil
}

func (s *Sender) OnUpdateGameSession(ctx context.Context, gs model.UpdateGameSession) error {
	s.logger.DebugContext(ctx, "OnUpdateGameSession called", slog.Any("gameSession", gs), slog.Any("orchestration config", s.cfg))
	if !s.cfg.EmitCustomEvents {
		s.logger.DebugContext(ctx, "emitting events to the orchestration layer is disabled")
		return nil
	}

	s.logger.DebugContext(ctx, fmt.Sprintf("emitting update game session event to the orchestration layer (sessionId=%s)", gs.GameSession.GameSessionID))

	isDormant := gs.GameSession.GameProperties["dormant"]
	if isDormant == "" {
		s.logger.DebugContext(ctx, "game property 'dormant' is missing")
	}

	placementType := PlacementActive.String()
	if isDormant == "true" {
		placementType = PlacementDormant.String()
	}

	event := s.newEvent()
	event.Id = gs.GameSession.GameSessionID
	event.Detail = Detail{
		Type:              placementType,
		PlacementId:       gs.GameSession.GameSessionID[strings.LastIndex(gs.GameSession.GameSessionID, "/")+1:],
		Port:              gs.GameSession.Port,
		GameSessionArn:    gs.GameSession.GameSessionID,
		IpAddress:         gs.GameSession.IPAddress,
		DnsName:           gs.GameSession.DNSName,
		StartTime:         time.Now(),
		GameSessionRegion: gs.GameSession.Location,
	}
	err := s.emitEvent(ctx, event)
	if err != nil {
		msg := fmt.Sprintf("Error emitting OnUpdateGameSession event: %v", err)
		s.logger.ErrorContext(ctx, msg)
		return errors.Wrap(err, msg)
	}

	return nil
}

func (s *Sender) newEvent() Event {
	return Event{
		Version:   gameLiftVersion,
		Type:      gameLiftQueuePlacementEvent,
		Source:    gameLiftSource,
		Account:   s.cfg.Account,
		Time:      time.Now(),
		Region:    s.region,
		Resources: s.cfg.Resources,
	}
}

func (s *Sender) emitEvent(ctx context.Context, event Event) error {

	if s == nil {
		return errors.New("sender is nil")
	}
	if s.auth == nil {
		return errors.New("sender.auth is nil")
	}
	if s.logger == nil {
		return errors.New("sender.logger is nil")
	}
	if s.requestHelper == nil {
		return errors.New("sender.requestHelper is nil")
	}
	if s.cfg == nil {
		return errors.New("sender.cfg is nil")
	}

	if !s.auth.IsValidAccessToken(ctx) {
		err := s.auth.AcquireAccessToken(ctx)
		if err != nil {
			msg := fmt.Sprintf("Error acquiring token: %v", err)
			s.logger.ErrorContext(ctx, msg)
			return errors.Wrap(err, msg)
		}
	}

	body, err := json.Marshal(event)
	if err != nil {
		msg := fmt.Sprintf("Error marshaling event: %v", err)
		s.logger.ErrorContext(ctx, msg)
		return errors.Wrap(err, msg)
	}

	request := helpers.HttpRequestDetails{
		Method:  s.cfg.Method,
		Url:     s.cfg.Url,
		Body:    string(body),
		Headers: map[string]string{s.cfg.HeaderKey: s.cfg.HeaderValue},
	}
	request.Headers[authHeaderKey] = s.auth.GetAuthHeader(ctx)

	_, err = s.requestHelper.Request(ctx, request)
	var ue = &helpers.UnauthorisedError{}
	if errors.As(err, &ue) {
		reacquireErr := s.auth.AcquireAccessToken(ctx)
		if reacquireErr != nil {
			msg := fmt.Sprintf("Error re-acquiring token: %v", reacquireErr)
			s.logger.ErrorContext(ctx, msg)
			return errors.Wrap(reacquireErr, msg)
		}

		request.Headers[authHeaderKey] = s.auth.GetAuthHeader(ctx)
		_, err = s.requestHelper.Request(ctx, request)
	} else if err != nil {
		msg := fmt.Sprintf("Error emitting event: %v", err)
		s.logger.ErrorContext(ctx, msg)
		return errors.Wrap(err, msg)
	}

	return nil
}

type requestHelper interface {
	Request(ctx context.Context, requestData helpers.HttpRequestDetails) (string, error)
}
