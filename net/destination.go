package net

import "net"

var (
	DestinationNotset = Destination{Network: Network_Unknown}
)

type Destination struct {
	Address Address
	Port    Port
	Network Network
}

func (d Destination) NetAddr() string {
	addr := ""
	if d.Network == Network_TCP || d.Network == Network_UDP {
		addr = net.JoinHostPort(d.Address.String(), d.Port.String())
	}
	return addr
}

func (d Destination) String() string {
	prefix := "unknown:"
	switch d.Network {
	case Network_TCP:
		prefix = "tcp:"
	case Network_UDP:
		prefix = "udp:"
	}
	return prefix + "//" + d.NetAddr()
}

func (d Destination) IsValid() bool {
	return d.Network != Network_Unknown
}
