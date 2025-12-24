package repo

import "github.com/malikbenkirane/groq-whisper/host/internal/domain/theme"

type Theatre interface {
	Themes() ([]theme.Description, error)
}
