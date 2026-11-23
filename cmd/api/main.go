package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v3"

	httpadapter "github.com/mofortech/go-fiber-starter/internal/adapter/http"
	jwtadapter "github.com/mofortech/go-fiber-starter/internal/adapter/jwt"
	"github.com/mofortech/go-fiber-starter/internal/adapter/postgres"
	"github.com/mofortech/go-fiber-starter/internal/config"
	"github.com/mofortech/go-fiber-starter/internal/usecase"
	"github.com/mofortech/go-fiber-starter/migrations"
)

func main() {
	if err := run(); err != nil {
		slog.Error("aplikasi berhenti", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := postgres.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err := postgres.Migrate(ctx, pool, migrations.FS); err != nil {
		return err
	}

	tokens := jwtadapter.New(cfg.JWTSecret, cfg.JWTTTL)
	app := httpadapter.New(httpadapter.Deps{
		Auth:        usecase.NewAuth(postgres.NewUsers(pool), tokens),
		Notes:       usecase.NewNote(postgres.NewNotes(pool)),
		Tokens:      tokens,
		CORSOrigins: cfg.CORSOrigins,
		Ping:        pool.Ping,
	})
	slog.Info("server berjalan", "addr", cfg.Addr, "env", cfg.Env)
	return app.Listen(cfg.Addr, fiber.ListenConfig{GracefulContext: ctx, ShutdownTimeout: 10 * time.Second, DisableStartupMessage: true})
}
