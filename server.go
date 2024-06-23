package fluxway

import (
	"context"
	"fluxway/common"
	"fluxway/net"
	"fluxway/proxy"
	"fluxway/proxy/tcp"
	"fluxway/proxy/udp"
	"github.com/bytepowered/assert-go"
	"github.com/sirupsen/logrus"
)

type Server struct {
	tag        string
	listener   proxy.Listener
	router     proxy.Router
	connectors map[net.Network]proxy.Connector
}

func NewServer(tag string) *Server {
	assert.MustNotEmpty(tag, "server tag is required")
	return &Server{
		tag: tag,
	}
}

func (s *Server) Init() error {
	s.listener = tcp.NewTcpListener()
	s.connectors = map[net.Network]proxy.Connector{
		net.Network_TCP: tcp.NewTcpConnector(),
		net.Network_UDP: udp.NewUdpConnector(),
	}
	s.router = proxy.NewStaticRouter()
	assert.MustNotNil(s.listener, "server %s listener is required", s.tag)
	assert.MustNotNil(s.router, "server %s router is required", s.tag)
	assert.MustTrue(len(s.connectors) != 0, "server %s forwarder is required", s.tag)
	return s.listener.Init(proxy.ListenerOptions{
		Network: net.Network_TCP,
		Address: "0.0.0.0",
		Port:    9999,
	})
}

func (s *Server) Serve(servContext context.Context) error {
	return s.listener.Serve(servContext, func(ctx context.Context, conn net.Connection) {
		assert.MustTrue(ctxatu != servContext, "server context must be new")
		connID := common.NewID()
		ctx = proxy.ContextWithID(ctx, connID)
		ctx = proxy.ContextWithProxyType(ctx, s.listener.ProxyType())
		ctx = proxy.ContextWithConnection(ctx, &conn)
		logFields := logrus.Fields{
			"server":  s.tag,
			"network": s.listener.Network(),
			"source":  conn.Address,
			"id":      connID,
		}
		// Route
		routed, err := s.router.Route(ctx, &conn)
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
		connector, ok := s.connectors[routed.Destination.Network]
		if !ok {
			logrus.WithFields(logFields).Errorf("unsupported network-type: %s", routed.Destination.Network)
			return
		}
		if err := connector.DailServe(ctx, &routed); err != nil {
			logrus.WithFields(logFields).Errorf("connector error: %s", err)
		}
	})
}
