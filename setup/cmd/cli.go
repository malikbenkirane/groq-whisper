package cmd

import "github.com/spf13/cobra"

func NewCLI() *cobra.Command {
	cmd := &cobra.Command{
		Use: "groq-setup",
	}
	cmd.AddCommand(
		newCommandInstall(),
		newCommandVersion(),
		newCommandInstallDeps())
	return cmd
}
