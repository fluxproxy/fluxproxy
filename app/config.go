package app

import (
	"context"
	"fmt"
	"github.com/knadh/koanf/v2"
	"github.com/rocket-proxy/rocket-proxy"
)

const (
	configPathAuthenticator = "authenticator"
	configPathResolver      = "resolver"
	configPathRuleset       = "ruleset"
	configPathServer        = "server"
	configPathServerHttp    = "server.http"
	configPathServerSocks   = "server.socks"
)

func unmarshalWith(ctx context.Context, path string, out any) error {
	if err := rocket.Configer(ctx).UnmarshalWithConf(path, out, koanf.UnmarshalConf{Tag: "yaml"}); err != nil {
		return fmt.Errorf("config unmarshal %s. %w", path, err)
	}
	return nil
}

////

type ServerConfig struct {
	Mode    string `yaml:"mode"`
	Verbose bool   `yaml:"verbose"`
}

////

type HttpConfig struct {
	Disabled bool   `yaml:"disabled"`
	Bind     string `yaml:"bind"`
	Port     int    `yaml:"port"`
}

////

type SocksConfig struct {
	Disabled bool   `yaml:"disabled"`
	Bind     string `yaml:"bind"`
	Port     int    `yaml:"port"`
}

////

type ResolverConfig struct {
	CacheSize int               `yaml:"cache_size"`
	CacheTTL  int               `yaml:"cache_ttl"`
	Hosts     map[string]string `yaml:"hosts"`
}

////

type AuthenticatorConfig struct {
	Enabled bool              `yaml:"enabled"`
	Basic   map[string]string `yaml:"basic"`
}

////

type RulesetConfig struct {
	Type    string   `yaml:"type"`
	Origin  string   `yaml:"origin"`
	Access  string   `yaml:"access"`
	Address []string `yaml:"address"`
}
