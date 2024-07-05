package app

import "github.com/rocket-proxy/rocket-proxy/net"

type Netport struct {
	Port    int
	Network net.Network
}
