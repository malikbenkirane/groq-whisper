package repo

//go:generate stringer -type=Error
type Error int

const (
	ErrThemes Error = iota
	ErrActors
	ErrLockActor
	ErrUnlockActor
	ErrGetUnlockedActors
	ErrIsActorLocked
	ErrResetActorLocks
	ErrStartSession
	ErrStopSession
	ErrCurrentSession
	ErrSaveTranscriptChunk
)

func (err Error) Error() string {
	return err.String()
}
