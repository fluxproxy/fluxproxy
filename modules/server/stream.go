package server

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
	_ rocket.Server = (*Stream)(nil)
)

type Stream struct {
	addr net.Address
	conn stdnet.Conn
	ctx  context.Context
	done context.CancelFunc
}

func NewStream(ctx context.Context, conn stdnet.Conn, addr net.Address) *Stream {
	return &Stream{
		ctx:  ctx,
		addr: addr,
		conn: conn,
	}
}

func (s *Stream) Address() net.Address {
	return s.addr
}

func (s *Stream) Connect(connector rocket.Connector) {
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

func (s *Stream) Close() error {
	s.done()
	return s.conn.Close()
}

func (s *Stream) Done() <-chan struct{} {
	return s.ctx.Done()
}
