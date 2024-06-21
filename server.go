package vanity

import (
	"context"
	"github.com/bytepowered/assert-go"
	"github.com/sirupsen/logrus"
	"vanity/common"
	"vanity/net"
	"vanity/proxy"
	"vanity/proxy/tcp"
)

type Server struct {
	tag       string
	listener  proxy.Listener
	forwarder proxy.Forwarder
	router    proxy.Router
}

func NewServer(tag string) *Server {
	assert.MustNotEmpty(tag, "server tag is required")
	return &Server{
		tag: tag,
	}
}

func (s *Server) Init() error {
	s.listener = tcp.NewListener()
	s.forwarder = tcp.NewForwarder()
	s.router = proxy.NewStaticRouter()
	assert.MustNotNil(s.listener, "server %s listener is required", s.tag)
	assert.MustNotNil(s.forwarder, "server %s forwarder is required", s.tag)
	assert.MustNotNil(s.router, "server %s router is required", s.tag)
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
		defer conn.Close()
		fields := logrus.Fields{
			"server":  s.tag,
			"network": s.listener.Network(),
			"source":  conn.Address,
			"id":      connID,
		}
		ctx = contextWithConnection(ctx, &conn)
		link, err := s.router.Router(ctx, &conn)
		if err != nil {
			logrus.WithFields(fields).Errorf("router error: %s", err)
			return
		}
		ctx = contextWithLink(ctx, &link)
		if err := s.forwarder.DailServe(ctx, &link); err != nil {
			logrus.WithFields(fields).WithField("destination", link.Destination).Errorf("forwarder error: %s", err)
			return
		}
		if link.KeepAlive {
			logrus.WithFields(fields).WithField("destination", link.Destination).Info("forwarder stop")
		}
	})
}
