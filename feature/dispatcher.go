package feature

import (
	"context"
	"github.com/bytepowered/goes"
	"github.com/rocket-proxy/rocket-proxy"
	"github.com/rocket-proxy/rocket-proxy/feature/dialer"
	"github.com/rocket-proxy/rocket-proxy/helper"
	"github.com/rocket-proxy/rocket-proxy/net"
	"github.com/sirupsen/logrus"
	"math"
)

var (
	_ rocket.Dispatcher = (*Dispatcher)(nil)
)

type Dispatcher struct {
	tunnels chan rocket.Tunnel
	dialer  map[string]rocket.Dialer
}

func NewDispatcher() *Dispatcher {
	return &Dispatcher{
		tunnels: make(chan rocket.Tunnel, math.MaxInt32),
	}
}

func (d *Dispatcher) Init(ctx context.Context) error {
	d.dialer = map[string]rocket.Dialer{
		dialer.DIRECT: dialer.NewDirect(),
		dialer.REJECT: dialer.NewReject(),
	}
	return nil
}

func (d *Dispatcher) Serve(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			logrus.Infof("dispatcher: serve:done")
			return ctx.Err()

		case server := <-d.tunnels:
			goes.Go(func() {
				d.handleServer(server)
			})
		}
	}
}

func (d *Dispatcher) Submit(s rocket.Tunnel) {
	d.tunnels <- s
}

func (d *Dispatcher) handleServer(local rocket.Tunnel) {
	defer helper.Close(local)
	destAddr := local.Destination()
	// TODO 身份认证
	remote, dErr := d.lookup(destAddr).Dial(local.Context(), destAddr)
	if dErr != nil {
		rocket.Logger(local.Context()).Errorf("dispatcher: dial: %s", dErr)
		return
	}
	defer helper.Close(remote)
	// call hook
	if hook, ok := rocket.LookupHookFunc(local.Context(), rocket.CtxHookFuncOnDialed); ok {
		if hErr := hook(local.Context()); hErr != nil {
			rocket.Logger(local.Context()).Errorf("dispatcher: hook:dial: %s", hErr)
			return
		}
	}
	// connect
	local.Connect(remote)
}

func (d *Dispatcher) lookup(addr net.Address) rocket.Dialer {
	return d.dialer[dialer.DIRECT]
}
