package theme

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/malikbenkirane/groq-whisper/host/internal/domain/theme"
	"github.com/malikbenkirane/groq-whisper/host/internal/repo"
	"github.com/spf13/cobra"
)

func newCommandList(r repo.Theatre) *cobra.Command {
	return &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			themes, err := r.Themes()
			if err != nil {
				return fmt.Errorf("repo themes: %w", err)
			}
			encoder := json.NewEncoder(os.Stdout)
			encoder.SetIndent("", "  ")
			for _, theme := range themes {
				j := fromTheme(theme)
				if err := encoder.Encode(j); err != nil {
					return fmt.Errorf("json encode: %w", err)
				}
			}
			return nil
		},
	}
}

type themeJson struct {
	Name       string
	Title      string
	Categories []categoryJson
}

type categoryJson struct {
	Name     string
	Keywords []string
}

func fromTheme(t theme.Description) themeJson {
	j := themeJson{
		Name:       t.Name,
		Title:      t.Title,
		Categories: []categoryJson{},
	}
	categories := make([]categoryJson, len(t.Categories))
	for i, c := range t.Categories {
		keywords := make([]string, 0, len(c.Keywords))
		for i, k := range c.Keywords {
			keywords[i] = string(k)
		}
		categories[i] = categoryJson{
			Name:     c.Name,
			Keywords: keywords,
		}
	}
	j.Categories = categories
	return j
}
