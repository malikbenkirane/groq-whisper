package cmd

import (
	"fmt"
	"os"

	"github.com/malikbenkirane/groq-whisper/setup/pkg/version"
	"github.com/spf13/cobra"
)

func newCommandInstall() *cobra.Command {
	var bucket, path, appName, installerName, versionSource *string
	cmd := &cobra.Command{
		Use:     "install",
		Aliases: []string{"i"},
		RunE: func(cmd *cobra.Command, args []string) error {
			err := os.MkdirAll(*path, 0700)
			if err != nil {
				return fmt.Errorf("mkdir %q: %w", *path, err)
			}
			if err = os.Chdir(*path); err != nil {
				return fmt.Errorf("chdir %q: %w", *path, err)
			}
			u, err := version.NewUpgrader(*bucket,
				version.UpgraderWithApp(*appName),
				version.UpgraderWithInstaller(*installerName),
				version.UpgraderWithVersionSource(*versionSource))
			if err != nil {
				return fmt.Errorf("new upgrader: %w", err)
			}
			if u.Version() == version.Version {
				fmt.Println("Already up to date")
				return nil
			}
			return u.Upgrade()
		},
	}

	cfg := version.DefaultUpgraderConfig()

	bucket = cmd.Flags().String("bucket", "groq-whisper", "gcloud bucket name")
	path = cmd.Flags().String("path", "groq", "installation path")
	appName = cmd.Flags().String("app-name", cfg.App, "groq whisper app name")
	installerName = cmd.Flags().String("intaller-name", cfg.Installer, "groq whisper installer app name")
	versionSource = cmd.Flags().String("version-source", cfg.VersionSrc, "version source url")

	return cmd
}
