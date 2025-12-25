package session

import (
	"time"

	"github.com/malikbenkirane/groq-whisper/host/internal/domain/actor"
	"github.com/malikbenkirane/groq-whisper/host/internal/domain/theme"
	"github.com/malikbenkirane/groq-whisper/host/internal/domain/transcript"
)

type Session struct {
	startAt *time.Time
	endAt   *time.Time

	Actors []actor.Description
	Theme  theme.Description
	Chunks []transcript.Chunk

	ID Id
}

func (s *Session) Start(t time.Time) { s.startAt = &t }
func (s *Session) End(t time.Time)   { s.endAt = &t }

type Id int
