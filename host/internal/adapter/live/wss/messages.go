package wss

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"filippo.io/age"
	"github.com/coder/websocket"
	"github.com/malikbenkirane/groq-whisper/setup/pkg/dcheck"
)

func (msg *eventMsg) validateJoin() error {
	if msg.Body == nil {
		return errNilBody
	}
	if msg.Body.Join == nil {
		return errNilBodyJoin
	}
	s := msg.Body.Join.Key
	_, err := age.ParseX25519Recipient(s)
	if err != nil {
		return fmt.Errorf("age parse recipient %q: %w", s, err)
	}
	return nil
}

func (a adapter) decodeThemeLock(msg *eventMsg) error {
	if err := a.decodeCipher(msg); err != nil {
		return fmt.Errorf("decode cipher: %w", err)
	}
	if msg.Body.tunnel.ThemeLock == nil {
		return errNilBodyThemeLock
	}
	return nil
}

func (a adapter) decodeActorLock(msg *eventMsg) error {
	if err := a.decodeCipher(msg); err != nil {
		return fmt.Errorf("decode cipher: %w", err)
	}
	if msg.Body.tunnel.ActorLock == nil {
		return errNilBodyActorLock
	}
	return nil
}

func (a adapter) decodeCipher(msg *eventMsg) error {
	if msg.Body == nil {
		return errNilBody
	}
	// slog.Debug("decodeCipher", "public_key", a.ageId.Recipient(), "tunnel", msg.Body.Tunnel)
	r := base64.NewDecoder(base64.StdEncoding, strings.NewReader(msg.Body.Tunnel))
	r, err := age.Decrypt(r, a.ageId)
	if err != nil {
		return fmt.Errorf("%w: %w", errAgeDecrypt, err)
	}
	var mt msgTunnel
	if err := json.NewDecoder(r).Decode(&mt); err != nil {
		return fmt.Errorf("%w: %w: %w", errAgeDecrypt, errJsonEncode, err)
	}
	msg.Body.tunnel = mt
	return nil
}

type eventMsg struct {
	Member string     `json:"member"`
	Event  *eventType `json:"event"`
	Body   *msgBody   `json:"payload"`
}

type msgBody struct {
	Tunnel string `json:"tunnel"`

	KeyBroadcast *msgKeyBroadcast `json:"key_broadcast"`
	Join         *msgJoin         `json:"join"`

	tunnel msgTunnel
}

type msgTunnel struct {
	IsHost bool       `json:"is_host"`
	From   string     `json:"from"`
	Event  *eventType `json:"event"`

	ThemeLock *msgThemeLock `json:"theme_lock"`
	ActorLock *msgActorLock `json:"actor_lock"`

	tunnel string
}

type msgThemeLock struct {
	Name string `json:"name"`
	Lock bool   `json:"lock"`
}

type msgActorLock struct {
	Name string `json:"name"`
	Lock bool   `json:"lock"`
}

type msgJoin struct {
	Key string `json:"key"`
}

type msgKeyBroadcast struct {
	Key string `json:"key"`
}

//go:generate stringer -type=errDecode
type errDecode int

const (
	errNilBody errDecode = iota
	errNilBodyJoin
	errNilBodyThemeLock
	errNilBodyActorLock
	errNilSocket
	errSocketClosed
	errConnWriter
	errCloseConnWriter
	errCipher
	errJsonEncode
	errAgeEncrypt
	errAgeDecrypt
	errCloseAgeEncrypt
	errCloseBase64Encoder
	errCopyBase64
)

func (err errDecode) Error() string {
	return fmt.Sprintf("decode message: %s", err.String())
}

func (mt *msgTunnel) cipher(r string) (err error) {
	var out bytes.Buffer
	recipient, err := age.ParseX25519Recipient(r)
	if err != nil {
		return fmt.Errorf("parse recipient %q: %w", r, err)
	}
	age, err := age.Encrypt(&out, recipient)
	if err != nil {
		return fmt.Errorf("%s: %w: %w", mt.tag(), errAgeEncrypt, err)
	}
	if err := json.NewEncoder(age).Encode(mt); err != nil {
		return fmt.Errorf("%s: %w: %w", mt.tag(), errJsonEncode, err)
	}
	if err := age.Close(); err != nil {
		return fmt.Errorf("%s: %w: %w", mt.tag(), errCloseAgeEncrypt, err)
	}
	var b64 bytes.Buffer
	b64e := base64.NewEncoder(base64.StdEncoding, &b64)
	if _, err := io.Copy(b64e, &out); err != nil {
		return fmt.Errorf("%s: %w: %w", mt.tag(), errCopyBase64, err)
	}
	if err := b64e.Close(); err != nil {
		return fmt.Errorf("%s: %w: %w", mt.tag(), errCloseBase64Encoder, err)
	}
	mt.tunnel = b64.String()
	return nil
}

func (ch channel) tag() string {
	if ch.socket == nil {
		return fmt.Sprintf("member %q (nil socket)", ch.member)
	}
	return fmt.Sprintf("member %q socket %q", ch.member, ch.socket.id)
}

func (mt msgTunnel) tag() string {
	if mt.IsHost {
		return "(tunnel) host msg"
	}
	return fmt.Sprintf("(tunnel) msg from %q", mt.From)
}

func (a adapter) broadcast(msg eventMsg, skip ...string) (err error) {
	{
		var debug bytes.Buffer
		encoder := json.NewEncoder(&debug)
		if err := encoder.Encode(msg.Body.tunnel); err != nil {
			return fmt.Errorf("%w: %w", errJsonEncode, err)
		}
		slog.Debug("(tunnel) broadcast", "tunnel", debug.String())
		debug.Reset()
		if err := encoder.Encode(msg); err != nil {
			return fmt.Errorf("%w: %w", errJsonEncode, err)
		}
		slog.Debug("broadcast", "msg", &debug)
	}
loop:
	for member, ch := range a.channels {
		for _, skip := range skip {
			if member == skip {
				slog.Info("broadcast skip", "member", member)
				continue loop
			}
		}
		if err := a.send(*ch, msg); err != nil {
			slog.Error("broadcast message", "err", err)
			if !errors.Is(err, errNilSocket) {
				delete(a.sockets, ch.socket.id)
			}
			delete(a.channels, member)
			slog.Debug("member deleted", "member", member)
		}

	}
	return nil
}

func (a *adapter) send(ch channel, msg eventMsg) (err error) {
	if ch.socket == nil {
		return fmt.Errorf("%s: %w", ch.tag(), errNilSocket)
	}
	if !ch.socket.isOpen() {
		return fmt.Errorf("%s: %w", ch.tag(), errSocketClosed)
	}
	w, err := ch.socket.conn.Writer(a.ctx, websocket.MessageText)
	if err != nil {
		return fmt.Errorf("%s: %w: %w", ch.tag(), errConnWriter, err)
	}
	defer func() {
		err = dcheck.Wrap(w.Close(), err, "%s: %s", ch.tag(), errCloseConnWriter)
	}()
	if msg.Event == nil {
		tunnel := &msg.Body.tunnel
		if err := tunnel.cipher(ch.recipient); err != nil {
			return fmt.Errorf("%s: %w: %w", ch.tag(), errCipher, err)
		}
		msg.Body.Tunnel = tunnel.tunnel
		slog.Debug("cipher", "member", ch.member, "id", ch.recipient)
	}
	if err := json.NewEncoder(w).Encode(msg); err != nil {
		return fmt.Errorf("%s: %w: %w: %w", ch.tag(), errCipher, errJsonEncode, err)
	}
	return nil
}
