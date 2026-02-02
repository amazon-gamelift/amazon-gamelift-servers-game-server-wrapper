/*
 * Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
 * SPDX-License-Identifier: Apache-2.0
 */

package process

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

// Result represents the outcome of a process execution.
type Result struct {
	ReturnCode int
	Signal     os.Signal
}

// State represents the current state of a process.
type State struct {
	Exited bool
	Pid    int
}

// Args contains the arguments and I/O configuration for process execution.
type Args struct {
	CliArgs []string
	Stdout  io.Writer
	Stderr  io.Writer
}

// Process defines the interface for managing the process lifecycle.
type Process interface {
	Init(ctx context.Context) error
	Run(ctx context.Context, args *Args, pidChan chan<- int) (*Result, error)
	State() *State
}

type process struct {
	cfg     *Config
	exePath string
	cmd     *exec.Cmd
	logger  *slog.Logger
}

func (process *process) State() *State {
	state := &State{}

	if process.cmd == nil {
		return state
	}

	if process.cmd.ProcessState != nil {
		state.Exited = process.cmd.ProcessState.Exited()
	}

	if process.cmd.Process != nil {
		state.Pid = process.cmd.Process.Pid
	}

	return state
}

func (process *process) Init(ctx context.Context) error {
	if process.cfg == nil {
		return errors.New("Process configuration is nil")
	}

	if !filepath.IsAbs(process.cfg.ExeName) {
		process.exePath = filepath.Join(process.cfg.WorkingDirectory, process.cfg.ExeName)
	} else {
		process.exePath = process.cfg.ExeName
	}

	fi, err := os.Stat(process.exePath)
	if err != nil {
		return errors.Wrapf(err, "Failed to access executable '%s'", process.exePath)
	}

	return ensureExecutable(fi, process.exePath)
}

func (process *process) Run(ctx context.Context, args *Args, pidChan chan<- int) (*Result, error) {
	res := &Result{
		ReturnCode: -1,
	}

	process.logger.DebugContext(ctx, "Preparing command", "path", process.exePath, "workingDir", process.cfg.WorkingDirectory)

	process.cmd = exec.CommandContext(ctx, process.exePath, args.CliArgs...)
	process.cmd.Stderr = args.Stderr
	process.cmd.Stdout = args.Stdout
	process.cmd.Dir = process.cfg.WorkingDirectory

	if process.cfg.EnvVars != nil {
		// Preserve parent environment and overlay configured variables
		base := os.Environ()
		envMap := make(map[string]string, len(base)+len(process.cfg.EnvVars))
		for _, kv := range base {
			for i := 0; i < len(kv); i++ {
				if kv[i] == '=' {
					envMap[kv[:i]] = kv[i+1:]
					break
				}
			}
		}
		for k, v := range process.cfg.EnvVars {
			envMap[k] = v
		}
		env := make([]string, 0, len(envMap))
		for k, v := range envMap {
			env = append(env, fmt.Sprintf("%s=%s", strings.ToUpper(k), v))
		}
		process.cmd.Env = env
	}

	process.logger.InfoContext(ctx, "Starting process", "path", process.exePath, "args", args)
	if process.cfg.DelayStart != "" {
		process.logger.InfoContext(ctx, "DelayStart requested", "delay", process.cfg.DelayStart)
		d, err := time.ParseDuration(process.cfg.DelayStart)
		if err != nil {
			process.logger.WarnContext(ctx, "Unable to parse duration, defaulting to 10s")
			time.Sleep(time.Duration(10) * time.Second)
		} else {
			time.Sleep(d)
		}
	}
	err := process.cmd.Start()
	if err != nil {
		return res, err
	}

	if pidChan != nil {
		go func() {
			pidChan <- process.cmd.Process.Pid
		}()
	}

	err = process.cmd.Wait()

	var ee *exec.ExitError
	if errors.As(err, &ee) {
		ws := ee.Sys().(syscall.WaitStatus)
		res.Signal = ws.Signal()
		// it's been killed by either the context or by an external pid termination command
		if ws.Signal() == syscall.SIGKILL {
			process.logger.DebugContext(ctx, "Process terminated by signal",
				"signal", "SIGKILL")
			err = nil
		}
	}

	process.logger.InfoContext(ctx, "Process finished", "err", err)

	if process.cmd.ProcessState != nil {
		res.ReturnCode = process.cmd.ProcessState.ExitCode()
	}

	return res, err
}

// Config contains the configuration for a process.
type Config struct {
	ExeName          string
	WorkingDirectory string
	EnvVars          map[string]string
	DelayStart       string
}

// New creates a new Process instance with the provided configuration and logger.
//
// Parameters:
//   - cfg: Process configuration
//   - logger: Logger for process operations
//
// Returns:
//   - Process: Configured process instance
func New(cfg *Config, logger *slog.Logger) Process {
	process := &process{
		cfg:    cfg,
		logger: logger,
	}

	return process
}
