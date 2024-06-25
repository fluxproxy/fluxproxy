package fluxway

import (
	"context"
	"fluxway/internal"
	"fluxway/proxy"
	"fluxway/proxy/route"
	"fluxway/proxy/socks"
	"fluxway/proxy/tcp"
)

var (
	_ proxy.Server = (*SocksServer)(nil)
)

type SocksOptions struct {
	Disabled bool `yaml:"disabled"`
}

type SocksServer struct {
	options SocksOptions
	*GenericServer
}

func NewSocksServer(serverOpts ServerOptions, socksOptions SocksOptions) *SocksServer {
	return &SocksServer{
		options:       socksOptions,
		GenericServer: NewGenericServer(serverOpts),
	}
}

func (s *SocksServer) Init(ctx context.Context) error {
	serverOpts := s.Options()
	listener := socks.NewSocksListener()
	router := route.NewProxyRouter()
	connector := tcp.NewTcpConnector()
	s.SetListener(listener)
	s.SetRouter(router)
	s.SetResolver(internal.NewDNSResolver())
	s.SetConnector(connector)
	return listener.Init(proxy.ListenerOptions{
		Address: serverOpts.Bind,
		Port:    serverOpts.SocksPort,
	})
}
