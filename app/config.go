package app

const (
	configPathResolver = "resolver"
	configPathServer   = "server"
	configPathHttp     = "server.http"
	configPathSocks    = "server.socks"
)

////

type ServerConfig struct {
	Mode    string `yaml:"mode"`
	Verbose bool   `yaml:"verbose"`
}

////

type HttpAuthConfig struct {
	Enabled bool              `yaml:"enabled"`
	Basic   map[string]string `yaml:"basic"`
}

type HttpConfig struct {
	Disabled bool           `yaml:"disabled"`
	Bind     string         `yaml:"bind"`
	Port     int            `yaml:"port"`
	Auth     HttpAuthConfig `yaml:"auth"`
}

////

type SocksAuthConfig struct {
	Enabled bool              `yaml:"enabled"`
	Users   map[string]string `yaml:"users"`
}

type SocksConfig struct {
	Disabled bool            `yaml:"disabled"`
	Bind     string          `yaml:"bind"`
	Port     int             `yaml:"port"`
	Auth     SocksAuthConfig `yaml:"auth"`
}

////

type ResolverConfig struct {
	CacheSize int               `yaml:"cache_size"`
	CacheTTL  int               `yaml:"cache_ttl"`
	Hosts     map[string]string `yaml:"hosts"`
}
