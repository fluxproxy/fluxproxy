package app

import (
	"context"
	"fmt"
	"github.com/knadh/koanf/v2"
)

const (
	configPathAuthenticator = "authenticator"
	configPathResolver      = "resolver"
	configPathRuleset       = "ruleset"
	configPathServer        = "server"
	configPathServerHttp    = "server.http"
	configPathServerSocks   = "server.socks"
)

////

type ServerConfig struct {
	Mode    string `toml:"mode"`
	Verbose bool   `toml:"verbose"`
}

////

type HttpConfig struct {
	Disabled bool   `toml:"disabled"`
	Bind     string `toml:"bind"`
	Port     int    `toml:"port"`
}

////

type SocksConfig struct {
	Disabled bool   `toml:"disabled"`
	Bind     string `toml:"bind"`
	Port     int    `toml:"port"`
}

////

type ResolverConfig struct {
	CacheSize int               `toml:"cache_size"`
	CacheTTL  int               `toml:"cache_ttl"`
	Hosts     map[string]string `toml:"hosts"`
}

////

type AuthenticatorConfig struct {
	Enabled bool              `toml:"enabled"`
	Basic   map[string]string `toml:"basic"`
}

////

type RulesetConfig struct {
	Type    string   `toml:"type"`
	Origin  string   `toml:"origin"`
	Access  string   `toml:"access"`
	Address []string `toml:"address"`
}

////

func unmarshalWith(ctx context.Context, path string, out any) error {
	if err := proxy.Configer(ctx).UnmarshalWithConf(path, out, koanf.UnmarshalConf{Tag: "toml"}); err != nil {
		return fmt.Errorf("config unmarshal %s. %w", path, err)
	}
	return nil
}
