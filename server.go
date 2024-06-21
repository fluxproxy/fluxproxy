package vanity

import (
	"context"
	"github.com/sirupsen/logrus"
	"vanity/common"
	"vanity/net"
	"vanity/proxy"
)

type Server struct {
	listener Listener
	inbound  Inbound
	outbound Outbound
	router   Router
}

func NewServer(listener Listener) *Server {
	return &Server{
		listener: listener,
		inbound:  new(proxy.RawInbound),
		outbound: new(proxy.DirectOutbound),
	}
}

func (d *Server) Serve(servContext context.Context) error {
	return d.listener.Serve(servContext, func(ctx context.Context, conn net.Connection) {
		connID := common.NewID()
		ctx = contextWithID(ctx, connID)
		defer conn.Close()
		fields := logrus.Fields{
			"listener": d.listener.Tag(),
			"network":  d.listener.Network(),
			"source":   conn.Address,
			"id":       connID,
		}
		ctx = contextWithConnection(ctx, &conn)
		if err := d.inbound.Process(ctx, &conn); err != nil {
			logrus.WithFields(fields).Errorf("inbound error: %s", err)
			return
		}
		link, err := d.router.Router(ctx, &conn)
		if err != nil {
			logrus.WithFields(fields).Errorf("router error: %s", err)
			return
		}
		ctx = contextWithLink(ctx, &link)
		if err := d.outbound.DailServe(ctx, &link); err != nil {
			logrus.WithFields(fields).WithField("destination", link.Destination).Errorf("outbound error: %s", err)
			return
		}
		if link.KeepAlive {
			logrus.WithFields(fields).WithField("destination", link.Destination).Info("outbound terminaled")
		}
	})
}
