package sqlite

import (
	"database/sql"
	"fmt"

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

func (a adapter) Themes() (map[string]theme.Description, error) {
	rows, err := a.db.Query(`SELECT (name, title, category, keyword) FROM themes`)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errSelectThemes, err)
	}
	themes := make(map[string]*theme.Description)    // index themes
	categories := make(map[string][]*theme.Category) // index theme categories
	keywords := make(map[string][]theme.Keyword)     // index category keywords
	for rows.Next() {
		var row struct {
			name     string
			title    string
			category string
			keyword  string
		}
		if err := rows.Scan(&row.name, &row.title, &row.category, &row.keyword); err != nil {
			return nil, fmt.Errorf("%w: %w", errSelectThemesScan, err)
		}
		if _, ok := themes[row.name]; !ok {
			categories[row.name] = []*theme.Category{}
			themes[row.name] = &theme.Description{
				Name:  row.name,
				Title: row.title,
			}
		}
		if _, ok := categories[row.category]; !ok {
			keywords[row.category] = []theme.Keyword{}
		}
		categories[row.name] = append(categories[row.name], &theme.Category{Name: row.category})
		keywords[row.category] = append(keywords[row.category], theme.Keyword(row.keyword))
	}
	for name, categories := range categories {
		final := make([]theme.Category, 0, len(categories))
		for _, category := range categories {
			category.Keywords = keywords[category.Name]
			final = append(final, *category)
		}
		themes[name].Categories = final
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
