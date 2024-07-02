package stream

import (
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/internal"
	"github.com/rocketmanapp/rocket-proxy/net"
)

var (
	_ rocket.Listener = (*UdpListener)(nil)
)

type UdpListener struct {
	*internal.UdpListener
}

func NewUdpListener() *UdpListener {
	return &UdpListener{
		UdpListener: internal.NewUdpListener("udp", net.DefaultUdpOptions()),
	}
}
