package dialer

import (
	"context"
	"fmt"
	"github.com/rocket-proxy/rocket-proxy"
	"github.com/rocket-proxy/rocket-proxy/net"
	stdnet "net"
	"time"
)

const (
	DIRECT = "DIRECT"
)

var (
	_ rocket.Dialer = (*TcpDirectDialer)(nil)
)

type TcpDirectDialer struct {
}

func NewTcpDirectDialer() *TcpDirectDialer {
	return &TcpDirectDialer{}
}

func (d *TcpDirectDialer) Name() string {
	return DIRECT
}

func (d *TcpDirectDialer) Dial(connCtx context.Context, remoteAddr net.Address) (rocket.Connection, error) {
	dialer := &stdnet.Dialer{
		Timeout:   time.Second * 5,
		KeepAlive: time.Duration(0),
	}
	conn, err := dialer.DialContext(connCtx, "tcp", remoteAddr.Addrport())
	if err != nil {
		return nil, fmt.Errorf("tcp dail. %w", err)
	}
	_ = (conn.(*stdnet.TCPConn)).SetKeepAlive(true)
	return rocket.NewDirectConnection(conn), nil
}
