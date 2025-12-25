package https

//go:generate stringer -type=errSys
type errSys int

const (
	errUnknown errSys = iota
	errGetThemes
	errRepoThemes
	errGetActors
	errRepoActors
	errJsonEncode
	errJsonDecode
	errDecodeTxPayload
	errExpectedContentTypeJSON
	errBadRequest
	errStrconvSession
	errMax
)

func (err errSys) Error() string {
	return err.String()
}
