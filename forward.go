package fluxway

import (
	"context"
	"fluxway/common"
	"fluxway/net"
	"fluxway/proxy"
	"fluxway/proxy/route"
	"fluxway/proxy/tcp"
	"fluxway/proxy/udp"
	"fmt"
	"github.com/bytepowered/assert-go"
	"github.com/sirupsen/logrus"
)

var (
	_ proxy.Server = (*ForwardServer)(nil)
)

type ForwardRootOptions struct {
	Rules []ForwardOptions `yaml:"rules"`
}

type ForwardOptions struct {
	Description string          `yaml:"description"`
	Network     string          `yaml:"network"`
	Port        int             `yaml:"port"`
	Disabled    bool            `yaml:"disabled"`
	Destination common.CAddress `yaml:"destination"`
}

type ForwardServer struct {
	options ForwardOptions
	*GenericServer
}

func NewForwardServer(serverOpts ServerOptions, forwardOpts ForwardOptions) *ForwardServer {
	if len(forwardOpts.Description) == 0 {
		forwardOpts.Description = fmt.Sprintf("forward-%d-to-%d", forwardOpts.Port, forwardOpts.Destination.Port)
	}
	return &ForwardServer{
		options:       forwardOpts,
		GenericServer: NewGenericServer(serverOpts),
	}
}

func (s *ForwardServer) Init(ctx context.Context) error {
	logrus.Infof("forward server init: %s:%s:%d, desc: %s", s.options.Network, s.Options().Bind, s.options.Port, s.options.Description)
	// 构建服务组件
	var listener proxy.Listener = nil
	var router proxy.Router = nil
	var connector proxy.Connector = nil
	network := net.ParseNetwork(s.options.Network)
	dest, err := ParseDestinationWith(network, s.options.Destination)
	if err != nil {
		return fmt.Errorf("invalid destination: %v, error: %w", s.options.Destination, err)
	}
	switch network {
	case net.Network_UDP:
		listener = udp.NewUdpListener()
		router = route.NewStaticRouter(dest)
		connector = udp.NewUdpConnector()
	case net.Network_TCP:
		listener = tcp.NewTcpListener()
		router = route.NewStaticRouter(dest)
		connector = tcp.NewTcpConnector()
	default:
		return fmt.Errorf("forward server unsupport network type: %s", s.options.Network)
	}
	s.SetListener(listener)
	s.SetRouter(router)
	s.SetConnector(connector)
	// 初始化
	assert.MustTrue(network == listener.Network(), "listener network error, was: %s", listener.Network())
	return listener.Init(proxy.ListenerOptions{
		Address: s.Options().Bind,
		Port:    s.options.Port,
	})
}
