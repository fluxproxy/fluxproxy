package ruleset

import (
	"context"
	"fmt"
	"github.com/fluxproxy/fluxproxy"
	"github.com/fluxproxy/fluxproxy/net"
	stdnet "net"
)

var (
	_ proxy.Ruleset = (*IPNet)(nil)
)

type IPNet struct {
	useSource bool
	isAllow   bool
	nets      []stdnet.IPNet
}

func NewIPNet(isAllow bool, useSource bool, nets []stdnet.IPNet) *IPNet {
	return &IPNet{
		isAllow:   isAllow,
		useSource: useSource,
		nets:      nets,
	}
}

func (i *IPNet) Allow(ctx context.Context, permit proxy.Permit) error {
	var target net.Address
	if i.useSource {
		target = permit.Source
	} else {
		target = permit.Destination
	}
	if i.match(target) {
		if i.isAllow {
			return nil
		} else {
			return fmt.Errorf("ipnet: deny: %s", target)
		}
	} else {
		return proxy.ErrNoRulesetMatched
	}
}

func (i *IPNet) match(target net.Address) bool {
	for _, r := range i.nets {
		if r.Contains(target.IP) {
			return true
		}
	}
	return false
}
