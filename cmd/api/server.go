package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func (app *application) serve() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", app.config.port),
		Handler:      app.routes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
	}

	var metricsSrv *http.Server
	if app.config.metrics.port > 0 {
		metricsSrv = &http.Server{
			Addr:         fmt.Sprintf(":%d", app.config.metrics.port),
			Handler:      app.metricsRoutes(),
			IdleTimeout:  time.Minute,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 10 * time.Second,
			ErrorLog:     slog.NewLogLogger(app.logger.Handler(), slog.LevelError),
		}
	}

	shutdownError := make(chan error, 2)

	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		s := <-quit

		app.logger.Info("shutting down server", "signal", s.String())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			shutdownError <- err
			return
		}

		if metricsSrv != nil {
			if err := metricsSrv.Shutdown(ctx); err != nil {
				shutdownError <- err
				return
			}
		}

		app.logger.Info("completing background tasks", "addr", srv.Addr)

		app.wg.Wait()
		shutdownError <- nil
	}()

	app.logger.Info("starting server", "addr", srv.Addr, "env", app.config.env)

	if metricsSrv != nil {
		go func() {
			app.logger.Info("starting metrics server", "addr", metricsSrv.Addr)
			err := metricsSrv.ListenAndServe()
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				shutdownError <- err
			}
		}()
	}

	err := srv.ListenAndServe()
	if !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	err = <-shutdownError
	if err != nil {
		return err
	}

	app.logger.Info("stopped server", "addr", srv.Addr)
	if metricsSrv != nil {
		app.logger.Info("stopped metrics server", "addr", metricsSrv.Addr)
	}

	return nil
}
