package net

import "fmt"

type Network uint8

const (
	NetworkUNKNOWN Network = iota
	NetworkTCP
	NetworkUDP
	NetworkKCP
)

func (n Network) String() string {
	switch n {
	case NetworkTCP:
		return "tcp"
	case NetworkUDP:
		return "udp"
	case NetworkKCP:
		return "kcp"
	default:
		return "unknown"
	}
}

func ParseNetworkE(network string) (Network, error) {
	switch network {
	case "tcp":
		return NetworkTCP, nil
	case "udp":
		return NetworkUDP, nil
	case "kcp":
		return NetworkKCP, nil
	default:
		return NetworkUNKNOWN, fmt.Errorf("unknown network: %s", network)
	}
}

func ParseNetwork(network string) Network {
	n, err := ParseNetworkE(network)
	if err != nil {
		panic(err)
	}
	return n
}
