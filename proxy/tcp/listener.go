package tcp

import (
	"vanity/net"
	"vanity/proxy"
	"vanity/proxy/internal"
)

var (
	_ proxy.Listener = (*Listener)(nil)
)

type Listener struct {
	*internal.TcpListener
}

func NewListener() *Listener {
	return &Listener{
		TcpListener: internal.NewTcpListener("tcp-listener", net.DefaultTcpOptions()),
	}
}
