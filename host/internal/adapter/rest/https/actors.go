package https

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/malikbenkirane/groq-whisper/host/internal/domain/theme"
	"github.com/malikbenkirane/groq-whisper/host/internal/repo"
)

func (a adapter) handleGetActors() customHandler {
	return func(w http.ResponseWriter, r *http.Request) (errUser error, errSys error) {
		w.Header().Add("Content-Type", "application/json")
		actors, err := a.repo.Actors()
		if err != nil {
			return errInternalError, fmt.Errorf("%w: %w: %w", errGetActors, errRepoActors, err)
		}
		toEncode := make([]actorJson, len(actors))
		var i = -1
		for name, site := range actors {
			i++
			toEncode[i] = actorJson{Name: name, Site: string(site)}
		}
		if err := json.NewEncoder(w).Encode(toEncode); err != nil {
			return errInternalError, fmt.Errorf("%w: %w", errJsonEncode, err)
		}
		return
	}
}

func (a adapter) handleGetActorsTheme() customHandler {
	return func(w http.ResponseWriter, r *http.Request) (errUser error, errSys error) {
		w.Header().Add("Content-Type", "application/json")
		name := r.PathValue("theme")
		actors, err := a.repo.UnlockedActors(theme.Name(name))
		if err != nil {
			return errInternalError, fmt.Errorf("%w: %w", repo.ErrGetUnlockedActors, err)
		}
		actorsJson := make([]actorJson, len(actors))
		for i, actor := range actors {
			actorsJson[i] = actorJson{
				Name: string(actor.Name),
				Site: string(actor.Site),
			}
		}
		if err := json.NewEncoder(w).Encode(actorsJson); err != nil {
			return errInternalError, fmt.Errorf("%w: %w", errJsonEncode, err)
		}
		return
	}
}

type actorJson struct {
	Name string `json:"name"`
	Site string `json:"site"`
}
