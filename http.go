package fluxway

import (
	"context"
	"fluxway/internal"
	"fluxway/proxy"
	"fluxway/proxy/http"
	"fluxway/proxy/route"
	"fluxway/proxy/tcp"
)

var (
	_ proxy.Server = (*HttpServer)(nil)
)

type HttpOptions struct {
	Disabled bool `yaml:"disabled"`
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
	listener := http.NewHttpListener()
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
	} else {
		serverPort = serverOpts.HttpPort
	}
	return listener.Init(proxy.ListenerOptions{
		Address: serverOpts.Bind,
		Port:    serverPort,
	})
}
