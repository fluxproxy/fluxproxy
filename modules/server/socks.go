package server

import (
	"context"
	"github.com/rocket-proxy/rocket-proxy"
	"github.com/rocket-proxy/rocket-proxy/net"
	stdnet "net"
)

var (
	_ rocket.Server = (*SocksStream)(nil)
)

type SocksStream struct {
	*Stream
}

func NewSocksStream(ctx context.Context, conn stdnet.Conn, addr net.Address) *SocksStream {
	return &SocksStream{
		Stream: NewStream(ctx, conn, addr),
	}
}
