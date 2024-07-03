package ruleset

import (
	"context"
	"fmt"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/net"
)

var (
	_ rocket.Ruleset = (*Loopback)(nil)
)

type Loopback struct {
	localAddrs []net.Destination
}

func NewLoopback(localAddrs []net.Destination) *Loopback {
	return &Loopback{localAddrs: localAddrs}
}

func (l *Loopback) Allow(ctx context.Context, permit rocket.Permit) (context.Context, error) {
	// 禁止回环访问
	for _, local := range l.localAddrs {
		if local.Address == permit.Destination.Address && local.Port == permit.Destination.Port {
			return ctx, fmt.Errorf("loopback address %s:%d", local.Address, local.Port)
		}
	}
	return ctx, nil
}
