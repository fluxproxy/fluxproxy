package app

import (
	"context"
	"github.com/bytepowered/goes"
	"github.com/rocket-proxy/rocket-proxy"
	"github.com/rocket-proxy/rocket-proxy/helper"
	"github.com/rocket-proxy/rocket-proxy/modules/proxy"
	"github.com/rocket-proxy/rocket-proxy/net"
	"github.com/sirupsen/logrus"
	"math"
)

var (
	_ rocket.Dispatcher = (*Dispatcher)(nil)
)

type Dispatcher struct {
	queued   chan rocket.Server
	proxiers map[string]rocket.Proxy
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		queued: make(chan rocket.Server, math.MaxInt32),
	}
}

func (d *Dispatcher) Init(ctx context.Context) error {
	d.proxiers = map[string]rocket.Proxy{
		proxy.DIRECT: proxy.NewDirect(),
		proxy.REJECT: proxy.NewReject(),
	}
	return nil
}

func (d *Dispatcher) Serve(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			logrus.Infof("direct: context done")
			return ctx.Err()

		case server := <-d.queued:
			goes.Go(func() {
				d.handleServer(server)
			})
		}
	}
}

func (d *Dispatcher) Submit(s rocket.Server) {
	d.queued <- s
}

func (d *Dispatcher) handleServer(local rocket.Server) {
	defer helper.Close(local)
	addr := local.Address()
	remote, err := d.lookup(addr).Generate(addr)
	if err != nil {
		logrus.WithError(err).Error("failed to generate rocket connector")
	}
	defer helper.Close(remote)
	local.Connect(remote)
}

func (d *Dispatcher) lookup(addr net.Address) rocket.Proxy {
	return d.proxiers[proxy.DIRECT]
}
