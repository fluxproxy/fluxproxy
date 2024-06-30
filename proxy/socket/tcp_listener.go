package socket

import (
	"rocket/internal"
	"rocket/net"
	"rocket/proxy"
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
