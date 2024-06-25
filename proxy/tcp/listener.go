package tcp

import (
	"fluxway/internal"
	"fluxway/net"
	"fluxway/proxy"
)

var (
	_ proxy.Listener = (*Listener)(nil)
)

type Listener struct {
	*internal.TcpListener
}

func NewTcpListener() *Listener {
	return &Listener{
		TcpListener: internal.NewTcpListener("tcp", net.DefaultTcpOptions()),
	}
}
