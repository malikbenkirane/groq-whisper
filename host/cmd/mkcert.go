package cmd

import (
	"github.com/malikbenkirane/groq-whisper/host/pkg/sec"
	"github.com/spf13/cobra"
)

func newCommandMkcert() *cobra.Command {
	return &cobra.Command{
		Use:  "mkcert HOSTS",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			certifier := sec.New(args[0])
			return certifier.NewCertificate()
		},
	}
}
