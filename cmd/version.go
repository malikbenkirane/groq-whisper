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
			fmt.Println(version.Version)
			return nil
		},
	}
}
