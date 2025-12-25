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
	"github.com/malikbenkirane/groq-whisper/internal/server"
	"github.com/spf13/cobra"
)

func newCommandServe() *cobra.Command {
	var serfMaster, httpAddr, root *string
	var serfPort *int
	var sys32, loop *bool
	cmd := &cobra.Command{
		Use: "serve",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := os.MkdirAll(*root, 0700); err != nil {
				return fmt.Errorf("mkdir all %q: %w", *root, err)
			}

			opts := []server.Option{}
			if !*sys32 {
				sampler := sampler.New(
					16000,
					time.Second*10,
					sampler.EncoderOptionRoot(*root))
				opts = append(opts, server.OptionSampler(sampler))
			}

			if !*loop {
				opts = append(opts,
					server.OptionNoHttp(),
					server.OptionNoLoop())
			}

			opts = append(opts,
				server.OptionSerfMaster(*serfMaster),
				server.OptionSerfPort(*serfPort),
				server.OptionHttpAddr(*httpAddr))

			sv, err := server.New(*root, opts...)
			if err != nil {
				return fmt.Errorf("new server: %w", err)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGINT)

			go sv.Serve(ctx)
			<-quit
			return nil
		},
	}

	defaultRoot, err := os.UserHomeDir()
	if err != nil {
		panic("user home dir")
	}
	defaultRoot = path.Join(defaultRoot, "Downloads", "groq-whisper-samples")

	root = cmd.Flags().String("root-fs", defaultRoot, "server root")

	serfMaster = cmd.Flags().String("serf-master", "192.168.117.1:7496", "")
	serfPort = cmd.Flags().Int("serf-port", 7496, "serf binding port")

	httpAddr = cmd.Flags().String("http-bind", ":7495", "addr bind for the http server")

	sys32 = cmd.Flags().Bool("sys32", true, "for windows x64 install (see docs/install.md)")

	loop = cmd.Flags().Bool("loop", true, "disable to only serf gossip (and combine with \"\" master)")

	return cmd
}
