package session

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/malikbenkirane/groq-whisper/host/internal/domain/theme"
	"github.com/malikbenkirane/groq-whisper/host/internal/repo"
	"github.com/spf13/cobra"
)

func newCommandCurrent(r repo.Theatre) *cobra.Command {
	return &cobra.Command{
		Use:  "current THEME",
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			s, err := r.CurrentSession(theme.Name(args[0]))
			if err != nil {
				return fmt.Errorf("%w: %w", repo.ErrCurrentSession, err)
			}
			actors := make([]actorJson, len(s.Actors))
			for i, actor := range s.Actors {
				actors[i] = actorJson{
					Name: string(actor.Name),
					Site: string(actor.Site),
				}
			}
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			return encoder.Encode(sessionJson{
				Actors: actors,
				Id:     int(s.ID),
			})
		},
	}
}

type sessionJson struct {
	Chunks []chunkJson
	Actors []actorJson
	Id     int
}

type actorJson struct {
	Name string
	Site string
}

type chunkJson struct{}
