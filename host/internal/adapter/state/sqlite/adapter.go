package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/malikbenkirane/groq-whisper/host/internal/domain/actor"
	"github.com/malikbenkirane/groq-whisper/host/internal/domain/session"
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
				Name: theme.Name(row.name),
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
		final[string(theme.Name)] = *theme
	}
	return final, nil
}

func (a adapter) setLock(isLocked bool, name actor.Name, id session.Id) error {
	if _, err := a.db.Exec(`
INSERT INTO actors_locks (islocked, actor, session) VALUES(?, ?, ?)
ON CONFLICT(actor) DO UPDATE SET islocked = ?
	`, isLocked, string(name), int(id), isLocked); err != nil {
		return fmt.Errorf("%w: %w", errExecSetLock, err)
	}
	return nil
}

func (a adapter) LockActor(name actor.Name, id session.Id) error {
	return a.setLock(true, name, id)
}

func (a adapter) UnlockActor(name actor.Name, id session.Id) error {
	return a.setLock(false, name, id)
}

func (a adapter) IsActorLocked(name actor.Name) (*session.Id, error) {
	row := a.db.QueryRow(`
SELECT session FROM actors_locks WHERE actor = ?
	`, string(name))
	var id int
	err := row.Scan(&id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: %w", errQueryRowActorsLocks, err)
	}
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	sid := new(session.Id)
	*sid = session.Id(id)
	return sid, nil
}

func (a adapter) ResetActorLocks() error {
	res, err := a.db.Exec(`DELETE FROM actors_locks`)
	if err != nil {
		return fmt.Errorf("%w: %w", errDeleteActorsLocks, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %w: %w", errDeleteActorsLocks, errReadRowsAffected, err)
	}
	slog.Warn("actor locks reset", "n", n)
	return nil
}

func (a adapter) StartSession(name theme.Name, t time.Time) error {
	if _, err := a.db.Exec(`
INSERT INTO sessions (theme, start8601) VALUES (?, ?)
	`, string(name), t.Format(iso8601)); err != nil {
		return fmt.Errorf("%w: %w", errInsertSession, err)
	}
	return nil
}

func (a adapter) StopSession(name theme.Name, t time.Time) error {
	if _, err := a.db.Exec(`
UPDATE sessions SET stop8601 = ? WHERE theme = ? AND stop8601 IS NULL
	`, t.Format(iso8601), string(name)); err != nil {
		return fmt.Errorf("%w: %w", errUpdateSession, err)
	}
	return nil
}

func (a adapter) CurrentSession(name theme.Name) (*session.Session, error) {
	row := a.db.QueryRow(`
SELECT id FROM sessions WHERE theme = ? AND stop8601 is NULL 
	`, string(name))
	var id int
	err := row.Scan(&id)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: %w", errQueryRowCurrentSession, err)
	}
	if err != nil {
		return nil, nil
	}
	actors, err := a.actors(session.Id(id))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errActorsSession, err)
	}
	themes, err := a.Themes()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errThemes, err)
	}
	return &session.Session{
		Actors: actors,
		Theme:  themes[string(name)],
		ID:     session.Id(id),
	}, nil
}

const iso8601 = "2006-01-02T15:04:05.000"

func (a adapter) actors(id session.Id) ([]actor.Description, error) {
	rows, err := a.db.Query(`
SELECT actor FROM actors_locks WHERE id = ?
	`, int(id))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errSelectActorsLocks, err)
	}
	names := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("%w: %w: %w", errSelectActorsLocks, errScan, err)
		}
		names = append(names, name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", errSelectActorsLocks, err)
	}
	sites := make([]string, len(names))
	for i, name := range names {
		var site string
		if err := a.db.QueryRow(`
SELECT site FROM actors WHERE name = ?
		`, name).Scan(&site); err != nil {
			return nil, fmt.Errorf("%w: %w", errQueryRowActors, err)
		}
		sites[i] = site
	}
	actors := make([]actor.Description, len(names))
	for i := range names {
		actors[i] = actor.Description{
			Name: actor.Name(names[i]),
			Site: actor.Call(sites[i]),
		}
	}
	return actors, nil
}

func (a adapter) UnlockedActors(name theme.Name) ([]actor.Description, error) {
	actors, err := a.Actors()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errActors, err)
	}
	rows, err := a.db.Query(`SELECT actor FROM actors_locks WHERE islocked = 1`)
	if err != nil {
		return nil, fmt.Errorf("%w, %w", errSelectActorsLocks, err)
	}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("%w: %w: %", errSelectActorsLocks, errScan, err)
		}
		delete(actors, name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%w: %w", errSelectActorsLocks, err)
	}
	names := make([]actor.Description, len(actors))
	i := -1
	for name, site := range actors {
		i++
		names[i] = actor.Description{Name: actor.Name(name), Site: site}
	}
	return names, nil
}
