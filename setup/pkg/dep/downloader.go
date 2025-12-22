package dep

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/bodgit/sevenzip"
	"github.com/schollz/progressbar/v3"

	"github.com/malikbenkirane/groq-whisper/setup/pkg/dcheck"
)

func (i installer) installFfmpeg() (err error) {
	if i.conf.overwrite {
		if err := os.RemoveAll(i.path); err != nil {
			return fmt.Errorf("remove all %q: %w", i.path, err)
		}
	}
	var archive bytes.Buffer
	err = i.bucket.Pull(&archive, i.ffmpeg7z)
	if err != nil {
		return fmt.Errorf("bucket pull %q: %w", i.ffmpeg7z, err)
	}
	reader := bytes.NewReader(archive.Bytes())
	{
		reader, err := sevenzip.NewReader(reader, reader.Size())
		if err != nil {
			return fmt.Errorf("7z new reader: %w", err)
		}
		bar := progressbar.NewOptions(len(reader.File),
			progressbar.OptionShowDescriptionAtLineEnd(),
			progressbar.OptionSetDescription("install ffmpeg"),
			progressbar.OptionSetMaxDetailRow(8))
		for _, f := range reader.File {
			dst, err := i.extract(f)
			if err != nil {
				return fmt.Errorf("extract: %w", err)
			}
			if err := bar.AddDetail(dst); err != nil {
				return fmt.Errorf("bar add detail: %w", err)
			}
			if err := bar.Add(1); err != nil {
				return fmt.Errorf("bar add: %w", err)
			}
		}
	}
	return nil
}

func (i installer) installPortaudio() (err error) {

	dst := i.pdst.path
	f, err := os.Create(dst)
	defer func() {
		err = dcheck.Wrap(f.Close(), err, "close %q", dst)
	}()
	url := i.psrc.Dll64Url
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http get %q: %w", url, err)
	}
	defer func() {
		err = dcheck.Wrap(resp.Body.Close(), err, "close body")
	}()
	if _, err = io.Copy(f, resp.Body); err != nil {
		return fmt.Errorf("copy to %q: %w", dst, err)
	}
	return err
}
