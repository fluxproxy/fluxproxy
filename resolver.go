package rocket

import (
	"context"
	"github.com/bytepowered/cache"
	"github.com/rocketmanapp/rocket-proxy/net"
	"github.com/rocketmanapp/rocket-proxy/proxy"
	"github.com/sirupsen/logrus"
	stdnet "net"
	"sync"
	"time"
)

var (
	_ proxy.Resolver = (*DNSResolver)(nil)
)

var (
	resolverOnce = sync.Once{}
	resolverInst *DNSResolver
)

type DNSOptions struct {
	CacheSize int `yaml:"cache_size"`
	CacheTTL  int `yaml:"cache_ttl"`
}

type DNSResolver struct {
	cached cache.Cache
}

func NewDNSResolverWith(ctx context.Context) *DNSResolver {
	resolverOnce.Do(func() {
		var opts DNSOptions
		_ = proxy.ConfigUnmarshalWith(ctx, "dns", &opts)
		if opts.CacheSize <= 0 {
			opts.CacheSize = 1024 * 10
		}
		if opts.CacheTTL <= 0 {
			opts.CacheTTL = 60
		}
		resolverInst = &DNSResolver{
			cached: cache.New(opts.CacheSize).
				LRU().
				Expiration(time.Minute * time.Duration(opts.CacheTTL)).
				Build(),
		}
	})
	return resolverInst
}

func (d *DNSResolver) Resolve(ctx context.Context, addr net.Address) (stdnet.IP, error) {
	configer := proxy.Configer(ctx)
	name := addr.String()
	ipv, err := d.cached.GetOrLoad(name, func(_ interface{}) (cache.Expirable, error) {
		// S1: 通过配置文件实现 resolve/rewrite
		if ip := configer.String("dns.hosts." + name); ip != "" {
			if rsv := net.ParseAddress(ip); rsv.Family().IsIP() {
				return cache.NewDefault(rsv.IP), nil
			} else {
				logrus.Warnf("dns.hosts.%s=%s is not ip address", name, ip)
			}
		}
		// S2: IP地址，直接返回
		if addr.Family().IsIP() {
			return cache.NewDefault(addr.IP), nil
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
