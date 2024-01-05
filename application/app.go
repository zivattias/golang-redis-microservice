package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

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

func (a *App) closeDb() error {
	if err := a.db.Close(); err != nil {
		return fmt.Errorf("failed to close redis: %w", err)
	}
	return nil
}

func (a *App) startServer(ctx context.Context, server *http.Server) error {
	ch := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		close(ch)
	}()

	select {
	case err := <-ch:
		return err
	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		if shutdownErr := server.Shutdown(timeout); shutdownErr != nil {
			return fmt.Errorf("failed to shutdown server: %w", shutdownErr)
		}

		return nil
	}
}

func (a *App) Start(ctx context.Context) error {
	if err := a.connectToDb(ctx); err != nil {
		return err
	}

	defer a.closeDb()

	fmt.Println("Starting server")

	server := &http.Server{
		Addr:    ":3000",
		Handler: a.router,
	}

	return a.startServer(ctx, server)
}
