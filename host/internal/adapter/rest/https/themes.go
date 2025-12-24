package https

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (a adapter) handleGetThemes() customHandler {
	return func(w http.ResponseWriter, r *http.Request) (errUser error, errSys error) {
		w.Header().Add("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode([]theme{
			{"hello"}, {"this"}, {"is"},
		}); err != nil {
			return errInternalError, fmt.Errorf("json encode themes: %w", err)
		}
		return
	}
}

type theme struct {
	Name string `json:"name"`
}
