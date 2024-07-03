package net

import "net"

var (
	DestinationNotset = Destination{Network: NetworkUnknown}
)

type Destination struct {
	Address Address
	Port    Port
	Network Network
}

func (d Destination) NetAddr() string {
	addr := ""
	if d.Network == NetworkTCP || d.Network == NetworkUDP {
		addr = net.JoinHostPort(d.Address.String(), d.Port.String())
	}
	return addr
}

func (d Destination) String() string {
	return d.Network.String() + "//" + d.NetAddr()
}

func (d Destination) IsValid() bool {
	return d.Network != NetworkUnknown
}
