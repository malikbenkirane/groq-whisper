package dep

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/bodgit/sevenzip"
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
	Install() error
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

func (i installer) Install() error {
	if err := i.installFfmpeg(); err != nil {
		return fmt.Errorf("install ffmpeg: %w", err)
	}
	if err := i.installPortaudio(); err != nil {
		return fmt.Errorf("install portaudio: %w", err)
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
	if parts[len(parts)-1] == "" {
		if err := os.MkdirAll(dst, 0700); err != nil {
			return "", fmt.Errorf("mkdir %q: %w", dst, err)
		}
		return dst, nil
	}
	_, err = os.Stat(dst)
	if err == nil {
		if !i.conf.overwrite {
			return dst, nil
		}
	}
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", fmt.Errorf("stat %q: %w", dst, err)
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
