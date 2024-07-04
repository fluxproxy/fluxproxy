package tunnel

import (
	"context"
	"github.com/rocket-proxy/rocket-proxy"
	"github.com/rocket-proxy/rocket-proxy/helper"
	"github.com/rocket-proxy/rocket-proxy/net"
	"github.com/sirupsen/logrus"
	"io"
	stdnet "net"
)

var (
	_ rocket.Tunnel = (*ConnStream)(nil)
)

type ConnStream struct {
	hook rocket.TunnelHook
	auth rocket.Authentication
	dest net.Address
	conn stdnet.Conn
	ctx  context.Context
	done context.CancelFunc
}

func NewConnStream(
	ctx context.Context, conn stdnet.Conn, dest net.Address,
	auth rocket.Authentication,
	hooks rocket.TunnelHook,
) *ConnStream {
	return &ConnStream{
		auth: auth,
		ctx:  ctx,
		dest: dest,
		conn: conn,
		hook: hooks,
	}
}

func (s *ConnStream) Connect(connector rocket.Connection) {
	s.ctx, s.done = context.WithCancel(s.ctx)
	defer s.done()
	ioErrors := make(chan error, 2)
	copier := func(name string, from io.Reader, to io.Writer) {
		ioErrors <- helper.Copier(from, to)
	}
	go copier("src-to-dest", s.conn, connector.Conn())
	go copier("dest-to-src", connector.Conn(), s.conn)
	err := <-ioErrors
	if err != nil {
		logrus.Errorf("failed to copy stream: %s", err)
	}
}

func (s *ConnStream) Close() error {
	s.done()
	return s.conn.Close()
}

func (s *ConnStream) Context() context.Context {
	return s.ctx
}

func (s *ConnStream) Destination() net.Address {
	return s.dest
}

func (s *ConnStream) Authentication() rocket.Authentication {
	return s.auth
}

func (s *ConnStream) Hook() rocket.TunnelHook {
	return s.hook
}
