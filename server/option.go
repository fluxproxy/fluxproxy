package server

type Options struct {
	// Director
	Mode string `yaml:"mode"`
	Bind string `yaml:"bind"`
	// Http proxy server only
	HttpPort  int `yaml:"http_port"`
	HttpsPort int `yaml:"https_port"`
	// Socks proxy server only
	SocksPort int `yaml:"socks_port"`
}
