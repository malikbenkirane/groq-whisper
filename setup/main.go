package main

import (
	"log/slog"
	"os"

	"github.com/malikbenkirane/groq-whisper/setup/cmd"
	"github.com/spf13/cobra"
)

func main() {
	h := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})
	logger := slog.New(h)
	slog.SetDefault(logger)
	cobra.CheckErr(cmd.NewCLI().Execute())
}
