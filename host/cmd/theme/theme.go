package theme

import (
	"fmt"

	"github.com/malikbenkirane/groq-whisper/host/internal/adapter/state/sqlite"
	"github.com/spf13/cobra"
)

func NewCommand() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use: "theme",
	}

	a, err := sqlite.New()
	if err != nil {
		return nil, fmt.Errorf("sqlite new: %w", err)
	}

	cmd.AddCommand(
		newCommandList(a))

	return cmd, nil
}
