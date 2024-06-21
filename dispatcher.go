package vanity

import (
	"context"
	"log"
	"vanity/net"
)

type Dispatcher struct {
	listener Listener
	routers  []Router
}

func NewDispatcher(listener Listener, routers []Router) *Dispatcher {
	return &Dispatcher{
		listener: listener,
		routers:  routers,
	}
}

func (d *Dispatcher) Process(ctx context.Context, conn net.Connection) {
	for _, router := range d.routers {
		if err := router.Route(ctx, conn); err != nil {
			log.Printf("%s listener route error. %s", d.listener.Tag(), err)
		}
	}
}
