package server

import (
	"context"
	"fmt"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/modules/resolver"
	"github.com/rocketmanapp/rocket-proxy/modules/router"
	"github.com/rocketmanapp/rocket-proxy/modules/socks"
	"github.com/rocketmanapp/rocket-proxy/modules/stream"
	"github.com/rocketmanapp/rocket-proxy/net"
	"github.com/sirupsen/logrus"
)

var (
	_ rocket.Server = (*SocksServer)(nil)
)

type SocksAuthOptions struct {
	Enabled bool              `yaml:"enabled"`
	Users   map[string]string `yaml:"users"`
}

type SocksOptions struct {
	Disabled bool             `yaml:"disabled"`
	Auth     SocksAuthOptions `yaml:"auth"`
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
	// components
	socksListener := socks.NewSocksListener(socks.Options{
		AuthEnabled: s.options.Auth.Enabled,
	})
	proxyRouter := router.NewProxyRouter()
	connector := stream.NewTcpConnector()
	s.SetServerType(rocket.ServerTypeSOCKS)
	s.SetListener(socksListener)
	s.SetRouter(proxyRouter)
	s.SetResolver(resolver.NewResolverWith(ctx))
	s.SetConnector(connector)
	// setup
	s.SetAuthorizer(s.doUserAuth)
	if s.options.Auth.Enabled {
		if len(s.options.Auth.Users) == 0 {
			return fmt.Errorf("no users defined for socks auth")
		} else {
			logrus.Infof("socks: auth enabled, users: %d", len(s.options.Auth.Users))
		}
	}
	// init
	return socksListener.Init(rocket.ListenerOptions{
		Address: serverOpts.Bind,
		Port:    serverOpts.SocksPort,
	})
}

func (s *SocksServer) doUserAuth(ctx context.Context, conn net.Connection, auth rocket.ListenerAuthorization) error {
	pass, ok := s.options.Auth.Users[auth.Username]
	if !ok {
		return fmt.Errorf("invalid user: %s", auth.Username)
	}
	if pass != auth.Password {
		return fmt.Errorf("invalid pass for user: %s", auth.Username)
	} else {
		return nil
	}
}

func (s *SocksServer) noAuth(ctx context.Context, conn net.Connection, auth rocket.ListenerAuthorization) error {
	return nil
}

func (s *SocksServer) Serve(ctx context.Context) error {
	defer logrus.Infof("socks: serve term")
	return s.Director.ServeListen(ctx)
}
