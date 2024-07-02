package server

import (
	"context"
	"fmt"
	"github.com/rocketmanapp/rocket-proxy/modules/https"
	"github.com/rocketmanapp/rocket-proxy/modules/resolver"
	"github.com/rocketmanapp/rocket-proxy/modules/router"
	"github.com/rocketmanapp/rocket-proxy/modules/socket"
	"github.com/rocketmanapp/rocket-proxy/net"
	"github.com/rocketmanapp/rocket-proxy/proxy"
	"github.com/sirupsen/logrus"
)

var (
	_ proxy.Server = (*HttpsServer)(nil)
)

type HttpsOptions struct {
	Disabled bool `yaml:"disabled"`
	// TLS
	TLSCertFile string `yaml:"tls_cert_file"`
	TLSKeyFile  string `yaml:"tls_key_file"`
}

type HttpsServer struct {
	isHttps bool
	options HttpsOptions
	*Director
}

func NewHttpsServer(serverOpts Options, httpOptions HttpsOptions, isHttps bool) *HttpsServer {
	return &HttpsServer{
		isHttps:  isHttps,
		options:  httpOptions,
		Director: NewDirector(serverOpts),
	}
}

func (s *HttpsServer) Init(ctx context.Context) error {
	httpListener := https.NewHttpListener(s.isHttps)
	proxyRouter := router.NewProxyRouter()
	tcpConnector := socket.NewTcpConnector()
	hstrConnector := https.NewHrtpConnector()
	s.SetServerType(proxy.ServerTypeHTTPS)
	s.SetListener(httpListener)
	s.SetRouter(proxyRouter)
	s.SetResolver(resolver.NewDNSResolverWith(ctx))
	s.SetConnectorSelector(func(conn *net.Connection) (proxy.Connector, bool) {
		switch conn.Destination.Network {
		case net.Network_TCP:
			return tcpConnector, true
		case net.Network_HRTP:
			return hstrConnector, true
		default:
			return nil, false
		}
	})
	// Listener init
	serverOpts := s.Options()
	var serverPort int
	if s.isHttps {
		serverPort = serverOpts.HttpsPort
		if len(s.options.TLSCertFile) < 3 {
			return fmt.Errorf("https.tls_cert_file is required in config")
		}
		if len(s.options.TLSKeyFile) < 3 {
			return fmt.Errorf("https.tls_key_file is required in config")
		}
	} else {
		serverPort = serverOpts.HttpPort
	}
	return httpListener.Init(proxy.ListenerOptions{
		Address: serverOpts.Bind,
		Port:    serverPort,
		// TLS
		TLSCertFile: s.options.TLSCertFile,
		TLSKeyFile:  s.options.TLSKeyFile,
	})
}

func (s *HttpsServer) Serve(ctx context.Context) error {
	defer logrus.Infof("https: serve term")
	return s.Director.ServeListen(ctx)
}
