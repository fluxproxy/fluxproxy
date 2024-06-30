package socket

import (
	"github.com/rocketmanapp/rocket-proxy/internal"
	"github.com/rocketmanapp/rocket-proxy/net"
	"github.com/rocketmanapp/rocket-proxy/proxy"
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
