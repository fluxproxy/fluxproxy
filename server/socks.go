package server

import (
	"context"
	"fmt"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/modules/resolver"
	"github.com/rocketmanapp/rocket-proxy/modules/router"
	"github.com/rocketmanapp/rocket-proxy/modules/socks"
	"github.com/rocketmanapp/rocket-proxy/modules/stream"
	"github.com/sirupsen/logrus"
)

var (
	_ rocket.Server = (*SocksServer)(nil)
)

type SocksAuthConfig struct {
	Enabled bool              `yaml:"enabled"`
	Users   map[string]string `yaml:"users"`
}

type SocksConfig struct {
	Disabled bool            `yaml:"disabled"`
	Auth     SocksAuthConfig `yaml:"auth"`
}

type SocksServer struct {
	config SocksConfig
	*Director
}

func NewSocksServer(serverOpts Options, socksConfig SocksConfig) *SocksServer {
	return &SocksServer{
		config:   socksConfig,
		Director: NewDirector(serverOpts),
	}
}

func (s *SocksServer) Init(ctx context.Context) error {
	// 检查参数
	serverOpts := s.Options()
	if s.config.Auth.Enabled {
		if len(s.config.Auth.Users) == 0 {
			return fmt.Errorf("no users defined for socks auth")
		} else {
			logrus.Infof("socks: users auth enabled, users: %d", len(s.config.Auth.Users))
		}
	}
	// 构建服务组件
	socksListener := socks.NewSocksListener(socks.Options{
		AuthEnabled: s.config.Auth.Enabled,
	})
	proxyRouter := router.NewProxyRouter()
	connector := stream.NewTcpConnector()
	s.SetServerType(rocket.ServerTypeSOCKS)
	s.SetListener(socksListener)
	s.SetRouter(proxyRouter)
	s.SetResolver(resolver.NewResolverWith(ctx))
	s.SetAuthorizer(authorizer.WithBasicUsers(s.config.Auth.Enabled, s.config.Auth.Users).Authorize)
	s.SetConnector(connector)
	// 初始化
	return socksListener.Init(rocket.ListenerOptions{
		Address: serverOpts.Bind,
		Port:    serverOpts.SocksPort,
	})
}

func (s *SocksServer) Serve(ctx context.Context) error {
	defer logrus.Infof("socks: serve term")
	return s.Director.ServeListen(ctx)
}
