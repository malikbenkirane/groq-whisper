package wss

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/coder/websocket"
	"github.com/google/uuid"
)

type socket struct {
	id   uuid.UUID
	conn *websocket.Conn
	ch   *channel
}

func (s socket) isOpen() bool { return s.conn != nil }

type channel struct {
	member    string
	recipient string
	socket    *socket
}

func pevent(t eventType) *eventType {
	p := new(eventType)
	*p = t
	return p
}

//go:generate stringer -type=eventType
type eventType int

const (
	eventError eventType = iota
	eventJoin
	eventKeyBroadcast
	eventThemeLock
	eventActorLock
	eventUnknown
)

func (a adapter) handle(p []byte, t websocket.MessageType, socket *socket) error {
	if t == websocket.MessageBinary {
		slog.Warn("ignoring binary message")
		return nil
	}
	var msg eventMsg
	if err := json.Unmarshal(p, &msg); err != nil {
		return fmt.Errorf("json unmarshal message %q: %w", string(p), err)
	}
	if msg.Event == nil {
		return fmt.Errorf("nil event in %q", string(p))
	}
	switch *msg.Event {
	case eventJoin:
		if err := msg.validateJoin(); err != nil {
			return fmt.Errorf("decodeJoin: %w", err)
		}
		if ch, ok := a.channels[msg.Member]; ok {
			slog.Debug("member may have already joined checking socket state", "member", msg.Member)
			isConnected := true
			var oldSocket string
			if ch.socket != nil {
				oldSocket = socket.id.String()
				slog.Debug("...found socket state", "member", msg.Member, "socket_id", oldSocket)
				isConnected = ch.socket.isOpen()
			}
			if isConnected {
				return fmt.Errorf("member %q already joined", msg.Member)
			}
			slog.Info("member rejoining with new socket", "member", msg.Member, "old_socket", oldSocket)
		}
		slog.Info("new member", "name", msg.Member, "socket", socket.id.String())
		ch := &channel{
			socket:    socket,
			member:    msg.Member,
			recipient: msg.Body.Join.Key,
		}
		a.channels[msg.Member] = ch
		socket.ch = ch
		a.sockets[socket.id] = socket
		if err := a.broadcast(eventMsg{
			Event: pevent(eventKeyBroadcast),
			Body: &msgBody{
				KeyBroadcast: &msgKeyBroadcast{
					Key: a.ageId.Recipient().String(),
				},
			},
		}); err != nil {
			return fmt.Errorf("broadcast key: %w", err)
		}
	case eventThemeLock:
		if err := a.decodeThemeLock(&msg); err != nil {
			return fmt.Errorf("decodeThemeLock: %w", err)
		}
		if err := a.broadcast(eventMsg{
			Body: &msgBody{
				tunnel: msgTunnel{
					IsHost: true,
					Event:  pevent(eventThemeLock),
					ThemeLock: &msgThemeLock{
						Name: msg.Body.tunnel.ThemeLock.Name,
						Lock: msg.Body.tunnel.ThemeLock.Lock,
					},
				},
			},
		}, msg.Body.tunnel.From); err != nil {
			return fmt.Errorf("broadcast theme lock: %q", err)
		}
	case eventActorLock:
		if err := a.decodeActorLock(&msg); err != nil {
			return fmt.Errorf("decodeActorLock: %w", err)
		}
		if err := a.broadcast(eventMsg{
			Body: &msgBody{
				tunnel: msgTunnel{
					IsHost: true,
					Event:  pevent(eventActorLock),
					ActorLock: &msgActorLock{
						Name: msg.Body.tunnel.ActorLock.Name,
						Lock: msg.Body.tunnel.ActorLock.Lock,
					},
				},
			},
		}, msg.Body.tunnel.From); err != nil {
			return fmt.Errorf("brodact actor lock: %w", err)
		}
	}
	return nil
}

var _ json.Unmarshaler = new(eventType)

func (t *eventType) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return fmt.Errorf("json unmarshal: %w", err)
	}
	for i := eventError; i < eventUnknown; i++ {
		if eventType(i).String() == s {
			*t = i
			return nil
		}
	}
	return fmt.Errorf("unknwon event %q", string(data))
}

var _ json.Marshaler = new(eventType)

func (t *eventType) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", t.String())), nil
}
