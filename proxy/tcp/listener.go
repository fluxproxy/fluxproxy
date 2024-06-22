package tcp

import (
	"vanity/internal"
	"vanity/net"
	"vanity/proxy"
)

var (
	_ proxy.Listener = (*Listener)(nil)
)

type Listener struct {
	*internal.TcpListener
}

func NewTcpListener() *Listener {
	return &Listener{
		TcpListener: internal.NewTcpListener("tcp-listener", net.DefaultTcpOptions()),
	}
}
