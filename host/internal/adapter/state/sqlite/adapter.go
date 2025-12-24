package sqlite

import (
	"database/sql"
	"fmt"

	"github.com/malikbenkirane/groq-whisper/host/internal/domain/actor"
	"github.com/malikbenkirane/groq-whisper/host/internal/domain/theme"
	"github.com/malikbenkirane/groq-whisper/host/internal/repo"

	_ "github.com/glebarez/go-sqlite"
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

func (a adapter) Actors() (map[string]actor.Call, error) {
	rows, err := a.db.Query(`SELECT name, site FROM actors`)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errSelectActors, err)
	}
	actors := make(map[string]actor.Call)
	for rows.Next() {
		var row struct {
			name, site string
		}
		if err := rows.Scan(&row.name, &row.site); err != nil {
			return nil, fmt.Errorf("%w: %w", errSelectActorsScan, err)
		}
		actors[row.name] = actor.Call(row.site)
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("%w: %w", errSelectActorsIter, err)
	}
	return actors, nil
}

func (a adapter) Themes() (map[string]theme.Description, error) {
	rows, err := a.db.Query(`SELECT name, title, category, keyword FROM themes`)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errSelectThemes, err)
	}
	themes := make(map[string]*theme.Description)           // index themes
	keywords := make(map[string]map[string][]theme.Keyword) // index category keywords
	for rows.Next() {
		var row struct {
			name     string
			title    sql.NullString
			category string
			keyword  string
		}
		if err := rows.Scan(&row.name, &row.title, &row.category, &row.keyword); err != nil {
			return nil, fmt.Errorf("%w: %w", errSelectThemesScan, err)
		}
		if _, ok := themes[row.name]; !ok {
			themes[row.name] = &theme.Description{
				Name: row.name,
			}
			if row.title.Valid {
				themes[row.name].Title = row.title.String
			}
			keywords[row.name] = make(map[string][]theme.Keyword)
		}
		if _, ok := keywords[row.name][row.category]; !ok {
			keywords[row.name][row.category] = make([]theme.Keyword, 0)
		}
		keywords[row.name][row.category] = append(keywords[row.name][row.category],
			theme.Keyword(row.keyword))
	}
	for name, categories := range keywords {
		for category, keywords := range categories {
			cat := theme.Category{
				Name:     category,
				Keywords: keywords,
			}
			themes[name].Categories = append(themes[name].Categories, cat)
		}
	}
	if rows.Err() != nil {
		return nil, fmt.Errorf("%w: %w", errSelectThemesIter, err)
	}
	final := make(map[string]theme.Description, len(themes))
	for _, theme := range themes {
		final[theme.Name] = *theme
	}
	return final, nil
}
