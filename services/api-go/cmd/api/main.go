// Command api is the TEP HTTP API server.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Bas-travel/tep/services/api-go/internal/api"
	"github.com/Bas-travel/tep/services/api-go/internal/config"
	"github.com/Bas-travel/tep/services/api-go/internal/db"
	"github.com/Bas-travel/tep/services/api-go/internal/httpx"
	"github.com/Bas-travel/tep/services/api-go/migrations"
)

// healthcheck probes the local liveness endpoint (used by Docker HEALTHCHECK; distroless has no curl).
func healthcheck() int {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	if addr[0] == ':' {
		addr = "127.0.0.1" + addr
	}
	c := http.Client{Timeout: 2 * time.Second}
	resp, err := c.Get("http://" + addr + "/health/live")
	if err != nil || resp.StatusCode != 200 {
		return 1
	}
	return 0
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "-healthcheck" {
		os.Exit(healthcheck())
	}
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Error("invalid configuration", "err", err)
		os.Exit(2)
	}
	checks := map[string]api.Check{}
	var opts []api.Option
	if cfg.MySQLDSN != "" {
		d, err := db.Open(cfg.MySQLDSN)
		if err != nil {
			log.Error("mysql unavailable", "err", err)
			os.Exit(1)
		}
		defer d.Close()
		if os.Getenv("MIGRATE_ON_START") == "true" {
			seed := os.Getenv("MIGRATE_SEED") == "true" && cfg.AppEnv != "production"
			applied, err := db.Migrate(context.Background(), d, migrations.FS, seed)
			if err != nil {
				log.Error("migration failed", "err", err)
				os.Exit(1)
			}
			log.Info("migrations applied", "versions", applied)
		}
		checks["mysql"] = func(ctx context.Context) error { return d.PingContext(ctx) }
		if cfg.APIMode == "db" {
			opts = append(opts, api.WithRepo(api.NewSQLRepo(d)))
		}
	}
	if cfg.RedisAddr != "" {
		addr := cfg.RedisAddr
		checks["redis"] = func(ctx context.Context) error { return db.PingRedis(ctx, addr) }
	}
	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpx.Chain(api.New(cfg, log, checks, opts...).Handler(), log, cfg.AppEnv == "dev"),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
	}
	go func() {
		log.Info("listening", "addr", cfg.HTTPAddr, "env", cfg.AppEnv, "mode", cfg.APIMode, "version", cfg.Version)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("server error", "err", err)
			os.Exit(1)
		}
	}()
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
