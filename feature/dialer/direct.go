package dialer

import (
	"context"
	"github.com/rocket-proxy/rocket-proxy"
	"github.com/rocket-proxy/rocket-proxy/net"
	stdnet "net"
)

const (
	DIRECT = "DIRECT"
)

var (
	_ rocket.Dialer = (*Direct)(nil)
)

type Direct struct {
}

func NewDirect() *Direct {
	return &Direct{}
}

func (d *Direct) Name() string {
	return DIRECT
}

func (d *Direct) Dial(srcConnCtx context.Context, remoteAddr net.Address) (rocket.Connection, error) {
	conn, err := stdnet.Dial("tcp", remoteAddr.Addrport())
	if err != nil {
		return nil, err
	}
	_ = (conn.(*stdnet.TCPConn)).SetKeepAlive(true)
	return rocket.NewDirectConnection(conn), nil
}
