package https

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (a adapter) handleGetThemes() customHandler {
	return func(w http.ResponseWriter, r *http.Request) (errUser error, errSys error) {
		w.Header().Add("Content-Type", "application/json")
		themes, err := a.repo.Themes()
		if err != nil {
			return errInternalError, fmt.Errorf("%w: %w: %w", errGetThemes, errRepoThemes, err)
		}
		toEncode := make([]theme, len(themes))
		var i = -1
		for name := range themes {
			i++
			toEncode[i] = theme{Name: name}
		}
		if err := json.NewEncoder(w).Encode(toEncode); err != nil {
			return errInternalError, fmt.Errorf("%w: %w", errJsonEncode, err)
		}
		return
	}
}

type theme struct {
	Name string `json:"name"`
}
