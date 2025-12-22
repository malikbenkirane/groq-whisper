package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/malikbenkirane/groq-whisper/setup/pkg/version"
)

func bucketFile(remote, version string) string {
	return fmt.Sprintf("%s/groq-%s.exe", remote, version)
}

func newCommandUpgrade() *cobra.Command {
	var bucket *string
	cmd := &cobra.Command{
		Use: "upgrade",
		RunE: func(cmd *cobra.Command, args []string) error {
			u, err := version.NewUpgrader(*bucket)
			if err != nil {
				return fmt.Errorf("new upgrader: %w", err)
			}
			if version.Version == u.Version() {
				fmt.Println("Already latest version", u.Version())
				return nil
			}
			if err := u.Upgrade(); err != nil {
				return fmt.Errorf("upgrade: %w", err)
			}
			fmt.Println("Upgrade successful from", version.Version, "to", u.Version())
			return nil
		},
	}
	bucket = cmd.Flags().String("bucket", "groq-whisper", "gcp storage bucket")
	return cmd
}
