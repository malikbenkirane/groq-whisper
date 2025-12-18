package version

import "strings"

type ServerConfig struct {
	secretKey string
}

type ServerOption func(ServerConfig) ServerConfig

func WithSecret(key string) ServerOption {
	return func(sc ServerConfig) ServerConfig {
		sc.secretKey = strings.TrimSpace(key)
		return sc
	}
}
