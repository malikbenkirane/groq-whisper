package server

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/malikbenkirane/groq-whisper/internal/sampler"
)

func defaultConfig(root string) (Config, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return Config{}, err
	}
	s := sampler.DefaultSys32(root)
	return Config{
		http:      ":50002",
		port:      7946,
		master:    "192.168.117.1:7496",
		name:      base64.StdEncoding.EncodeToString(b),
		startHttp: true,
		startLoop: true,
		sampler:   s,
	}, nil
}

type Option func(Config) Config

func OptionSerfPort(port int) Option {
	return func(c Config) Config {
		c.port = port
		return c
	}
}

func OptionSerfMaster(host string) Option {
	return func(c Config) Config {
		c.master = host
		return c
	}
}

func OptionSerfName(name string) Option {
	return func(c Config) Config {
		c.name = name
		return c
	}
}

func OptionHttpAddr(addr string) Option {
	return func(c Config) Config {
		c.http = addr
		return c
	}
}

func OptionSampler(s sampler.Sampler) Option {
	return func(c Config) Config {
		c.sampler = s
		return c
	}
}

func OptionNoHttp() Option {
	return func(c Config) Config {
		c.startHttp = false
		return c
	}
}

func OptionNoLoop() Option {
	return func(c Config) Config {
		c.startLoop = false
		return c
	}
}
