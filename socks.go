package fluxway

import (
	"context"
	"fluxway/proxy"
	"fluxway/proxy/route"
	"fluxway/proxy/socks"
	"fluxway/proxy/tcp"
)

var (
	_ proxy.Server = (*SocksProxyServer)(nil)
)

type SocksProxyOptions struct {
	Disabled bool `yaml:"disabled"`
}

type SocksProxyServer struct {
	options SocksProxyOptions
	*GenericServer
}

func NewSocksProxyServer(serverOpts ServerOptions, socksProxyOptions SocksProxyOptions) *SocksProxyServer {
	return &SocksProxyServer{
		options:       socksProxyOptions,
		GenericServer: NewGenericServer(serverOpts),
	}
}

func (s *SocksProxyServer) Init(ctx context.Context) error {
	listener := socks.NewSocksListener()
	router := route.NewProxyRouter()
	connector := tcp.NewTcpConnector()
	s.SetListener(listener)
	s.SetRouter(router)
	s.SetConnector(connector)
	serverOpts := s.GenericServer.Options()
	return listener.Init(proxy.ListenerOptions{
		Address: serverOpts.Bind,
		Port:    serverOpts.HttpPort,
	})
}
