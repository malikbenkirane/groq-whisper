package version

import "net/http"

//go:generate stringer -type=ErrVersion
type ErrVersion int

const (
	ErrNoVersion ErrVersion = iota
	ErrNoVersionFile
	ErrOpenVersionFile
	ErrStatVersion
	ErrCopyVersion
	ErrBearerNotFound
	ErrTokenNoEqSecretKey
	ErrPostVersionEmpty
	ErrPostVersionMustStartWithV
	ErrPostVersionTooLong
	ErrAuth
)

func (err ErrVersion) Error() string {
	return err.String()
}

func (err ErrVersion) Status() int {
	switch err {
	case ErrAuth:
		return http.StatusUnauthorized
	}
	return http.StatusInternalServerError
}
