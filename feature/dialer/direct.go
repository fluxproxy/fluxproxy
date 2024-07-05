package dialer

import (
	"context"
	"fmt"
	"github.com/fluxproxy/fluxproxy/net"
	stdnet "net"
	"time"
)

const (
	DIRECT = "DIRECT"
)

var (
	_ proxy.Dialer = (*TcpDirectDialer)(nil)
)

type TcpDirectDialer struct {
}

func NewTcpDirectDialer() *TcpDirectDialer {
	return &TcpDirectDialer{}
}

func (d *TcpDirectDialer) Name() string {
	return DIRECT
}

func (d *TcpDirectDialer) Dial(connCtx context.Context, remoteAddr net.Address) (proxy.Connection, error) {
	dialer := &stdnet.Dialer{
		Timeout:   time.Second * 5,
		KeepAlive: time.Duration(0),
	}
	conn, err := dialer.DialContext(connCtx, "tcp", remoteAddr.Addrport())
	if err != nil {
		return nil, fmt.Errorf("tcp dail. %w", err)
	}
	_ = (conn.(*stdnet.TCPConn)).SetKeepAlive(true)
	return proxy.NewDirectConnection(conn), nil
}
