package gcloud

import (
	"fmt"
	"io"
	"net/http"

	"github.com/schollz/progressbar/v3"
)

const (
	storageUrl = "https://storage.googleapis.com"
)

type Bucket interface {
	Push(dst, src string) error
	Pull(dst io.Writer, name string) error
}

type BucketConfig struct {
	login bool
}

type BucketOption func(BucketConfig) BucketConfig

func BucketWithLogin() BucketOption {
	return func(bc BucketConfig) BucketConfig {
		bc.login = true
		return bc
	}
}

func NewBucket(name string, opts ...BucketOption) Bucket {
	c := BucketConfig{}
	for _, opt := range opts {
		c = opt(c)
	}
	return &bucket{name: name, c: c}
}

type bucket struct {
	name string
	c    BucketConfig
}

func (b bucket) Pull(dst io.Writer, name string) (err error) {
	url := b.getUrl(name)
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("http get %q: %w", url, err)
	}
	defer func() {
		err = errDefer(resp.Body.Close(), err)
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status %q", resp.Status)
	}
	bar := progressbar.DefaultBytes(resp.ContentLength, fmt.Sprintf("pulling %s", url))
	_, err = io.Copy(io.MultiWriter(bar, dst), resp.Body)
	return err
}

func (b bucket) Push(dst, src string) error {
	if b.c.login {
		err := gcloud(nil, "auth", "login")
		if err != nil {
			return fmt.Errorf("gcloud auth login: %w", err)
		}
	}
	dst = b.pushUrl(dst)
	err := gcloud(nil, "storage", "cp", "file://"+src, dst)
	if err != nil {
		return fmt.Errorf("gcloud storage cp %q %q: %w", src, dst, err)
	}
	return nil
}

func (b bucket) getUrl(name string) string {
	return fmt.Sprintf("%s/%s/%s", storageUrl, b.name, name)
}

func (b bucket) pushUrl(name string) string {
	return fmt.Sprintf("gs://%s/%s", b.name, name)
}

func errDefer(errDefer, err error) error {
	if err != nil && errDefer != nil {
		return fmt.Errorf("%w then %w", err, errDefer)
	}
	if errDefer != nil {
		return errDefer
	}
	return err
}
