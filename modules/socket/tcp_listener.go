package socket

import (
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/internal"
	"github.com/rocketmanapp/rocket-proxy/net"
)

var (
	_ rocket.Listener = (*TcpListener)(nil)
)

type TcpListener struct {
	*internal.TcpListener
}

func NewTcpListener() *TcpListener {
	return &TcpListener{
		TcpListener: internal.NewTcpListener("tcp", net.DefaultTcpOptions()),
	}
}
