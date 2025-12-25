package server

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/hashicorp/serf/serf"
	"github.com/malikbenkirane/groq-whisper/internal/sampler"
)

func New(root string, opts ...Option) (Server, error) {

	sv := &server{}
	sv.swarm = make(map[string]bool)

	c, err := defaultConfig(root)
	if err != nil {
		return nil, fmt.Errorf("default config: %w", err)
	}
	for _, opt := range opts {
		c = opt(c)
	}
	sv.conf = c

	events := make(chan serf.Event)
	conf := serf.DefaultConfig()
	conf.Init()
	conf.NodeName = c.name
	conf.MemberlistConfig.BindAddr = "0.0.0.0"
	conf.MemberlistConfig.BindPort = c.port
	conf.EventCh = events
	sv.serfCh = events

	slog.Info("serf config",
		"bind_addr", conf.MemberlistConfig.BindAddr,
		"bind_port", conf.MemberlistConfig.BindPort,
		"node_name", conf.NodeName,
		"join_addr", c.master)

	instance, err := serf.Create(conf)
	if err != nil {
		return nil, fmt.Errorf("serf create: %w", err)
	}
	sv.serf = instance

	sig := make(chan signal)
	sv.sig = sig

	mux := http.NewServeMux()
	mux.Handle("POST /record", wrap(func(w http.ResponseWriter, r *http.Request) (err error) {
		slog.Info("POST /record")
		sig <- signalStart
		slog.Info("signaled start")
		return nil
	}))
	mux.Handle("DELETE /record", wrap(func(w http.ResponseWriter, r *http.Request) (err error) {
		slog.Info("DELETE /record")
		sig <- signalStop
		slog.Info("signaled stop")
		return nil
	}))
	mux.Handle("GET /state", wrap(func(w http.ResponseWriter, r *http.Request) (err error) {
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(sv.swarm)
	}))
	sv.mux = mux

	return sv, nil

}

type Server interface {
	Serve(ctx context.Context)
}

type signal int

const (
	signalStart signal = iota
	signalStop
)

type Config struct {
	port   int
	master string
	name   string
	http   string

	startHttp bool
	startLoop bool

	sampler sampler.Sampler
}

type server struct {
	mux    *http.ServeMux
	serf   *serf.Serf
	serfCh chan serf.Event
	conf   Config
	sig    chan signal

	swarm map[string]bool
}

func (s server) Serve(ctx context.Context) {
	ech := make(chan error)
	defer close(ech)
	go func() {
		for {
			select {
			case e, ok := <-ech:
				if !ok {
					slog.Info("err chan was closed")
				}
				slog.Error("serve messed up", "err", e)
			case <-ctx.Done():
				slog.Info("err check routine: context done")
			}
		}
	}()
	if s.conf.startHttp {
		go s.serveHttp(ech)
	}
	go s.gossip()
	if s.conf.master != "" {
		_, err := s.serf.Join([]string{s.conf.master}, false)
		if err != nil {
			ech <- err
		}
	}
	if !s.conf.startLoop {
		<-ctx.Done()
		return
	}
	sampling := false
	var (
		scope  context.Context
		cancel context.CancelFunc
	)
loop:
	for {
		select {
		case signal, ok := <-s.sig:
			if !ok {
				slog.Info("sig channel closed")
				return
			}
			switch signal {
			case signalStart:
				if sampling {
					slog.Warn("sampler asked to start twice")
					continue loop
				}
				scope, cancel = context.WithCancel(ctx)
				defer cancel()
				go s.conf.sampler.Sample(scope)
				sampling = true
			case signalStop:
				if !sampling {
					slog.Warn("sampler asked to stop twice")
					continue loop
				}
				cancel()
				sampling = false
			}
		case <-ctx.Done():
			slog.Info("serve done")
			return
		}
	}
}

func (s server) serveHttp(err chan error) {
	err <- http.ListenAndServe(s.conf.http, s.mux)
}

func (s server) gossip() {
	for e := range s.serfCh {
		switch ev := e.(type) {
		case serf.MemberEvent:
			for _, m := range ev.Members {
				switch ev.EventType() {
				case serf.EventMemberJoin:
					slog.Info("serf: member joined", "name", m.Name, "addr", m.Addr)
					s.swarm[m.Addr.String()] = true
				case serf.EventMemberFailed:
					slog.Warn("serf: member failed", "name", m.Name, "addr", m.Addr)
					s.swarm[m.Addr.String()] = false
				case serf.EventMemberLeave:
					slog.Info("serf: member leaved", "name", m.Name, "addr", m.Addr)
					s.swarm[m.Addr.String()] = false
				default:
					slog.Info("serf: other event", "type", ev.EventType(), "from", m.Name)
				}
			}
		}
	}
}

type customHandler func(w http.ResponseWriter, r *http.Request) (err error)

func wrap(h customHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}
}

// for sec generate host
// func () {
// 	addrs,err := net.InterfaceAddrs()
// 	if err != nil {
// 		return fmt.Errorf("interface addrs: %w", err)
// 	}
// 	for _, addr := range addrs {
// 		if ipnet, check := addr.(*net.IPNet); check &&
// 		!ipnet.IP.IsLoopback() {
// 			if ipnet.IP.To4() != nil {
// 			}
// 		}
// 	}
// }
