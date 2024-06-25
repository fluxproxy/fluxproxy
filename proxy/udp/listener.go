package udp

import (
	"fluxway/internal"
	"fluxway/net"
	"fluxway/proxy"
)

var (
	_ proxy.Listener = (*Listener)(nil)
)

type Listener struct {
	*internal.UdpListener
}

func NewUdpListener() *Listener {
	return &Listener{
		UdpListener: internal.NewUdpListener("udp", net.DefaultUdpOptions()),
	}
}
