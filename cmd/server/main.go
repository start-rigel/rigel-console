package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/rigel-labs/rigel-console/internal/app"
	"github.com/rigel-labs/rigel-console/internal/client/buildengine"
	"github.com/rigel-labs/rigel-console/internal/client/jdcollector"
	"github.com/rigel-labs/rigel-console/internal/config"
	consoleservice "github.com/rigel-labs/rigel-console/internal/service/console"
)

func main() {
	configPath := flag.String("config", config.DefaultPath(), "path to YAML config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	buildClient := buildengine.New(cfg.BuildEngineBaseURL, cfg.BuildEngineToken)
	jdCollectorClient := jdcollector.New(cfg.JDCollectorBaseURL)
	opts := []consoleservice.Option{
		consoleservice.WithLimits(cfg.IPHourlyLimit, cfg.DeviceHourlyLimit),
		consoleservice.WithChallengePassTTL(time.Duration(cfg.ChallengePassSeconds) * time.Second),
		consoleservice.WithSessionTTL(time.Duration(cfg.SessionTTLSeconds) * time.Second),
		consoleservice.WithChallengeVerifier(consoleservice.NewChallengeVerifier(cfg.ChallengeProvider, cfg.ChallengeVerifyURL, cfg.ChallengeSecret)),
	}
	if cfg.RedisAddr != "" {
		store, err := consoleservice.NewRedisSecurityStore(cfg.RedisAddr)
		if err != nil {
			log.Fatalf("init redis security store: %v", err)
		}
		opts = append(opts, consoleservice.WithStore(store))
	}
	consoleService := consoleservice.New(
		buildClient,
		jdCollectorClient,
		cfg.AdminUsername,
		cfg.AdminPassword,
		cfg.AnonymousHourlyLimit,
		time.Duration(cfg.CooldownSeconds)*time.Second,
		opts...,
	)
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
