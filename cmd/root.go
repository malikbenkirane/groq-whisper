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
	log, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}
	cmd.AddCommand(newCommandRecord(log), newCommandSidecar(log))
	return cmd
}
