package https

import (
	"log/slog"
	"net/http"

	"github.com/malikbenkirane/groq-whisper/host/internal/repo"
)

func New(r repo.Theatre, opts ...Option) (Adapter, error) {
	conf := DefaultConfig()
	for _, opt := range opts {
		conf = opt(conf)
	}
	mux := http.NewServeMux()
	a := &adapter{
		config: conf,
		mux:    mux,
		repo:   r,
	}
	mux.Handle("GET /themes", wrap(a.handleGetThemes()))
	mux.Handle("GET /actors", wrap(a.handleGetActors()))
	mux.Handle("GET /actors/{theme}", wrap(a.handleGetActorsTheme()))
	mux.Handle("POST /session/{theme}", wrap(a.handlePostSession()))
	mux.Handle("DELETE /session/{theme}", wrap(a.handleDeleteSession()))
	mux.Handle("POST /lock/actor/{theme}/{actor}", wrap(a.handlePostLockActor()))
	mux.Handle("DELETE /lock/actor/{theme}/{actor}", wrap(a.handleDeleteLockActor()))
	return a, nil
}

type Adapter interface {
	Serve() error
}

type Config struct {
	addr, certFile, keyFile string
}

type Option func(Config) Config

func DefaultConfig() Config {
	return Config{
		addr:     ":50001",
		certFile: "cert.pem",
		keyFile:  "cert.key",
	}
}

func OptionAddr(addr string) Option {
	return func(c Config) Config {
		c.addr = addr
		return c
	}
}

func OptionCertFile(path string) Option {
	return func(c Config) Config {
		c.certFile = path
		return c
	}
}

func OptionKeyFile(path string) Option {
	return func(c Config) Config {
		c.keyFile = path
		return c
	}
}

type adapter struct {
	config Config
	mux    *http.ServeMux

	repo repo.Theatre
}

func (a adapter) Serve() error {
	slog.Info("starting https server", "addr", a.config.addr, "cert", a.config.certFile)
	return http.ListenAndServeTLS(a.config.addr, a.config.certFile, a.config.keyFile, a.mux)
}
