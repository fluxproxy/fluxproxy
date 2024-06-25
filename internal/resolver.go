package internal

import (
	"context"
	"fluxway/proxy"
	"net"
)

var (
	_ proxy.Resolver = (*DNSResolver)(nil)
)

type DNSResolver struct{}

func NewDNSResolver() *DNSResolver {
	return &DNSResolver{}
}

func (d DNSResolver) Resolve(_ context.Context, name string) (net.IP, error) {
	addr, err := net.ResolveIPAddr("ip", name)
	if err != nil {
		return nil, err
	}
	return addr.IP, err
}
