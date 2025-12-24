package sqlite

//go:generate stringer -type=errAdapter
type errAdapter int

func (err errAdapter) Error() string {
	return err.String()
}

const (
	errZero errAdapter = iota
	errOpenDB
	errSelectThemes
	errSelectThemesIter
	errSelectThemesScan
	errUnknown
)
