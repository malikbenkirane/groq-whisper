package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const version = "v0.2.2"

func NewCLI() *cobra.Command {
	cmd := &cobra.Command{
		Use: "groq-whisper-deploy",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(
		newCommandVersion(),
	)
	return cmd
}

func mustLogger(lvl zapcore.Level) *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.Level.SetLevel(lvl)
	logger, err := config.Build()
	if err != nil {
		panic(err)
	}
	return logger
}
