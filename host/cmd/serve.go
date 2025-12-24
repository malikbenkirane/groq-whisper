package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/malikbenkirane/groq-whisper/host/internal/adapter/live/wss"
	"github.com/malikbenkirane/groq-whisper/host/internal/adapter/rest/https"
	"github.com/malikbenkirane/groq-whisper/host/internal/repo"
	"github.com/spf13/cobra"
)

func newCommandServe(r repo.Theatre) *cobra.Command {
	return &cobra.Command{
		Use: "serve",
		RunE: func(cmd *cobra.Command, args []string) error {
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
			ctx, cancel := context.WithCancel(cmd.Context())
			defer cancel()

			errChan := make(chan error)
			defer close(errChan)
			go func() {
				for err := range errChan {
					slog.Error("server failed", "err", err)
				}
			}()

			s, err := https.New(r)
			if err != nil {
				return fmt.Errorf("new https server: %w", err)
			}
			go func() { errChan <- s.Serve() }()

			w, err := wss.New(ctx)
			if err != nil {
				return fmt.Errorf("new wss server: %w", err)
			}

			go func() { errChan <- w.Serve() }()

			<-quit
			return nil
		},
	}
}
