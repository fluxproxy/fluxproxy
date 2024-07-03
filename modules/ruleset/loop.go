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

// Loopback 禁止回环访问本机本服务的端口
type Loopback struct {
	localAddrs []net.Destination
}

func NewLoopback(locals []net.Destination) *Loopback {
	return &Loopback{
		localAddrs: locals,
	}
}

func (l *Loopback) Allow(ctx context.Context, permit rocket.Permit) (context.Context, error) {
	for _, local := range l.localAddrs {
		if local.Address.Equal(permit.Destination.Address) &&
			local.Port == permit.Destination.Port {
			return ctx, fmt.Errorf("loopback address %s:%d", local.Address, local.Port)
		}
	}
	//logrus.Infof("loopback: not matched: %+v", permit)
	return ctx, rocket.ErrRulesetNotMatched
}
