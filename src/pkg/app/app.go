

package app

import (
	"context"
	"log/slog"
	"runtime/debug"
	"time"

	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/internal"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/observability"
	"github.com/amazon-gamelift/amazon-gamelift-servers-game-server-wrapper/pkg/runner"
	"github.com/pkg/errors"
)


type Service struct {
	logger  *slog.Logger          
	runner  *runner.Runner      
	spanner observability.Spanner 
}


func (service *Service) Run(ctx context.Context) error {
	defer func() {
		if err := recover(); err != nil {
			service.logger.ErrorContext(ctx, "App panic detected", "err", err, "stack", string(debug.Stack()))
		}
	}()

	service.logger.InfoContext(ctx, "Starting game server wrapper application", "version", internal.Version())

	service.logger.DebugContext(ctx, "Initializing game server runner")
	ctx, span, _ := service.spanner.NewSpan(ctx, "runner run", nil)
	if err := service.runner.Run(ctx); err != nil {
		return errors.Wrapf(err, "Game server execution failed")
	}
	span.End()

	service.logger.DebugContext(ctx, "Game server runner completed successfully")

	<-time.After(time.Second)

	return nil
}



func New(l *slog.Logger, rnr *runner.Runner, sp observability.Spanner) *Service {
	s := &Service{
		logger:  l,
		runner:  rnr,
		spanner: sp,
	}
	return s
}
