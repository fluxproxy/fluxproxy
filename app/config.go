package app

const (
	configPathAuthenticator = "authenticator"
	configPathResolver      = "resolver"
	configPathServer        = "server"
	configPathServerHttp    = "server.http"
	configPathServerSocks   = "server.socks"
)

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
