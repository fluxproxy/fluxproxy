package vanity

import (
	"context"
	"github.com/bytepowered/assert-go"
	"github.com/sirupsen/logrus"
	"vanity/common"
	"vanity/net"
	"vanity/proxy"
	"vanity/proxy/tcp"
	"vanity/proxy/udp"
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
	s.listener = tcp.NewListener()
	s.connectors = map[net.Network]proxy.Connector{
		net.Network_TCP: tcp.NewConnector(),
		net.Network_UDP: udp.NewConnector(),
	}
	s.router = proxy.NewStaticRouter()
	assert.MustNotNil(s.listener, "server %s listener is required", s.tag)
	assert.MustNotNil(s.router, "server %s router is required", s.tag)
	assert.MustNotNil(len(s.connectors) != 0, "server %s forwarder is required", s.tag)
	return s.listener.Init(proxy.ListenerOptions{
		Network: net.Network_TCP,
		Address: "0.0.0.0",
		Port:    9999,
	})
}

func (s *Server) Serve(servContext context.Context) error {
	return s.listener.Serve(servContext, func(ctx context.Context, conn net.Connection) {
		connID := common.NewID()
		ctx = contextWithID(ctx, connID)
		fields := logrus.Fields{
			"server":  s.tag,
			"network": s.listener.Network(),
			"source":  conn.Address,
			"id":      connID,
		}
		ctx = contextWithConnection(ctx, &conn)
		// Route
		routed, err := s.router.Router(ctx, &conn)
		if err != nil {
			logrus.WithFields(fields).Errorf("router error: %s", err)
			return
		}
		assert.MustNotNil(routed.ReadWriteCloser, "routed.read write conn is nil")
		assert.MustTrue(routed.Destination.IsValid(), "routed.dest is invalid")
		if s.listener.Network() == net.Network_TCP {
			assert.MustNotNil(routed.TCPConn, "routed.TCPConn is nil")
		} else {
			assert.MustNil(routed.TCPConn, "routed.TCPConn is not nil")
		}
		fields["destination"] = routed.Destination
		// Connect
		connector, ok := s.connectors[routed.Destination.Network]
		if !ok {
			logrus.WithFields(fields).Errorf("unsupported network type error, %s: %s", routed.Destination.Network, err)
		}
		if err := connector.DailServe(ctx, &routed); err != nil {
			logrus.WithFields(fields).Errorf("connector error: %s", err)
			return
		}
	})
}
