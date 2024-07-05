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
	_ rocket.Tunnel = (*ConnStreamTunnel)(nil)
)

type ConnStreamTunnel struct {
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
) *ConnStreamTunnel {
	ctx, cancel := context.WithCancel(ctx)
	return &ConnStreamTunnel{
		auth:       auth,
		src:        auth.Source,
		dest:       dest,
		conn:       conn,
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

func (s *ConnStreamTunnel) Connect(connection rocket.Connection) error {
	defer s.cancelFunc()
	ioErrors := make(chan error, 2)
	copier := func(name string, from io.Reader, to io.Writer) {
		ioErrors <- helper.Copier(from, to)
	}

	go copier("src-to-dest", s.conn, connection.Conn())
	go copier("dest-to-src", connection.Conn(), s.conn)

	select {
	case err := <-ioErrors:
		return err
	case <-s.ctx.Done():
		return s.ctx.Err()
	}
}

func (s *ConnStreamTunnel) Close() error {
	s.cancelFunc()
	return s.conn.Close()
}

func (s *ConnStreamTunnel) Context() context.Context {
	return s.ctx
}

func (s *ConnStreamTunnel) Source() net.Address {
	return s.auth.Source
}

func (s *ConnStreamTunnel) Destination() net.Address {
	return s.dest
}

func (s *ConnStreamTunnel) Authentication() rocket.Authentication {
	return s.auth
}
