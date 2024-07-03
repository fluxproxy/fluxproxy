package server

type ServerConfig struct {
	// Director
	Mode string `yaml:"mode"`
	Bind string `yaml:"bind"`
	// Http proxy server only
	HttpPort  int `yaml:"http_port"`
	HttpsPort int `yaml:"https_port"`
	// Socks proxy server only
	SocksPort int `yaml:"socks_port"`
}

////

type HttpsAuthConfig struct {
	Enabled bool              `yaml:"enabled"`
	Basic   map[string]string `yaml:"basic"`
}

type HttpsConfig struct {
	UseHttps bool `yaml:"-"`
	Disabled bool `yaml:"disabled"`
	// TLS
	TLSCertFile string `yaml:"tls_cert_file"`
	TLSKeyFile  string `yaml:"tls_key_file"`
	// Auth
	Auth HttpsAuthConfig `yaml:"auth"`
}

////

type ForwardConfig struct {
	Rules []ForwardRuleConfig `yaml:"rules"`
}

type ForwardRuleConfig struct {
	Description string   `yaml:"description"`
	Network     string   `yaml:"network"`
	Port        int      `yaml:"port"`
	Disabled    bool     `yaml:"disabled"`
	Destination CAddress `yaml:"destination"`
}

////

type SocksAuthConfig struct {
	Enabled bool              `yaml:"enabled"`
	Users   map[string]string `yaml:"users"`
}

type SocksConfig struct {
	Disabled bool            `yaml:"disabled"`
	Auth     SocksAuthConfig `yaml:"auth"`
}

////

type RulesetConfig struct {
	Type    string   `yaml:"type"`
	Origin  string   `yaml:"origin"`
	Access  string   `yaml:"access"`
	Address []string `yaml:"address"`
}
