package server

import (
	"context"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/modules/resolver"
	"github.com/rocketmanapp/rocket-proxy/modules/router"
	"github.com/rocketmanapp/rocket-proxy/modules/socket"
	"github.com/rocketmanapp/rocket-proxy/modules/socks"
	"github.com/sirupsen/logrus"
)

var (
	_ rocket.Server = (*SocksServer)(nil)
)

type SocksOptions struct {
	Disabled bool `yaml:"disabled"`
}

type SocksServer struct {
	options SocksOptions
	*Director
}

func NewSocksServer(serverOpts Options, socksOptions SocksOptions) *SocksServer {
	return &SocksServer{
		options:  socksOptions,
		Director: NewDirector(serverOpts),
	}
}

func (s *SocksServer) Init(ctx context.Context) error {
	serverOpts := s.Options()
	socksListener := socks.NewSocksListener()
	proxyRouter := router.NewProxyRouter()
	connector := socket.NewTcpConnector()
	s.SetServerType(rocket.ServerTypeSOCKS)
	s.SetListener(socksListener)
	s.SetRouter(proxyRouter)
	s.SetResolver(resolver.NewDNSResolverWith(ctx))
	s.SetConnector(connector)
	return socksListener.Init(rocket.ListenerOptions{
		Address: serverOpts.Bind,
		Port:    serverOpts.SocksPort,
	})
}

func (s *SocksServer) Serve(ctx context.Context) error {
	defer logrus.Infof("socks: serve term")
	return s.Director.ServeListen(ctx)
}
