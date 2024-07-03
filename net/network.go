package net

import "strings"

type Network int32

const (
	NetworkUnknown Network = 0
	NetworkTCP     Network = 1
	NetworkUDP     Network = 2
	NetworkHRTP    Network = 4 // Http round trip
)

func (n Network) String() string {
	switch n {
	case NetworkTCP:
		return "tcp"
	case NetworkUDP:
		return "udp"
	case NetworkHRTP:
		return "hrtp"
	default:
		return "unknown"
	}
}

func ParseNetwork(net string) Network {
	switch strings.ToLower(net) {
	case "tcp":
		return NetworkTCP
	case "udp":
		return NetworkUDP
	case "hrtp":
		return NetworkHRTP
	default:
		return NetworkUnknown
	}
}
