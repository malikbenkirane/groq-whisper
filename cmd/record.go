package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/malikbenkirane/groq-whisper/internal/sampler"

	"github.com/spf13/cobra"
)

func newCommandRecord() (*cobra.Command, error) {
	var freq *int
	var debug, sys32 *bool
	var samplesDir *string

	cmd := &cobra.Command{
		Use:     "record",
		Aliases: []string{"rec", "r"},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			log := newLogger(*debug)
			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

			encoderOpts := []sampler.EncoderOption{}
			if *sys32 {
				encoderOpts = append(encoderOpts,
					sampler.EncoderOptionPath(
						"C:\\Windows\\System32\\groq\\groq-deps\\bin\\ffmpeg"))
			}

			if err := os.MkdirAll(*samplesDir, 0700); err != nil {
				return fmt.Errorf("mkdir %q: %w", *samplesDir, err)
			}
			encoderOpts = append(encoderOpts, sampler.EncoderOptionRoot(*samplesDir))

			ctx, cancel := context.WithCancel(cmd.Context())
			s := sampler.New(log, float64(*freq), time.Duration(time.Second*10), encoderOpts...)
			go s.Sample(ctx)

			<-quit
			cancel()

			return nil
		},
	}

	defaultDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("user home dir: %w", err)
	}
	defaultDir = path.Join(defaultDir, "groq-whisper-samples")

	freq = cmd.Flags().IntP("freq", "f", 16000, "sample rate")
	debug = cmd.Flags().Bool("debug", false, "set log level at debug")
	sys32 = cmd.Flags().Bool("ffmpeg-sys32", true, "use ffmpeg from windows/sys32/groq-deps")
	samplesDir = cmd.Flags().String("samples-dir", defaultDir, "where recorded samples are processed")

	return cmd, nil
}
