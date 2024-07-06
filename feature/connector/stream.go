package connector

import (
	"context"
	"github.com/fluxproxy/fluxproxy"
	"github.com/fluxproxy/fluxproxy/helper"
	"github.com/fluxproxy/fluxproxy/net"
	"io"
	stdnet "net"
)

var (
	_ proxy.Connector = (*StreamConnector)(nil)
)

type StreamConnector struct {
	src        net.Address
	dest       net.Address
	conn       stdnet.Conn
	ctx        context.Context
	cancelFunc context.CancelFunc
}

func NewStreamConnector(
	ctx context.Context,
	conn stdnet.Conn,
	dest net.Address,
	src net.Address,
) *StreamConnector {
	ctx, cancel := context.WithCancel(ctx)
	return &StreamConnector{
		src:        src,
		dest:       dest,
		conn:       conn,
		ctx:        ctx,
		cancelFunc: cancel,
	}
}

func (s *StreamConnector) Connect(connection proxy.Connection) error {
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

func (s *StreamConnector) Close() error {
	s.cancelFunc()
	return s.conn.Close()
}

func (s *StreamConnector) HookFunc(key any) (proxy.HookFunc, bool) {
	v, ok := s.ctx.Value(key).(proxy.HookFunc)
	return v, ok
}

func (s *StreamConnector) Context() context.Context {
	return s.ctx
}

func (s *StreamConnector) Source() net.Address {
	return s.src
}

func (s *StreamConnector) Destination() net.Address {
	return s.dest
}
