package dep

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/bodgit/sevenzip"

	"github.com/malikbenkirane/groq-whisper/setup/pkg/gcloud"
)

func NewInstaller(path, bucket, ffmpeg7z string) (Installer, error) {
	pdst, err := DefaultPortAudioDst()
	if err != nil {
		return nil, fmt.Errorf("default port audio dest: %w", err)
	}
	return &installer{
		pdst:     pdst,
		psrc:     DefaultPortAudioSrc(),
		bucket:   gcloud.NewBucket(bucket),
		ffmpeg7z: ffmpeg7z,
		path:     path,
	}, nil
}

type Installer interface {
	Download() error
}

type installer struct {
	psrc     PortAudioSrc
	pdst     *PortAudioDst
	bucket   gcloud.Bucket
	ffmpeg7z string
	path     string
}

func (i installer) Download() error {
	return i.download()
}

func (i installer) download() (err error) {
	{
		var archive bytes.Buffer
		err := i.bucket.Pull(&archive, i.ffmpeg7z)
		if err != nil {
			return fmt.Errorf("bucket pull %q: %w", i.ffmpeg7z, err)
		}
		reader := bytes.NewReader(archive.Bytes())
		{
			reader, err := sevenzip.NewReader(reader, reader.Size())
			if err != nil {
				return fmt.Errorf("7z new reader: %w", err)
			}
			for _, f := range reader.File {
				rc, err := f.Open()
				if err != nil {
					return fmt.Errorf("7z open %q: %w", f.Name, err)
				}
				if err := rc.Close(); err != nil {
					return fmt.Errorf("7z close %q: %w", f.Name, err)
				}
				fmt.Println(f.Name)
			}
		}
	}
	return nil
}

func DefaultPortAudioDst() (*PortAudioDst, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("user home dir: %w", err)
	}
	parts := strings.Split(home, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("unable to extract volume from %q", home)
	}
	volume := parts[0]
	return &PortAudioDst{
		path: path.Join(volume, "Windows", "System32", "libportaudio.dll"),
	}, nil
}

type PortAudioDst struct {
	path string
}
