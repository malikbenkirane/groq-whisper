package main

import (
	"log/slog"
	"os"

	"github.com/malikbenkirane/groq-whisper/host/cmd"
	"github.com/spf13/cobra"
)

func main() {
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(h)
	slog.SetDefault(logger)
	cmd, err := cmd.NewCLI()
	if err != nil {
		slog.Error("init cli", "err", err)
		os.Exit(1)
	}
	cobra.CheckErr(cmd.Execute())
}
