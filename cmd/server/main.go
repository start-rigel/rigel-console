package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/rigel-labs/rigel-console/internal/app"
	"github.com/rigel-labs/rigel-console/internal/client/buildengine"
	"github.com/rigel-labs/rigel-console/internal/config"
	consoleservice "github.com/rigel-labs/rigel-console/internal/service/console"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	buildClient := buildengine.New(cfg.BuildEngineBaseURL)
	consoleService := consoleservice.New(buildClient)
	application := app.New(cfg, consoleService)
	server := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      application.Handler(),
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("shutdown server: %v", err)
		}
	}()

	log.Printf("starting %s on :%s", cfg.ServiceName, cfg.HTTPPort)
	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server exited: %v", err)
	}
}
