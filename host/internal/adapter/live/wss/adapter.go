package wss

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"filippo.io/age"
	"github.com/coder/websocket"
	"github.com/google/uuid"
	"github.com/malikbenkirane/groq-whisper/setup/pkg/dcheck"
)

func New(ctx context.Context) (_ Adapter, err error) {
	id, err := age.GenerateX25519Identity()
	if err != nil {
		return nil, fmt.Errorf("generate age identity: %w", err)
	}
	f, err := os.OpenFile("age.key", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return nil, fmt.Errorf("age.key: %w", err)
	}
	defer func() {
		err = dcheck.Wrap(f.Close(), err, "close age.key")
	}()
	_, err = io.Copy(f, strings.NewReader(id.String()))
	if err != nil {
		return nil, fmt.Errorf("copy key to age.key: %w", err)
	}
	slog.Debug("age.key ready to use in cwd", "pub", id.Recipient().String())

	conf := DefaultConfig()
	mux := http.NewServeMux()
	a := &adapter{
		config:   conf,
		mux:      mux,
		ctx:      ctx,
		channels: make(map[string]*channel),
		sockets:  make(map[uuid.UUID]*socket),
		ageId:    id,
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		insecureSkip := len(conf.origins) == 0
		conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
			OriginPatterns:     conf.origins,
			InsecureSkipVerify: insecureSkip,
		})
		if err != nil {
			slog.Warn("websocket accept", "err", err)
			return
		}

		socket := &socket{
			conn: conn,
			id:   uuid.New(),
		}

		if err := func() (err error) {
			defer func() {
				if err != nil {
					err = dcheck.Wrap(
						conn.Close(websocket.StatusInternalError, "sorry"),
						err, "close connection")
					slog.Info("delted connection", "socket_id", socket.id.String())
					if socket.ch != nil {
						delete(a.channels, socket.ch.member)
						slog.Info("bye", "member", socket.ch.member, "socket", socket.id.String())
					}
					socket.conn = nil
				}
			}()
			for {
				select {
				case <-ctx.Done():
					slog.Info("ending websocket")
					return
				default:
					typ, p, err := conn.Read(r.Context())
					if errors.Is(err, io.EOF) {
						return nil
					}
					if err != nil {
						slog.Warn("conn read", "err", err)
						continue
					}

					errChan := make(chan error)
					defer close(errChan)

					go func() {
						for err := range errChan {
							if err != nil {
								slog.Warn("handle", "err", err, "socket_id", socket.id.String())
							}
						}
					}()
					go func() {
						errChan <- a.handle(p, typ, socket)
					}()
				}
			}
		}(); err != nil {
			slog.Error("websocket loop", "err", err)
		}
	})
	return a, nil
}

type Adapter interface {
	Serve() error
}

func DefaultConfig() Config {
	return Config{
		addr:    ":50002",
		origins: []string{},
	}
}

type Config struct {
	addr    string
	origins []string
}

type adapter struct {
	config Config
	mux    *http.ServeMux
	ctx    context.Context

	channels map[string]*channel
	sockets  map[uuid.UUID]*socket

	ageId *age.X25519Identity
}

func (a adapter) Serve() error {
	slog.Info("serving wss/age", "addr", a.config.addr)
	return http.ListenAndServe(a.config.addr, a.mux)
}
