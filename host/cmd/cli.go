package cmd

import (
	"fmt"

	"github.com/malikbenkirane/groq-whisper/host/cmd/actor"
	"github.com/malikbenkirane/groq-whisper/host/cmd/session"
	"github.com/malikbenkirane/groq-whisper/host/cmd/theme"
	"github.com/malikbenkirane/groq-whisper/host/internal/adapter/state/sqlite"
	"github.com/spf13/cobra"
)

func NewCLI() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use: "groq-host",
	}

	a, err := sqlite.New()
	if err != nil {
		return nil, fmt.Errorf("sqlite new: %w", err)
	}
	cmd.AddCommand(
		newCommandMkcert(),
		newCommandServe(a),
		theme.NewCommand(a),
		actor.NewCommand(a),
		session.NewCommand(a))
	return cmd, nil
}
