package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCommandVersion() *cobra.Command {
	return &cobra.Command{
		Use: "version",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println(version)
			return nil
		},
	}
}
