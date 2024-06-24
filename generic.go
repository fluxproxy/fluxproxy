package fluxway

import (
	"context"
	"fluxway/common"
	"fluxway/net"
	"fluxway/proxy"
	"fmt"
	"github.com/bytepowered/assert-go"
	"github.com/sirupsen/logrus"
)

type ServerOptions struct {
	// Generic
	Mode     string `yaml:"mode"`
	AllowLan bool   `yaml:"allow_lan"`
	Bind     string `yaml:"bind"`
	// Http proxy server only
	HttpPort int `yaml:"http_port"`
	// Socks proxy server only
	SocksPort int `yaml:"socks_port"`
}

type GenericServer struct {
	opts     ServerOptions
	listener proxy.Listener
	router   proxy.Router
	selector proxy.ConnectorSelector
}

func NewGenericServer(opts ServerOptions) *GenericServer {
	assert.MustNotEmpty(opts.Mode, "server mode is required")
	return &GenericServer{
		opts: opts,
	}
}

func (s *GenericServer) Options() ServerOptions {
	return s.opts
}

func (s *GenericServer) SetListener(listener proxy.Listener) {
	s.listener = listener
}

func (s *GenericServer) SetRouter(router proxy.Router) {
	s.router = router
}

func (s *GenericServer) SetConnectorSelector(f proxy.ConnectorSelector) {
	s.selector = f
}

func (s *GenericServer) Serve(servContext context.Context) error {
	assert.MustNotNil(s.listener, "server listener is nil")
	assert.MustNotNil(s.router, "server router is nil")
	assert.MustNotNil(s.selector, "server connector-selector is nil")
	return s.listener.Serve(servContext, func(connCtx context.Context, conn net.Connection) {
		assert.MustTrue(connCtx != servContext, "server context must be a new context")
		connID := common.NewID()
		connCtx = proxy.ContextWithID(connCtx, connID)
		connCtx = proxy.ContextWithProxyType(connCtx, s.listener.ProxyType())
		connCtx = proxy.ContextWithConnection(connCtx, &conn)
		logFields := logrus.Fields{
			"server":  s.opts.Mode,
			"network": s.listener.Network(),
			"source":  conn.Address,
			"id":      connID,
		}
		// Route
		routed, err := s.router.Route(connCtx, &conn)
		if err != nil {
			logrus.WithFields(logFields).Errorf("router error: %s", err)
			return
		}
		assert.MustNotNil(routed.ReadWriter, "routed.read-writer is nil")
		assert.MustTrue(routed.Destination.IsValid(), "routed.dest is invalid")
		if s.listener.Network() == net.Network_TCP {
			assert.MustNotNil(routed.TCPConn, "routed.TCPConn is nil")
		} else {
			assert.MustNil(routed.TCPConn, "routed.TCPConn is not nil")
		}
		logFields["destination"] = routed.Destination
		// Connect
		connector, ok := s.selector(&routed)
		if !ok {
			logrus.WithFields(logFields).Errorf("unsupported network-type: %s", routed.Destination.Network)
			return
		}
		if err := connector.DailServe(connCtx, &routed); err != nil {
			logrus.WithFields(logFields).Errorf("connector error: %s", err)
		}
	})
}

func ParseDestinationWith(network net.Network, addr common.AddressOptions) (net.Destination, error) {
	port, err := net.PortFromInt(uint32(addr.Port))
	if err != nil {
		return net.DestinationNotset, fmt.Errorf("invalid destination port: %d, error: %w", addr.Port, err)
	}
	return net.Destination{
		Network: network,
		Address: net.ParseAddress(addr.Address),
		Port:    port,
	}, nil
}
