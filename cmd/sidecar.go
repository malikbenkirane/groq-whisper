package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

func groqKey(file string) (string, error) {
	var b bytes.Buffer
	f, err := os.Open(file)
	if err != nil {
		return "", fmt.Errorf("open %q: %w", file, err)
	}
	if _, err = io.Copy(&b, f); err != nil {
		return "", fmt.Errorf("copy key: %w", err)
	}
	return strings.TrimSpace(b.String()), nil
}

type groqClient struct {
	key   string
	log   *zap.Logger
	tx    groqTx
	lang  string // iso-693-1
	start time.Time
	root  string
}

func (gc groqClient) file() (io.WriteCloser, error) {
	p := gc.start.Format("2006-01-02_15-04-05.000.txt")
	base := path.Join(gc.root, "whisper-v3-tx")
	_, err := os.Stat(base)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("stat %q: %w", base, err)
	}
	if err != nil { // error is os.ErrNotExist
		if err := os.MkdirAll(base, 0700); err != nil {
			return nil, fmt.Errorf("mkdir base %q: %w", base, err)
		}
	}
	p = path.Join(base, p)
	f, err := os.OpenFile(p, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return nil, fmt.Errorf("open file %q: %w", p, err)
	}
	return f, nil
}

type groqTx struct {
	Text  string `json:"text"`
	XGroq groqX  `json:"x_groq"`
}

type groqX struct {
	Id string `json:"id"`
}

const gtxUrl = "https://api.groq.com/openai/v1/audio/transcriptions"

func (gc groqClient) newRequest(audio, model string) (req *http.Request, err error) {
	gc.log = gc.log.Named("gc request builder")
	defer func() {
		if err != nil {
			gc.log.Error("groq client failed to build request", zap.Error(err))
		}
	}()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	base := filepath.Base(audio)
	gc.log.Debug("create form file", zap.String("base", base))
	part, err := writer.CreateFormFile("file", base)
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	gc.log.Debug("opening audio", zap.String("file", audio))
	stat, err := os.Stat(audio)
	if err != nil {
		return nil, fmt.Errorf("os stat %q: %w", audio, err)
	}
	f, err := os.Open(audio)
	if err != nil {
		return nil, fmt.Errorf("open %q: %w", audio, err)
	}
	defer func() {
		err = f.Close()
	}()
	n, err := io.Copy(part, f)
	if err != nil {
		return nil, fmt.Errorf("copy file part: %w", err)
	}
	gc.log.Debug("form file written", zap.Int64("copied", n), zap.Int64("stated", stat.Size()))
	gc.log.Debug("write model field", zap.String("model", model))
	for k, v := range map[string]string{
		"language": gc.lang,
		"model":    model,
	} {
		if err = writer.WriteField(k, v); err != nil {
			return nil, fmt.Errorf("write field %q: %w", k, err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("multipart writer close: %w", err)
	}
	gc.log.Debug("prepare post request", zap.String("url", gtxUrl))
	req, err = http.NewRequest(http.MethodPost, gtxUrl, &body)
	if err != nil {
		return nil, fmt.Errorf("http new request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Add("Authorization", "bearer "+gc.key)
	return
}

func (gc *groqClient) post(audio, model string) (err error) {
	defer func() {
		if err != nil {
			gc.log.Error("groq client failed to post", zap.Error(err))
		}
	}()
	gc.log.Debug("new request",
		zap.String("audio", audio), zap.String("model", model))
	req, err := gc.newRequest(audio, model)
	if err != nil {
		return fmt.Errorf("gc new request: %w", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("response: %q", resp.Status)
	}
	defer func() {
		err = resp.Body.Close()
	}()
	var tx groqTx
	if err = json.NewDecoder(resp.Body).Decode(&tx); err != nil {
		return fmt.Errorf("decode response body: %w", err)
	}
	gc.tx = tx
	return nil
}

func newCommandSidecar() (*cobra.Command, error) {
	var dry, debug *bool
	var samplesDir *string
	cmd := &cobra.Command{
		Use:     "sidecar",
		Aliases: []string{"watch", "w", "s"},
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			log := newLogger(*debug)

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

			defer func() {
				if err != nil {
					log.Error("sidecar failed", zap.Error(err))
				}
			}()

			var gc groqClient
			{
				key, err := groqKey("key.txt")
				if err != nil {
					return fmt.Errorf("groq key: %w", err)
				}

				gc = groqClient{
					key:   key,
					log:   log,
					lang:  "fr",
					start: time.Now(),
					root:  *samplesDir,
				}
				log.Debug("new client",
					zap.String("lang", gc.lang), zap.String("url", gtxUrl))
			}

			txOut, err := gc.file()
			if err != nil {
				return fmt.Errorf("tx out: %w", err)
			}
			defer func() {
				if err := txOut.Close(); err != nil {
					slog.Warn("unable to close txOut", "err", err)
				}
			}()

			w, err := fsnotify.NewWatcher()
			if err != nil {
				return fmt.Errorf("fsnotify new watcher: %w", err)
			}
			defer func() {
				err = w.Close()
			}()

			ctx, cancel := context.WithCancel(cmd.Context())
			go func(ctx context.Context) {
				counter := make(map[string]struct{})
			loop:
				for {
					select {
					case event, ok := <-w.Events:
						if !ok {
							return
						}
						log.Debug("fsnotify event",
							zap.String("event", event.Op.String()),
							zap.String("name", event.Name),
						)
						if event.Has(fsnotify.Write) &&
							strings.HasSuffix(event.Name, ".flac") {
							if _, ok := counter[event.Name]; ok {
								continue loop
							}
							counter[event.Name] = struct{}{}
							stat, err := os.Stat(event.Name)
							if err != nil {
								log.Error("unable to stat sample", zap.String("file", event.Name))
								continue loop
							}
							log.Info("found new sample",
								zap.String("file", event.Name), zap.Int64("size", stat.Size()))
							if !*dry {
								if err = gc.post(event.Name, "whisper-large-v3"); err != nil {
									log.Error("gc post failed", zap.Error(err))
									continue loop
								}
								_, err := io.Copy(io.MultiWriter(os.Stdout, txOut), strings.NewReader(gc.tx.Text+"\n"))
								if err != nil {
									slog.Error("tx not written", "err", err)
									continue loop
								}
							}
						}
					case err, ok := <-w.Errors:
						if !ok {
							return
						}
						log.Error("fsnotify", zap.Error(err))
					}
				}
			}(ctx)

			if err = w.Add(*samplesDir); err != nil {
				cancel()
				return fmt.Errorf("fsnotify add cwd: %w", err)
			}

			<-quit
			cancel()

			return nil
		},
	}

	defaultDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("user home dir: %w", err)
	}

	defaultDir = path.Join(defaultDir, "groq-whisper-samples")
	dry = cmd.Flags().Bool("dry", false, "don't post to groq")
	debug = cmd.Flags().Bool("debug", false, "set log level at debug")
	samplesDir = cmd.Flags().String("samples-dir", defaultDir, "where recorded samples are processed")

	return cmd, nil
}
