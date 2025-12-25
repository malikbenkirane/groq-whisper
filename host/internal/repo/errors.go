package repo

//go:generate stringer -type=Error
type Error int

const (
	ErrThemes Error = iota
	ErrActors
	ErrLockActor
	ErrUnlockActor
	ErrIsActorLocked
	ErrResetActorLocks
	ErrStartSession
	ErrStopSession
	ErrCurrentSession
)
