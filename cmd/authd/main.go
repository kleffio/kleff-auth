package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kleffio/kleff-auth/internal/infrastructure/bootstrap"
)

func main() {
	ctx := context.Background()

	app, err := bootstrap.NewApp(ctx)
	if err != nil {
		log.Fatalf("bootstrap: %v", err)
	}

	// start HTTP server
	go func() {
		if err := app.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("listen: %v", err)
		}
	}()

	// graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := app.Server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	} else {
		log.Printf("server stopped")
	}

	// close DB, etc.
	if err := app.Shutdown(context.Background()); err != nil {
		log.Printf("cleanup failed: %v", err)
	}
}
