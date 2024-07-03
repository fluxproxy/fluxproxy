package resolver

import (
	"context"
	"github.com/bytepowered/cache"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/net"
	"github.com/sirupsen/logrus"
	stdnet "net"
	"time"
)

var (
	_ rocket.Resolver = (*CacheResolver)(nil)
)

type Options struct {
	CacheSize int
	CacheTTL  time.Duration
	Hosts     map[string]string
}

type CacheResolver struct {
	cached cache.Cache
}

func NewResolverWith(opts Options) *CacheResolver {
	return &CacheResolver{
		cached: cache.New(opts.CacheSize).
			LRU().
			Expiration(opts.CacheTTL).
			Build(),
	}
}

func (d *CacheResolver) Resolve(ctx context.Context, addr net.Address) (stdnet.IP, error) {
	configer := rocket.Configer(ctx)
	name := addr.String()
	ipv, err := d.cached.GetOrLoad(name, func(_ interface{}) (cache.Expirable, error) {
		// S1: 通过配置文件实现 resolve/rewrite
		if ip := configer.String("resolver.hosts." + name); ip != "" {
			if rsv := net.ParseAddress(ip); rsv.Family().IsIP() {
				return cache.NewDefault(rsv.IP()), nil
			} else {
				logrus.Warnf("resolver.hosts.%s=%s is not ip address", name, ip)
			}
		}
		// S2: IP地址，直接返回
		if addr.Family().IsIP() {
			return cache.NewDefault(addr.IP()), nil
		}
		// S3: 尝试解析域名
		addr, err := stdnet.ResolveIPAddr("ip", name)
		if err != nil {
			return cache.Expirable{Value: nil}, err
		}
		return cache.NewDefault(addr.IP), err
	})
	if err != nil {
		return nil, err
	}
	return ipv.(net.IP), nil
}

func (d *CacheResolver) Set(name string, ip stdnet.IP) {
	_ = d.cached.Set(name, ip)
}
