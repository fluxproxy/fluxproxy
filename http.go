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
	options HttpOptions
	*GenericServer
}

func NewHttpServer(serverOpts ServerOptions, httpOptions HttpOptions) *HttpServer {
	return &HttpServer{
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
	return listener.Init(proxy.ListenerOptions{
		Address: serverOpts.Bind,
		Port:    serverOpts.HttpPort,
	})
}
