package rocket

import (
	"context"
	"github.com/bytepowered/cache"
	"net"
	"rocket/proxy"
	"time"
)

var (
	_ proxy.Resolver = (*DNSResolver)(nil)
)

type DNSResolver struct {
	cached cache.Cache
}

func NewDNSResolver() *DNSResolver {
	return &DNSResolver{
		cached: cache.New(1000).
			LRU().
			Expiration(time.Minute * 10).
			Build(),
	}
}

func (d *DNSResolver) Resolve(_ context.Context, name string) (net.IP, error) {
	ipv, err := d.cached.GetOrLoad(name, func(_ interface{}) (cache.Expirable, error) {
		addr, err := net.ResolveIPAddr("ip", name)
		if err != nil {
			return cache.Expirable{}, err
		}
		return cache.NewDefault(addr.IP), err
	})
	if err != nil {
		return nil, err
	}
	return ipv.(net.IP), nil
}
