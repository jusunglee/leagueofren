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

	"github.com/joho/godotenv"
	"github.com/jusunglee/leagueofren/internal/db/postgres"
	"github.com/jusunglee/leagueofren/internal/logger"
	"github.com/jusunglee/leagueofren/internal/web"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
)

func main() {
	if err := mainE(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
	slog.Info("exiting without error")
}

func mainE() error {
	_ = godotenv.Load()

	fs := ff.NewFlagSet("leagueofren-web")

	var (
		port          = fs.Int64Long("port", 3000, "HTTP server port")
		databaseURL   = fs.StringLong("database-url", "", "PostgreSQL connection URL")
		adminPassword = fs.StringLong("admin-password", "admin", "Admin panel password")
	)

	if err := ff.Parse(fs, os.Args[1:], ff.WithEnvVars()); err != nil {
		fmt.Printf("%s\n", ffhelp.Flags(fs))
		return fmt.Errorf("parsing flags: %w", err)
	}

	if *databaseURL == "" {
		return errors.New("database-url is required")
	}

	log := logger.New()

	ctx, cancel := context.WithCancelCause(context.Background())

	repo, err := postgres.New(ctx, *databaseURL)
	if err != nil {
		return fmt.Errorf("creating PostgreSQL connection: %w", err)
	}
	defer repo.Close()
	log.InfoContext(ctx, "connected to PostgreSQL database")

	router := web.NewRouter(repo, log, web.Config{
		AdminPassword: *adminPassword,
	})

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", *port),
		Handler:           router.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.InfoContext(ctx, "received signal, shutting down gracefully", "signal", sig)
		cancel(errors.New("signal received"))

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.ErrorContext(ctx, "server shutdown error", "error", err)
		}
	}()

	log.InfoContext(ctx, "starting web server", "port", *port)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
