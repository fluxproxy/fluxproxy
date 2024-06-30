package socket

import (
	"rocket/internal"
	"rocket/net"
	"rocket/proxy"
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
