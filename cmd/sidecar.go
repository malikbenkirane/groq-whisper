package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func newCommandSidecar(log *zap.Logger) *cobra.Command {
	return &cobra.Command{
		Use:     "sidecar",
		Aliases: []string{"watch", "w"},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

			defer func() {
				if err != nil {
					log.Error("sidecar failed", zap.Error(err))
				}
			}()

			w, err := fsnotify.NewWatcher()
			if err != nil {
				return fmt.Errorf("fsnotify new watcher: %w", err)
			}
			defer func() {
				err = w.Close()
			}()

			ctx, cancel := context.WithCancel(cmd.Context())
			go func(ctx context.Context) {
				for {
					select {
					case event, ok := <-w.Events:
						if !ok {
							return
						}
						if event.Has(fsnotify.Create) &&
							strings.HasSuffix(event.Name, ".mp3") {
							log.Info("notify write", zap.String("file", event.Name))
						}
					case err, ok := <-w.Errors:
						if !ok {
							return
						}
						log.Error("fsnotify", zap.Error(err))
					}
				}
			}(ctx)

			if err = w.Add("."); err != nil {
				cancel()
				return fmt.Errorf("fsnotify add cwd: %w", err)
			}

			<-quit
			cancel()

			return nil
		},
	}
}
