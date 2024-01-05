package application

import (
	"context"
	"fmt"
	"net/http"

	"github.com/redis/go-redis/v9"
)

type App struct {
	router http.Handler
	db     *redis.Client
}

func New() *App {
	app := &App{
		router: loadRouter(),
		db:     redis.NewClient(&redis.Options{}),
	}

	return app
}

func (a *App) connectToDb(ctx context.Context) error {
	if err := a.db.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}
	return nil
}

func (a *App) startServer() error {
	server := &http.Server{
		Addr:    ":3000",
		Handler: a.router,
	}

	if err := server.ListenAndServe(); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	return nil
}

func (a *App) Start(ctx context.Context) error {
	if err := a.connectToDb(ctx); err != nil {
		return err
	}

	if err := a.startServer(); err != nil {
		return err
	}

	return nil
}
