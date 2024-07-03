package server

import (
	"github.com/rocketmanapp/rocket-proxy/net"
	"strconv"
)

////

type CAddress struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

func (c CAddress) String() string {
	return c.Address + ":" + strconv.Itoa(c.Port)
}

////

type CNetport struct {
	Port    int
	Network net.Network
}
