package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"planner-go/internal/api"
	"planner-go/internal/api/spec"
	"syscall"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGKILL)

	defer cancel()

	if err := run(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	fmt.Println("exiting...")
}

func run(ctx context.Context) error {
	cfg := zap.NewDevelopmentConfig()

	//colour for dev env
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, err := cfg.Build()

	if err != nil {
		return err
	}

	logger = logger.Named("planner_app")

	defer logger.Sync()

	pool, err := pgxpool.New(ctx, fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s",
		os.Getenv("PLANNER_DATABASE_USER"),
		os.Getenv("PLANNER_DATABASE_PASSWORD"),
		os.Getenv("PLANNER_DATABASE_HOST"),
		os.Getenv("PLANNER_DATABASE_PORT"),
		os.Getenv("PLANNER_DATABASE_NAME"),
	))

	if err != nil {
		return err
	}

	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return err
	}

	si := api.NewApi(pool, logger)
	r := chi.NewMux()
	r.Use(middleware.RequestID, middleware.Recoverer)
	r.Mount("/", spec.Handler(&si))

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	defer func() {
		const timeout = 30 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			logger.Error("Failed to shutdown the server", zap.Error(err))
		}
	}()

	errChannel := make(chan error, 1)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			errChannel <- err
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errChannel:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}
	return nil
}
