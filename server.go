package siber

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
)

// CancelOnInterrupt cancels a context when it receives an interrupt signal
func CancelOnInterrupt(cancel context.CancelFunc, log zerolog.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// wait for Quit signal
	<-quit

	// cancel context so dependencies can gracefully shutdown
	log.Info().Msg("Signal caught. Shutting down...")
	cancel()
}

func RunServer(ctx context.Context, log zerolog.Logger, server *http.Server) {
	// shutdown the server when context is done
	go func() {
		<-ctx.Done()

		// shutdown server in 5s
		shutCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		if err := server.Shutdown(shutCtx); err != nil {
			log.Err(err).Msg("Could not shut down server cleanly.....")
		}
	}()

	log.Info().Msgf("Serving api at http://127.0.0.1%s", server.Addr)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Err(err).Msg("Could not start the server")
	}
}
