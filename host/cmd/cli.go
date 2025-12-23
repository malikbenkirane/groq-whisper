package cmd

import "github.com/spf13/cobra"

func NewCLI() *cobra.Command {
	cmd := &cobra.Command{
		Use: "groq-host",
	}
	cmd.AddCommand(
		newCommandMkcert())
	return cmd
}
