package https

import (
	"errors"
	"log/slog"
	"net/http"
)

type customHandler func(w http.ResponseWriter, r *http.Request) (errUser, errSys error)

func wrap(handler customHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		errUser, errSys := handler(w, r)
		if errUser != nil {
			slog.Warn("HTTP user error", "path", r.URL.Path, "err", errUser)
		}
		if errSys != nil {
			slog.Error("HTTP sys error", "path", r.URL.Path, "err", errSys)
		}
		if errUser != nil || errSys != nil {
			http.Error(w, errUser.Error(), http.StatusInternalServerError)
		}
	}
}

var errInternalError = errors.New("internal server error")
