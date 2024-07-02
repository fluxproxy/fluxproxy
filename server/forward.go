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

type ForwardRootOptions struct {
	Rules []ForwardOptions `yaml:"rules"`
}

type ForwardOptions struct {
	Description string   `yaml:"description"`
	Network     string   `yaml:"network"`
	Port        int      `yaml:"port"`
	Disabled    bool     `yaml:"disabled"`
	Destination CAddress `yaml:"destination"`
}

type ForwardServer struct {
	options ForwardOptions
	*Director
}

func NewForwardServer(serverOpts Options, forwardOpts ForwardOptions) *ForwardServer {
	if len(forwardOpts.Description) == 0 {
		forwardOpts.Description = fmt.Sprintf("forward-%d-to-%d", forwardOpts.Port, forwardOpts.Destination.Port)
	}
	return &ForwardServer{
		options:  forwardOpts,
		Director: NewDirector(serverOpts),
	}
}

func (s *ForwardServer) Init(ctx context.Context) error {
	logrus.Infof("forward: init: %s:%s:%d, desc: %s", s.options.Network, s.Options().Bind, s.options.Port, s.options.Description)
	// 构建服务组件
	var listener rocket.Listener = nil
	var proxyRouter rocket.Router = nil
	var connector rocket.Connector = nil
	network := net.ParseNetwork(s.options.Network)
	dest, err := parseDestinationWith(network, s.options.Destination)
	if err != nil {
		return fmt.Errorf("invalid destination: %v, error: %w", s.options.Destination, err)
	}
	switch network {
	case net.Network_UDP:
		listener = stream.NewUdpListener()
		proxyRouter = router.NewStaticRouter(dest)
		connector = stream.NewUdpConnector()
		s.SetServerType(rocket.ServerTypeUDP)
	case net.Network_TCP:
		listener = stream.NewTcpListener()
		proxyRouter = router.NewStaticRouter(dest)
		connector = stream.NewTcpConnector()
		s.SetServerType(rocket.ServerTypeTCP)
	default:
		return fmt.Errorf("forward unsupport network: %s", s.options.Network)
	}
	s.SetListener(listener)
	s.SetRouter(proxyRouter)
	s.SetResolver(resolver.NewDNSResolverWith(ctx))
	s.SetConnector(connector)
	// 初始化
	assert.MustTrue(network == listener.Network(), "server network is not match listener, was: %s", listener.Network())
	return listener.Init(rocket.ListenerOptions{
		Address: s.Options().Bind,
		Port:    s.options.Port,
	})
}

func (s *ForwardServer) Serve(ctx context.Context) error {
	defer logrus.Infof("forward: %s serve term", s.options.Network)
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
