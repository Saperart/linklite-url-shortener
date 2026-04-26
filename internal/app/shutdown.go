package app

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func (a *App) shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), a.cfg.ShutdownTimeout)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		a.closeResources()
		return fmt.Errorf("shutdown http server: %w", err)
	}

	a.closeResources()

	a.logger.Info("shutdown complete", zap.Duration("timeout", a.cfg.ShutdownTimeout))
	return nil
}
