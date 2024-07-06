package ruleset

import (
	"context"
	"fmt"
	"github.com/fluxproxy/fluxproxy"
	"github.com/fluxproxy/fluxproxy/net"
)

var (
	_ proxy.Ruleset = (*LoopBack)(nil)
)

// LoopBack 禁止回环访问本机本服务的端口
type LoopBack struct {
	localAddrs []net.Address
}

func NewLoopback(locals []net.Address) *LoopBack {
	return &LoopBack{
		localAddrs: locals,
	}
}

func (l *LoopBack) Allow(ctx context.Context, permit proxy.Permit) error {
	for _, local := range l.localAddrs {
		if local.Equal(permit.Destination) &&
			local.Port == permit.Destination.Port {
			return fmt.Errorf("loopback: deny: %s", local.String())
		}
	}
	return proxy.ErrNoRulesetMatched
}
