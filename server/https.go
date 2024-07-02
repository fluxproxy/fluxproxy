package server

import (
	"context"
	"fmt"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/modules/auth"
	"github.com/rocketmanapp/rocket-proxy/modules/https"
	"github.com/rocketmanapp/rocket-proxy/modules/resolver"
	"github.com/rocketmanapp/rocket-proxy/modules/router"
	"github.com/rocketmanapp/rocket-proxy/modules/stream"
	"github.com/rocketmanapp/rocket-proxy/net"
	"github.com/sirupsen/logrus"
)

var (
	_ rocket.Server = (*HttpsServer)(nil)
)

type HttpsAuthConfig struct {
	Enabled bool              `yaml:"enabled"`
	Basic   map[string]string `yaml:"basic"`
}

type HttpsConfig struct {
	UseHttps bool `yaml:"-"`
	Disabled bool `yaml:"disabled"`
	// TLS
	TLSCertFile string `yaml:"tls_cert_file"`
	TLSKeyFile  string `yaml:"tls_key_file"`
	// Auth
	Auth HttpsAuthConfig `yaml:"auth"`
}

type HttpsServer struct {
	config HttpsConfig
	*Director
}

func NewHttpsServer(serverOpts Options, httpsConfig HttpsConfig) *HttpsServer {
	return &HttpsServer{
		config:   httpsConfig,
		Director: NewDirector(serverOpts),
	}
}

func (s *HttpsServer) Init(ctx context.Context) error {
	// 检查参数
	serverOpts := s.Options()
	if s.config.Auth.Enabled {
		if len(s.config.Auth.Basic) == 0 {
			return fmt.Errorf("no users defined for https auth")
		} else {
			logrus.Infof("https: basic auth enabled, users: %d", len(s.config.Auth.Basic))
		}
	}
	var serverPort int
	if s.config.UseHttps {
		serverPort = serverOpts.HttpsPort
		if len(s.config.TLSCertFile) < 3 {
			return fmt.Errorf("https.tls_cert_file is required in config")
		}
		if len(s.config.TLSKeyFile) < 3 {
			return fmt.Errorf("https.tls_key_file is required in config")
		}
	} else {
		serverPort = serverOpts.HttpPort
	}
	// 构建服务组件
	httpListener := https.NewHttpsListener(https.Options{
		UseHttps: s.config.UseHttps,
	})
	proxyRouter := router.NewProxyRouter()
	tcpConnector := stream.NewTcpConnector()
	hstrConnector := https.NewHrtpConnector()
	s.SetServerType(rocket.ServerTypeHTTPS)
	s.SetListener(httpListener)
	s.SetRouter(proxyRouter)
	s.SetResolver(resolver.NewResolverWith(ctx))
	s.SetAuthorizer(auth.WithBasicUsers(s.config.Auth.Enabled, s.config.Auth.Basic).Authorize)
	s.SetConnectorSelector(func(conn *net.Connection) (rocket.Connector, bool) {
		switch conn.Destination.Network {
		case net.Network_TCP:
			return tcpConnector, true
		case net.Network_HRTP:
			return hstrConnector, true
		default:
			return nil, false
		}
	})
	// 初始化
	return httpListener.Init(rocket.ListenerOptions{
		Address: serverOpts.Bind,
		Port:    serverPort,
		// TLS
		TLSCertFile: s.config.TLSCertFile,
		TLSKeyFile:  s.config.TLSKeyFile,
	})
}

func (s *HttpsServer) Serve(ctx context.Context) error {
	defer logrus.Infof("https: serve term")
	return s.Director.ServeListen(ctx)
}
