package sqlite

import (
	"database/sql"
	"fmt"

	_ "github.com/glebarez/go-sqlite"
	"github.com/malikbenkirane/groq-whisper/host/internal/repo"
)

func New(opts ...Option) (repo.Theatre, error) {
	conf := defaultConfig()
	for _, opt := range opts {
		conf = opt(conf)
	}
	db, err := sql.Open("sqlite", conf.path)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w: %w", conf.path, errOpenDB, err)
	}
	a := adapter{
		conf: conf,
		db:   db,
	}
	return a, nil
}

type Config struct {
	path string
}

type Option func(Config) Config

func defaultConfig() Config {
	return Config{
		path: "state.db",
	}
}

type adapter struct {
	db   *sql.DB
	conf Config
}
