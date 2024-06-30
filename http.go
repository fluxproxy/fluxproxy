package rocket

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"rocket/net"
	"rocket/proxy"
	"rocket/proxy/http"
	"rocket/proxy/route"
	"rocket/proxy/socket"
)

var (
	_ proxy.Server = (*HttpServer)(nil)
)

type HttpOptions struct {
	Disabled bool `yaml:"disabled"`
	// TLS
	TLSCertFile string `yaml:"tls_cert_file"`
	TLSKeyFile  string `yaml:"tls_key_file"`
}

type HttpServer struct {
	isHttps bool
	options HttpOptions
	*DirectServer
}

func NewHttpServer(serverOpts ServerOptions, httpOptions HttpOptions, isHttps bool) *HttpServer {
	return &HttpServer{
		isHttps:      isHttps,
		options:      httpOptions,
		DirectServer: NewGenericServer(serverOpts),
	}
}

func (s *HttpServer) Init(ctx context.Context) error {
	listener := http.NewHttpListener(s.isHttps)
	router := route.NewProxyRouter()
	tcpConnector := socket.NewTcpConnector()
	hstrConnector := http.NewHrtpConnector()
	s.SetServerType(proxy.ServerType_HTTPS)
	s.SetListener(listener)
	s.SetRouter(router)
	s.SetResolver(NewDNSResolver())
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
			return fmt.Errorf("http.tls_cert_file is required in config")
		}
		if len(s.options.TLSKeyFile) < 3 {
			return fmt.Errorf("http.tls_key_file is required in config")
		}
	} else {
		serverPort = serverOpts.HttpPort
	}
	return listener.Init(proxy.ListenerOptions{
		Address: serverOpts.Bind,
		Port:    serverPort,
		// TLS
		TLSCertFile: s.options.TLSCertFile,
		TLSKeyFile:  s.options.TLSKeyFile,
	})
}

func (s *HttpServer) Serve(ctx context.Context) error {
	defer logrus.Infof("http: serve term")
	return s.DirectServer.Serve(ctx)
}
