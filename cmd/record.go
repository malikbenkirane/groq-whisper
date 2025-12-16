package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"groq/internal/sampler"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func newCommandRecord(log *zap.Logger) *cobra.Command {
	var freq *int
	cmd := &cobra.Command{
		Use:     "record",
		Aliases: []string{"rec", "r"},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

			ctx, cancel := context.WithCancel(cmd.Context())
			s := sampler.New(log, float64(*freq), time.Duration(time.Second*10))
			go s.Sample(ctx)

			<-quit
			cancel()

			return nil
		},
	}
	freq = cmd.Flags().IntP("freq", "f", 16000, "sample rate")
	return cmd
}
