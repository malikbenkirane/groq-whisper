package version

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"go.uber.org/zap"
)

var ErrNoKey = errors.New("no key")

func NewDeploymentServer(log *zap.Logger, opts ...ServerOption) (DeploymentServer, error) {
	conf := ServerConfig{}
	for _, opt := range opts {
		conf = opt(conf)
	}
	if len(conf.secretKey) == 0 {
		return nil, ErrNoKey
	}
	mux := http.NewServeMux()
	ds := deploymentServer{
		mux:  mux,
		log:  log,
		conf: conf,
	}
	mux.Handle("GET /version", handler(ds.getVersion))
	return ds, nil
}

type DeploymentServer interface {
	Serve(addr string) error
}

type deploymentServer struct {
	mux  *http.ServeMux
	log  *zap.Logger
	conf ServerConfig
}

func (ds deploymentServer) Serve(addr string) error {
	return http.ListenAndServe(addr, ds.mux)
}

func (ds deploymentServer) postVersion(w http.ResponseWriter, r *http.Request) (errUser, err error) {
	version := strings.TrimSpace(r.FormValue("version"))
	if len(version) == 0 {
		return ErrPostVersionEmpty, ErrPostVersionEmpty
	}
	if version[0] != 'v' {
		return ErrPostVersionMustStartWithV,
			fmt.Errorf("%w: got %q", ErrPostVersionMustStartWithV, version)
	}
	if len(version) > 10 {
		return ErrPostVersionTooLong,
			fmt.Errorf("%w: got %d characters", ErrPostVersionTooLong)
	}
	os.Create()
}

func (ds deploymentServer) getVersion(w http.ResponseWriter, r *http.Request) (errUser, err error) {
	defer func() {
		if err != nil {
			ds.log.Error("getVersion", zap.Error(err))
		}
	}()

	_, err = os.Stat(ds.versionFile)
	if errors.Is(err, os.ErrNotExist) {
		return ErrNoVersion, fmt.Errorf("%w: %w", ErrNoVersionFile, err)
	}
	if err != nil {
		return ErrNoVersion, fmt.Errorf("%w: %w", ErrStatVersion, err)
	}
	f, err := os.Open("version.txt")
	if err != nil {
		return ErrNoVersion, fmt.Errorf("%w: %w", ErrNoVersionFile, err)
	}
	var b bytes.Buffer
	if _, err = io.Copy(&b, f); err != nil {
		return ErrNoVersion, fmt.Errorf("%w: %w", ErrCopyVersion, err)
	}
	version := bytes.TrimSpace(b.Bytes())
	w.Write(version)
	return nil, nil
}

func handler(f func(http.ResponseWriter, *http.Request) (error, error)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		errUser, _ := f(w, r)
		if errUser != nil {
			err, ok := errUser.(ErrVersion)
			if !ok {
				http.Error(w, errUser.Error(), http.StatusInternalServerError)
			}
			http.Error(w, err.Error(), err.Status())
		}
	}
}

func (ds deploymentServer) authMiddleware(next http.Handler) http.Handler {
	return handler(func(w http.ResponseWriter, r *http.Request) (errUser, err error) {
		bearer := r.Header.Get("Authorization")
		token, found := strings.CutPrefix(bearer, "Bearer ")
		if !found {
			return ErrAuth, ErrBearerNotFound
		}
		if ds.conf.secretKey != token {
			return ErrAuth, ErrTokenNoEqSecretKey
		}
		next.ServeHTTP(w, r)
		return
	})
}
