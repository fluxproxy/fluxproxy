package server

import (
	"context"
	"fmt"
	"github.com/bytepowered/assert"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/modules/resolver"
	"github.com/rocketmanapp/rocket-proxy/modules/router"
	"github.com/rocketmanapp/rocket-proxy/modules/stream"
	"github.com/rocketmanapp/rocket-proxy/net"
	"github.com/sirupsen/logrus"
)

var (
	_ rocket.Server = (*ForwardServer)(nil)
)

type ForwardServer struct {
	ruleConfig ForwardRuleConfig
	*Director
}

func NewForwardServer(serverOpts ServerConfig, ruleConfig ForwardRuleConfig) *ForwardServer {
	if len(ruleConfig.Description) == 0 {
		ruleConfig.Description = fmt.Sprintf("forward-%d-to-%d", ruleConfig.Port, ruleConfig.Destination.Port)
	}
	return &ForwardServer{
		ruleConfig: ruleConfig,
		Director:   NewDirector(serverOpts),
	}
}

func (s *ForwardServer) Init(ctx context.Context) error {
	logrus.Infof("forward: init: %s:%s:%d, desc: %s", s.ruleConfig.Network, s.ServerConfig().Bind, s.ruleConfig.Port, s.ruleConfig.Description)
	// 构建服务组件
	var listener rocket.Listener = nil
	var proxyRouter rocket.Router = nil
	var connector rocket.Connector = nil
	network := net.ParseNetwork(s.ruleConfig.Network)
	dest, err := parseDestinationWith(network, s.ruleConfig.Destination)
	if err != nil {
		return fmt.Errorf("invalid destination: %v, error: %w", s.ruleConfig.Destination, err)
	}
	switch network {
	case net.NetworkUDP:
		listener = stream.NewUdpListener()
		proxyRouter = router.NewStaticRouter(dest)
		connector = stream.NewUdpConnector()
		s.SetServerType(rocket.ServerTypeUDP)
	case net.NetworkTCP:
		listener = stream.NewTcpListener()
		proxyRouter = router.NewStaticRouter(dest)
		connector = stream.NewTcpConnector()
		s.SetServerType(rocket.ServerTypeTCP)
	default:
		return fmt.Errorf("forward unsupport network: %s", s.ruleConfig.Network)
	}
	s.SetListener(listener)
	s.SetRouter(proxyRouter)
	s.SetResolver(resolver.NewResolverWith(ctx))
	s.SetConnector(connector)
	// 初始化
	assert.MustTrue(network == listener.Network(), "server network is not match listener, was: %s", listener.Network())
	return listener.Init(rocket.ListenerOptions{
		Address: s.ServerConfig().Bind,
		Port:    s.ruleConfig.Port,
	})
}

func (s *ForwardServer) Serve(ctx context.Context) error {
	defer logrus.Infof("forward: %s serve term", s.ruleConfig.Network)
	return s.Director.ServeListen(ctx)
}

func parseDestinationWith(network net.Network, addr CAddress) (net.Destination, error) {
	port, err := net.PortFromInt(uint32(addr.Port))
	if err != nil {
		return net.DestinationNotset, fmt.Errorf("invalid port: %d, error: %w", addr.Port, err)
	}
	return net.Destination{
		Network: network,
		Address: net.ParseAddress(addr.Address),
		Port:    port,
	}, nil
}
