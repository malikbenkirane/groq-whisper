package repo

import (
	"time"

	"github.com/malikbenkirane/groq-whisper/host/internal/domain/actor"
	"github.com/malikbenkirane/groq-whisper/host/internal/domain/session"
	"github.com/malikbenkirane/groq-whisper/host/internal/domain/theme"
	"github.com/malikbenkirane/groq-whisper/host/internal/domain/transcript"
)

type Theatre interface {
	Themes() (map[string]theme.Description, error)
	Actors() (map[string]actor.Call, error)

	LockActor(name actor.Name, id session.Id) error
	UnlockActor(name actor.Name, id session.Id) error
	IsActorLocked(name actor.Name) (*session.Id, error)
	UnlockedActors(name theme.Name) ([]actor.Description, error)
	ResetActorLocks() error

	StartSession(name theme.Name, t time.Time) error
	StopSession(name theme.Name, t time.Time) error
	CurrentSession(name theme.Name) (*session.Session, error)

	SaveTranscriptChunk(chunk transcript.Chunk, id session.Id) error
}
