package cmd

import (
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewCLI(version string) *cobra.Command {
	cmd := &cobra.Command{
		Use: "groq",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(
		newCommandRecord(),
		newCommandSidecar(),
		newCommandVersion(version))
	return cmd
}

func newLogger(debug bool) *zap.Logger {
	config := zap.NewProductionConfig()
	if debug {
		config = zap.NewDevelopmentConfig()
	}
	log, err := config.Build()
	if err != nil {
		panic(err)
	}
	return log
}
