package ruleset

import (
	"context"
	"fmt"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/net"
	stdnet "net"
)

var (
	_ rocket.Ruleset = (*IPNet)(nil)
)

type IPNet struct {
	useSource bool //
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

func (i *IPNet) Allow(ctx context.Context, permit rocket.Permit) (context.Context, error) {
	var target net.Address
	if i.useSource {
		target = permit.Source
	} else {
		target = permit.Destination.Address
	}
	if i.match(target) && i.isAllow {
		return ctx, nil
	} else {
		return ctx, fmt.Errorf("%s deny", target)
	}
}

func (i *IPNet) match(target net.Address) bool {
	ip := target.IP()
	for _, r := range i.nets {
		if r.Contains(ip) {
			return true
		}
	}
	return false
}
