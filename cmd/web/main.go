package main

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/jusunglee/leagueofren/internal/anthropic"
	"github.com/jusunglee/leagueofren/internal/db/postgres"
	"github.com/jusunglee/leagueofren/internal/google"
	"github.com/jusunglee/leagueofren/internal/llm"
	"github.com/jusunglee/leagueofren/internal/logger"
	"github.com/jusunglee/leagueofren/internal/riot"
	"github.com/jusunglee/leagueofren/internal/translation"
	"github.com/jusunglee/leagueofren/internal/web"
	"github.com/peterbourgon/ff/v4"
	"github.com/peterbourgon/ff/v4/ffhelp"
)

//go:embed all:dist
var staticFiles embed.FS

func main() {
	if err := mainE(); err != nil {
		slog.Error("fatal", "error", err)
		os.Exit(1)
	}
	slog.Info("exiting without error")
}

func mainE() error {
	_ = godotenv.Load()

	fs_ := ff.NewFlagSet("leagueofren-web")

	var (
		port            = fs_.Int64Long("port", 3000, "HTTP server port")
		databaseURL     = fs_.StringLong("database-url", "", "PostgreSQL connection URL")
		riotAPIKey      = fs_.StringLong("riot-api-key", "", "Riot API key for username validation")
		llmProvider     = fs_.StringEnumLong("llm-provider", "LLM provider for server-side translation", "anthropic", "google")
		llmModel        = fs_.StringLong("llm-model", "", "LLM model name")
		anthropicAPIKey = fs_.StringLong("anthropic-api-key", "", "Anthropic API key")
		googleAPIKey    = fs_.StringLong("google-api-key", "", "Google API key")
	)

	if err := ff.Parse(fs_, os.Args[1:], ff.WithEnvVars()); err != nil {
		fmt.Printf("%s\n", ffhelp.Flags(fs_))
		return fmt.Errorf("parsing flags: %w", err)
	}

	if *databaseURL == "" {
		return errors.New("database-url is required")
	}
	if *riotAPIKey == "" {
		return errors.New("riot-api-key is required")
	}
	if *llmModel == "" {
		return errors.New("llm-model is required")
	}

	log := logger.New()

	var llmClient llm.Client
	switch *llmProvider {
	case "anthropic":
		if *anthropicAPIKey == "" {
			return errors.New("anthropic-api-key is required when using anthropic provider")
		}
		llmClient = anthropic.NewClient(*anthropicAPIKey, anthropic.Model(*llmModel))
	case "google":
		if *googleAPIKey == "" {
			return errors.New("google-api-key is required when using google provider")
		}
		var err error
		llmClient, err = google.NewClient(context.Background(), *googleAPIKey, google.Model(*llmModel))
		if err != nil {
			return fmt.Errorf("creating Google client: %w", err)
		}
	}

	ctx, cancel := context.WithCancelCause(context.Background())

	repo, err := postgres.New(ctx, *databaseURL)
	if err != nil {
		return fmt.Errorf("creating PostgreSQL connection: %w", err)
	}
	defer repo.Close()
	log.InfoContext(ctx, "connected to PostgreSQL database")

	riotClient := riot.NewDirectClient(*riotAPIKey)
	translator := translation.NewTranslator(llmClient, repo, *llmProvider, *llmModel)

	router := web.NewRouter(repo, log, riotClient, translator)
	apiHandler := router.Handler()

	// Serve API routes first, fall back to embedded static files for the SPA
	distFS, err := fs.Sub(staticFiles, "dist")
	if err != nil {
		return fmt.Errorf("creating sub filesystem: %w", err)
	}
	fileServer := http.FileServer(http.FS(distFS))

	mux := http.NewServeMux()
	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// API routes go to the router
		if strings.HasPrefix(r.URL.Path, "/api/") {
			apiHandler.ServeHTTP(w, r)
			return
		}

		// Try to serve a static file
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		if _, err := fs.Stat(distFS, strings.TrimPrefix(path, "/")); err == nil {
			fileServer.ServeHTTP(w, r)
			return
		}

		// SPA fallback: serve index.html for any unmatched path
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	}))

	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", *port),
		Handler:           mux,
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
