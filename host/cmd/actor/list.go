package actor

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/malikbenkirane/groq-whisper/host/internal/repo"
	"github.com/spf13/cobra"
)

func newCommandList(r repo.Theatre) *cobra.Command {
	return &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			actors, err := r.Actors()
			if err != nil {
				return fmt.Errorf("repo actors: %w", err)
			}
			j := make([]actorJson, 0, len(actors))
			for name, site := range actors {
				j = append(j, actorJson{name, string(site)})
			}
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			if err := encoder.Encode(j); err != nil {
				return fmt.Errorf("encode actor json: %w", err)
			}
			return nil
		},
	}
}

type actorJson struct {
	Name string
	Call string
}
