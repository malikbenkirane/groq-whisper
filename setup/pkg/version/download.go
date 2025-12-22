package version

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/malikbenkirane/groq-whisper/setup/pkg/dcheck"
	"github.com/malikbenkirane/groq-whisper/setup/pkg/gcloud"
)

type Upgrader interface {
	Upgrade() error
	Version() string
}

type Upgrade struct {
	Name     string
	Upgraded bool
}

type UpgraderConfig struct {
	VersionSrc string
	App        string
	Installer  string
}

type UpgraderOption func(UpgraderConfig) UpgraderConfig

func UpgraderWithVersionSource(url string) UpgraderOption {
	return func(uc UpgraderConfig) UpgraderConfig {
		uc.VersionSrc = url
		return uc
	}
}
func UpgraderWithApp(prefix string) UpgraderOption {
	return func(uc UpgraderConfig) UpgraderConfig {
		uc.App = prefix
		return uc
	}
}
func UpgraderWithInstaller(prefix string) UpgraderOption {
	return func(uc UpgraderConfig) UpgraderConfig {
		uc.Installer = prefix
		return uc
	}
}

func DefaultUpgraderConfig() UpgraderConfig {
	return UpgraderConfig{
		VersionSrc: Url,
		App:        "groq",
		Installer:  "groq-setup",
	}
}

func NewUpgrader(bucket string, opts ...UpgraderOption) (Upgrader, error) {
	c := DefaultUpgraderConfig()
	for _, opt := range opts {
		opt(c)
	}
	upstream, err := Get(c.VersionSrc)
	if err != nil {
		return nil, fmt.Errorf("get version: %w", err)
	}
	b := gcloud.NewBucket(bucket)
	return &upgrader{
		b:       b,
		c:       c,
		version: upstream,
	}, nil
}

type upgrader struct {
	b       gcloud.Bucket
	c       UpgraderConfig
	version string
}

func (u *upgrader) Version() string {
	return u.version
}

func (u *upgrader) Upgrade() error {
	var app, installer bytes.Buffer
	appUpgrade, installerUpgrade, err := u.download(&app, &installer)
	if err != nil {
		return err
	}
	for _, u := range []upgrade{
		{u.c.Installer, installer, *installerUpgrade},
		{u.c.App, app, *appUpgrade},
	} {
		if err := u.handle(); err != nil {
			return fmt.Errorf("handle upgrade: %w", err)
		}
	}
	return nil
}

func (u *upgrader) download(app, installer io.Writer) (appUpgrade, installerUpgrade *Upgrade, err error) {
	slog.Debug("downloading", "app", u.c.App)
	appUpgrade, err = u.downloadObject(app, u.c.App)
	if err != nil {
		err = fmt.Errorf("download app: %w", err)
		return
	}
	slog.Debug("downloading", "installer", u.c.Installer)
	installerUpgrade, err = u.downloadObject(installer, u.c.Installer)
	if err != nil {
		err = fmt.Errorf("download installer: %w", err)
		return
	}
	return
}

func (u upgrader) downloadObject(w io.Writer, object string) (*Upgrade, error) {
	if u.version == Version {
		return &Upgrade{
			Name:     Executable(object, Version),
			Upgraded: false,
		}, nil
	}
	exe := Executable(object, u.version)
	if err := u.b.Pull(w, exe); err != nil {
		return nil, fmt.Errorf("bucket pull: %w", err)
	}
	return &Upgrade{
		Name:     exe,
		Upgraded: true,
	}, nil
}

type upgrade struct {
	name    string
	buffer  bytes.Buffer
	upgrade Upgrade
}

func (u upgrade) handle() (err error) {
	if u.upgrade.Upgraded {
		return nil
	}
	var cpy bytes.Buffer
	{
		dst := u.upgrade.Name
		f, err := os.Create(dst)
		if err != nil {
			return fmt.Errorf("create dst %q: %w", dst, err)
		}
		defer func() {
			err = dcheck.Wrap(f.Close(), err, "close %q", dst)
		}()
		_, err = io.Copy(io.MultiWriter(&cpy, f), &u.buffer)
		if err != nil {
			return fmt.Errorf("copy upgrade buffer to %q", dst)
		}
	}
	exe := Executable(u.name, "")
	_, err = os.Stat(exe)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat %q: %w", exe, err)
	}
	if err != nil { // error is os.ErrNotExist
		exeBackup := Executable(u.name, Version)
		if err := os.Rename(exe, exeBackup); err != nil {
			return fmt.Errorf("backup %q to %q", exe, exeBackup)
		}
	}
	{
		f, err := os.Create(exe)
		if err != nil {
			return fmt.Errorf("create exe %q: %w", exe, err)
		}
		defer func() {
			err = dcheck.Wrap(f.Close(), err, "close %q", exe)
		}()
		_, err = io.Copy(f, &cpy)
		if err != nil {
			return fmt.Errorf("copy upgrade to %q", exe)
		}
	}
	return nil
}
