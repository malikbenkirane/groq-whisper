package repo

import (
	"github.com/malikbenkirane/groq-whisper/host/internal/domain/actor"
	"github.com/malikbenkirane/groq-whisper/host/internal/domain/theme"
)

type Theatre interface {
	Themes() (map[string]theme.Description, error)
	Actors() (map[string]actor.Call, error)
}
