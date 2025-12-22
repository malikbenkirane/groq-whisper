package cmd

import (
	"fmt"

	"github.com/malikbenkirane/groq-whisper/setup/pkg/version"
	"github.com/spf13/cobra"
)

func newCommandVersion() *cobra.Command {
	return &cobra.Command{
		Use: "version",
		RunE: func(cmd *cobra.Command, args []string) error {
			u, err := version.NewUpgrader("")
			if err != nil {
				return fmt.Errorf("new upgrader: %w", err)
			}
			fmt.Println("local:", version.Version, "remote:", u.Version())
			return nil
		},
	}
}
