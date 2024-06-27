package fluxway

import (
	"context"
	"fluxway/internal"
	"fluxway/proxy"
	"fluxway/proxy/http"
	"fluxway/proxy/route"
	"fluxway/proxy/tcp"
	"github.com/bytepowered/assert-go"
)

var (
	_ proxy.Server = (*HttpServer)(nil)
)

type HttpOptions struct {
	Disabled bool `yaml:"disabled"`
	// TLS
	TLSCertFile string `yaml:"tls_cert_file"`
	TLSKeyFile  string `yaml:"tls_key_file"`
}

type HttpServer struct {
	isHttps bool
	options HttpOptions
	*GenericServer
}

func NewHttpServer(serverOpts ServerOptions, httpOptions HttpOptions, isHttps bool) *HttpServer {
	return &HttpServer{
		isHttps:       isHttps,
		options:       httpOptions,
		GenericServer: NewGenericServer(serverOpts),
	}
}

func (s *HttpServer) Init(ctx context.Context) error {
	serverOpts := s.Options()
	listener := http.NewHttpListener(s.isHttps)
	router := route.NewProxyRouter()
	connector := tcp.NewTcpConnector()
	s.SetListener(listener)
	s.SetRouter(router)
	s.SetResolver(internal.NewDNSResolver())
	s.SetConnector(connector)
	// Listener init
	var serverPort int
	if s.isHttps {
		serverPort = serverOpts.HttpsPort
		assert.MustNotEmpty(s.options.TLSCertFile, "TLSCertFile is required in https server")
		assert.MustNotEmpty(s.options.TLSKeyFile, "TLSKeyFile is required in https server")
	} else {
		serverPort = serverOpts.HttpPort
	}
	return listener.Init(proxy.ListenerOptions{
		Address: serverOpts.Bind,
		Port:    serverPort,
		// TLS
		TLSCertFile: s.options.TLSCertFile,
		TLSKeyFile:  s.options.TLSKeyFile,
	})
}
