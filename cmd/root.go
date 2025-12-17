package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewCLI() *cobra.Command {
	cmd := &cobra.Command{
		Use: "groq",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(
		newCommandRecord(),
		newCommandSidecar())
	return cmd
}

func newLogger(debug bool) *zap.Logger {
	lvl := zap.InfoLevel
	if debug {
		lvl = zap.DebugLevel
	}
	config := zap.Config{
		Level: zap.NewAtomicLevelAt(lvl),
	}
	log, err := config.Build()
	if err != nil {
		panic(err)
	}
	return log
}
