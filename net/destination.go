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
	return net.JoinHostPort(d.Address.String(), d.Port.String())
}

func (d Destination) String() string {
	return d.Network.String() + "//" + d.NetAddr()
}

func (d Destination) IsValid() bool {
	return d.Network != NetworkUnknown
}

func (d Destination) ToTCPAddr() *net.TCPAddr {
	return &net.TCPAddr{IP: d.Address.IP(), Port: int(d.Port)}
}

func (d Destination) ToUDPAddr() *net.UDPAddr {
	return &net.UDPAddr{IP: d.Address.IP(), Port: int(d.Port)}
}
