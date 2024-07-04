package server

import (
	"context"
	"github.com/rocket-proxy/rocket-proxy"
	"github.com/rocket-proxy/rocket-proxy/net"
	stdnet "net"
)

var (
	_ rocket.Server = (*HttpStream)(nil)
)

type HttpStream struct {
	*Stream
}

func NewHttpStream(ctx context.Context, conn stdnet.Conn, addr net.Address) *HttpStream {
	return &HttpStream{
		Stream: NewStream(ctx, conn, addr),
	}
}
