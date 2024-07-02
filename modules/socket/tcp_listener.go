package socket

import (
	"github.com/rocketmanapp/rocket-proxy/internal"
	"github.com/rocketmanapp/rocket-proxy/net"
	"github.com/rocketmanapp/rocket-proxy/proxy"
)

var (
	_ proxy.Listener = (*TcpListener)(nil)
)

type TcpListener struct {
	*internal.TcpListener
}

func NewTcpListener() *TcpListener {
	return &TcpListener{
		TcpListener: internal.NewTcpListener("tcp", net.DefaultTcpOptions()),
	}
}
