package rocket

import (
	"context"
	"github.com/rocketmanapp/rocket-proxy/proxy"
	"github.com/rocketmanapp/rocket-proxy/proxy/route"
	"github.com/rocketmanapp/rocket-proxy/proxy/socket"
	"github.com/rocketmanapp/rocket-proxy/proxy/socks"
	"github.com/sirupsen/logrus"
)

var (
	_ proxy.Server = (*SocksServer)(nil)
)

type SocksOptions struct {
	Disabled bool `yaml:"disabled"`
}

type SocksServer struct {
	options SocksOptions
	*DirectServer
}

func NewSocksServer(serverOpts ServerOptions, socksOptions SocksOptions) *SocksServer {
	return &SocksServer{
		options:      socksOptions,
		DirectServer: NewGenericServer(serverOpts),
	}
}

func (s *SocksServer) Init(ctx context.Context) error {
	serverOpts := s.Options()
	listener := socks.NewSocksListener()
	router := route.NewProxyRouter()
	connector := socket.NewTcpConnector()
	s.SetServerType(proxy.ServerType_SOCKS5)
	s.SetListener(listener)
	s.SetRouter(router)
	s.SetResolver(NewDNSResolver())
	s.SetConnector(connector)
	return listener.Init(proxy.ListenerOptions{
		Address: serverOpts.Bind,
		Port:    serverOpts.SocksPort,
	})
}

func (s *SocksServer) Serve(ctx context.Context) error {
	defer logrus.Infof("socks: serve term")
	return s.DirectServer.Serve(ctx)
}
