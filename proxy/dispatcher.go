package proxy

import (
	"context"
	"vanity/net"
)

////

var (
	_ Dispatcher = (*StaticDispatcher)(nil)
)

type StaticDispatcher struct {
	target net.Destination
}

func NewStaticDispatcher() *StaticDispatcher {
	return &StaticDispatcher{
		target: net.TCPDestination(net.LocalHostIP, net.Port(1234)),
	}
}

func (d *StaticDispatcher) Dispatch(_ context.Context, income *net.Connection) (target net.Connection, err error) {
	// Socks5, Http
	if income.Destination.IsValid() {
		return *income, nil
	}
	// Static target
	return income.WithDestination(d.target), nil
}
