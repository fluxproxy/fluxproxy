package tcp

import (
	"vanity/net"
	"vanity/proxy"
	"vanity/proxy/common"
)

var (
	_ proxy.Listener = (*Listener)(nil)
)

type Listener struct {
	*common.TcpListener
}

func NewListener() *Listener {
	return &Listener{
		TcpListener: common.NewTcpListener("tcp-listener", net.DefaultTcpOptions()),
	}
}
