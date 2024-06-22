package net

import "strings"

type Network int32

const (
	Network_Unknown Network = 0
	Network_TCP     Network = 1
	Network_UDP     Network = 2
	Network_UNIX    Network = 3
)

func (n Network) String() string {
	return n.SystemString()
}

func (n Network) Enum() *Network {
	p := new(Network)
	*p = n
	return p
}

func (n Network) SystemString() string {
	switch n {
	case Network_TCP:
		return "tcp"
	case Network_UDP:
		return "udp"
	case Network_UNIX:
		return "unix"
	default:
		return "unknown"
	}
}

func ParseNetwork(net string) Network {
	switch strings.ToLower(net) {
	case "tcp":
		return Network_TCP
	case "udp":
		return Network_UDP
	case "unix":
		return Network_UNIX
	default:
		return Network_Unknown
	}
}
