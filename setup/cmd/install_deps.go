package cmd

import (
	"fmt"

	"github.com/malikbenkirane/groq-whisper/setup/pkg/dep"
	"github.com/spf13/cobra"
)

func newCommandInstallDeps() *cobra.Command {
	var path, bucket, ffmpeg *string

	cmd := &cobra.Command{
		Use:     "install-deps",
		Aliases: []string{"deps", "d"},
		RunE: func(cmd *cobra.Command, args []string) error {
			i, err := dep.NewInstaller(*path, *bucket, *ffmpeg)
			if err != nil {
				return fmt.Errorf("new installer: %w", err)
			}
			return i.Download()
		},
	}

	bucket = cmd.Flags().String("bucket", "groq-whisper", "gcp ffmpeg bucket name")
	ffmpeg = cmd.Flags().String("ffmpeg-7z-object", "ffmpeg-8.0.1-full_build.7z", "gcp ffmpeg 7z object name")
	path = cmd.Flags().String("path", "groq-whisper", "installation path")

	return cmd
}
