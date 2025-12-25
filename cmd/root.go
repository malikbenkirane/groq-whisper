package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func NewCLI() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use: "groq",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	record, err := newCommandRecord()
	if err != nil {
		return nil, fmt.Errorf("new command record: %w", err)
	}

	sidecar, err := newCommandSidecar()
	if err != nil {
		return nil, fmt.Errorf("new command sidecar: %w", err)
	}

	cmd.AddCommand(
		sidecar,
		newCommandVersion(),
		newCommandUpgrade(),
		newCommandDev(),
		newCommandServe(),
		record)

	return cmd, nil
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
