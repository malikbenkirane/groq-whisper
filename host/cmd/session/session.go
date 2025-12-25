package session

import (
	"github.com/malikbenkirane/groq-whisper/host/internal/repo"
	"github.com/spf13/cobra"
)

func NewCommand(r repo.Theatre) *cobra.Command {
	cmd := &cobra.Command{Use: "session"}
	cmd.AddCommand(newCommandCurrent(r))
	return cmd
}
