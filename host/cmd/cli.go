package cmd

import (
	"fmt"

	"github.com/malikbenkirane/groq-whisper/host/cmd/theme"
	"github.com/spf13/cobra"
)

func NewCLI() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use: "groq-host",
	}

	cmdTheme, err := theme.NewCommand()
	if err != nil {
		return nil, fmt.Errorf("init theme command: %w", err)
	}

	cmd.AddCommand(
		newCommandMkcert(),
		newCommandServe(),
		cmdTheme)
	return cmd, nil
}
