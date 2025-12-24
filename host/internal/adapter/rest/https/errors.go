package https

//go:generate stringer -type=errSys
type errSys int

const (
	errUnknown errSys = iota
	errGetThemes
	errRepoThemes
	errJsonEncode
	errMax
)

func (err errSys) Error() string {
	return err.String()
}
