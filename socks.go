package fluxway

import (
	"context"
	"fluxway/net"
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
	connector := tcp.NewTcpConnector()
	s.GenericServer.SetListener(listener)
	s.GenericServer.SetRouter(route.NewProxyRouter())
	s.SetConnectorSelector(func(conn *net.Connection) (proxy.Connector, bool) {
		return connector, true
	})
	serverOpts := s.GenericServer.Options()
	return listener.Init(proxy.ListenerOptions{
		Network: listener.Network(),
		Address: serverOpts.Bind,
		Port:    serverOpts.HttpPort,
	})
}
