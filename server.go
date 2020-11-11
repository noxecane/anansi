package siber

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// WithCancel replicates context.WithCancel but listens for Interrupt and SIGTERM signals
func WithCancel(parent context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(parent)
	go func() {
		defer cancel()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

		<-quit
	}()

	return ctx, cancel
}
