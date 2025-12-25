package https

import (
	"fmt"
	"net/http"
	"time"

	"github.com/malikbenkirane/groq-whisper/host/internal/domain/theme"
	"github.com/malikbenkirane/groq-whisper/host/internal/repo"
)

func (a adapter) handlePostSession() customHandler {
	return func(w http.ResponseWriter, r *http.Request) (errUser error, errSys error) {
		if err := a.repo.StartSession(theme.Name(r.PathValue("theme")), time.Now()); err != nil {
			return errInternalError, fmt.Errorf("%w: %w", repo.ErrStartSession, err)
		}
		return
	}
}

func (a adapter) handleDeleteSession() customHandler {
	return func(w http.ResponseWriter, r *http.Request) (errUser error, errSys error) {
		if err := a.repo.StopSession(theme.Name(r.PathValue("theme")), time.Now()); err != nil {
			return errInternalError, fmt.Errorf("%w: %w", repo.ErrStopSession, err)
		}
		return
	}
}
