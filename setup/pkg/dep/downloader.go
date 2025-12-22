package dep

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/bodgit/sevenzip"
	"github.com/schollz/progressbar/v3"

	"github.com/malikbenkirane/groq-whisper/setup/pkg/dcheck"
	"github.com/malikbenkirane/groq-whisper/setup/pkg/gcloud"
)

func NewInstaller(path, bucket, ffmpeg7z string, opts ...InstallerOption) (Installer, error) {
	conf := InstallerConfig{}
	for _, opt := range opts {
		conf = opt(conf)
	}
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
		conf:     conf,
	}, nil
}

type Installer interface {
	Download() error
}

type InstallerConfig struct {
	overwrite bool
}

type InstallerOption func(InstallerConfig) InstallerConfig

func InstallerWithOverwrite() InstallerOption {
	return func(ic InstallerConfig) InstallerConfig {
		ic.overwrite = true
		return ic
	}
}

type installer struct {
	psrc     PortAudioSrc
	pdst     *PortAudioDst
	bucket   gcloud.Bucket
	ffmpeg7z string
	path     string

	conf InstallerConfig
}

func (i installer) Download() error {
	return i.download()
}

func (i installer) download() (err error) {
	if i.conf.overwrite {
		if err := os.RemoveAll(i.path); err != nil {
			return fmt.Errorf("remove all %q: %w", i.path, err)
		}
	}
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
			bar := progressbar.NewOptions(len(reader.File),
				progressbar.OptionShowDescriptionAtLineEnd(),
				progressbar.OptionClearOnFinish(),
				progressbar.OptionSetDescription("install ffmpeg"),
				progressbar.OptionSetMaxDetailRow(8))
			for _, f := range reader.File {
				dst, err := i.extract(f)
				if err != nil {
					return fmt.Errorf("extract: %w", err)
				}
				if err := bar.AddDetail(dst); err != nil {
					return fmt.Errorf("bar add: %w", err)
				}
			}
		}
	}
	return nil
}

func (i installer) extract(fz *sevenzip.File) (dst string, err error) {
	rc, err := fz.Open()
	if err != nil {
		return "", fmt.Errorf("7z open %q: %w", fz.Name, err)
	}
	defer func() {
		err = dcheck.Wrap(rc.Close(), err, "close %q", fz.Name)
	}()
	parts := strings.Split(fz.Name, "/")
	parts[0] = i.path
	dst = path.Join(parts...)
	_, err = os.Stat(dst)
	if !errors.Is(err, os.ErrNotExist) {
		if !i.conf.overwrite {
			return dst, nil
		}
	}
	if err != nil {
		return "", fmt.Errorf("stat %q: %w", dst, err)
	}
	if parts[len(parts)-1] == "" {
		if err := os.MkdirAll(dst, 0700); err != nil {
			return "", fmt.Errorf("mkdir %q: %w", dst, err)
		}
		return dst, nil
	}
	f, err := os.Create(dst)
	if err != nil {
		return "", fmt.Errorf("create %q: %w", dst, err)
	}
	defer func() {
		err = dcheck.Wrap(f.Close(), err, "close %q", dst)
	}()
	_, err = io.Copy(f, rc)
	if err != nil {
		return "", fmt.Errorf("install to %q: %w", dst, err)
	}
	return dst, nil
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
