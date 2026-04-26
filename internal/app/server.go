package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func (a *App) Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErr := make(chan error, 1)

	go func() {
		a.logger.Info("starting http server",
			zap.String("addr", a.cfg.HTTPAddr),
			zap.String("storage", a.cfg.StorageType),
		)

		err := a.server.ListenAndServe()
		if errors.Is(err, http.ErrServerClosed) {
			err = nil
		}

		serverErr <- err
		close(serverErr)
	}()

	select {
	case <-ctx.Done():
		a.logger.Info("shutdown signal received")

		if err := a.shutdown(); err != nil {
			return err
		}

		if err := <-serverErr; err != nil {
			return fmt.Errorf("http server failed after shutdown: %w", err)
		}

		return nil

	case err := <-serverErr:
		a.closeResources()

		if err != nil {
			return fmt.Errorf("http server failed: %w", err)
		}

		return nil
	}
}
