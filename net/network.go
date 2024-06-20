package net

import "strings"

type Network int32

const (
	Network_Unknown Network = 0
	Network_TCP     Network = 1
	Network_UDP     Network = 2
	Network_UNIX    Network = 3
)

var (
	Network_name = map[int32]string{
		0: "Unknown",
		1: "TCP",
		2: "UDP",
		3: "UNIX",
	}
	Network_value = map[string]int32{
		"Unknown": 0,
		"TCP":     1,
		"UDP":     2,
		"UNIX":    3,
	}
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
