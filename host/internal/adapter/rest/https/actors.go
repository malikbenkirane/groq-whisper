package https

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (a adapter) handleGetActors() customHandler {
	return func(w http.ResponseWriter, r *http.Request) (errUser error, errSys error) {
		w.Header().Add("Content-Type", "application/json")
		actors, err := a.repo.Actors()
		if err != nil {
			return errInternalError, fmt.Errorf("%w: %w: %w", errGetActors, errRepoActors, err)
		}
		toEncode := make([]actor, len(actors))
		var i = -1
		for name, site := range actors {
			i++
			toEncode[i] = actor{Name: name, Site: string(site)}
		}
		if err := json.NewEncoder(w).Encode(toEncode); err != nil {
			return errInternalError, fmt.Errorf("%w: %w", errJsonEncode, err)
		}
		return
	}
}

type actor struct {
	Name string `json:"name"`
	Site string `json:"site"`
}
