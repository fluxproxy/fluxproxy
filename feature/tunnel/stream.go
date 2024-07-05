package tunnel

import (
	"context"
	"github.com/rocket-proxy/rocket-proxy"
	"github.com/rocket-proxy/rocket-proxy/helper"
	"github.com/rocket-proxy/rocket-proxy/net"
	"io"
	stdnet "net"
)

var (
	_ rocket.Tunnel = (*ConnStream)(nil)
)

type ConnStream struct {
	auth       rocket.Authentication
	src        net.Address
	dest       net.Address
	conn       stdnet.Conn
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewConnStream(
	ctx context.Context, conn stdnet.Conn, dest net.Address,
	auth rocket.Authentication,
) *ConnStream {
	ctx, cancel := context.WithCancel(ctx)
	return &ConnStream{
		auth:       auth,
		src:        auth.Source,
		dest:       dest,
		conn:       conn,
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

func (s *ConnStream) Connect(connector rocket.Connection) error {
	defer s.cancelFunc()
	ioErrors := make(chan error, 2)
	copier := func(name string, from io.Reader, to io.Writer) {
		ioErrors <- helper.Copier(from, to)
	}

	go copier("src-to-dest", s.conn, connector.Conn())
	go copier("dest-to-src", connector.Conn(), s.conn)

	select {
	case err := <-ioErrors:
		return err
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

func (s *ConnStream) Close() error {
	s.cancelFunc()
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
