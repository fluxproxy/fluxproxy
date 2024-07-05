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
	auth rocket.Authentication
	src  net.Address
	dest net.Address
	conn stdnet.Conn
	ctx  context.Context
	done context.CancelFunc
}

func NewConnStream(
	ctx context.Context, conn stdnet.Conn, dest net.Address,
	auth rocket.Authentication,
) *ConnStream {
	ctx, done := context.WithCancel(ctx)
	return &ConnStream{
		auth: auth,
		src:  auth.Source,
		dest: dest,
		conn: conn,
		ctx:  ctx,
		done: done,
	}
}

func (s *ConnStream) Connect(connector rocket.Connection) {
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

func (s *ConnStream) Source() net.Address {
	return s.auth.Source
}

func (s *ConnStream) Destination() net.Address {
	return s.dest
}

func (s *ConnStream) Authentication() rocket.Authentication {
	return s.auth
}
