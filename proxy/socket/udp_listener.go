package socket

import (
	"fluxway/internal"
	"fluxway/net"
	"fluxway/proxy"
)

var (
	_ proxy.Listener = (*UdpListener)(nil)
)

type UdpListener struct {
	*internal.UdpListener
}

func NewUdpListener() *UdpListener {
	return &UdpListener{
		UdpListener: internal.NewUdpListener("udp", net.DefaultUdpOptions()),
	}
}
