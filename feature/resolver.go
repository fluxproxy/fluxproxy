package feature

import (
	"context"
	"github.com/bytepowered/assert"
	"github.com/bytepowered/cache"
	"github.com/fluxproxy/fluxproxy"
	"github.com/fluxproxy/fluxproxy/net"
	"github.com/sirupsen/logrus"
	stdnet "net"
	"sync"
	"time"
)

var (
	_ proxy.Resolver = (*CacheResolver)(nil)
)

var (
	resolverOnce = sync.Once{}
	resolverInst *CacheResolver
)

type Options struct {
	CacheSize int
	CacheTTL  time.Duration
	Hosts     map[string]string
}

type CacheResolver struct {
	cached cache.Cache
}

func (d *CacheResolver) Resolve(ctx context.Context, addr net.Address) (stdnet.IP, error) {
	configer := proxy.Configer(ctx)
	name := addr.Addr()
	ipv, err := d.cached.GetOrLoad(name, func(_ interface{}) (cache.Expirable, error) {
		// S1: 通过配置文件实现 resolve/rewrite
		if userIP := configer.String("resolver.hosts." + name); userIP != "" {
			rAddr, err := net.ParseAddress(net.NetworkTCP, userIP+":80")
			if err != nil {
				logrus.Warnf("resolver.hosts.%s=%s is not ip address", name, userIP)
			} else if rAddr.IsIP() {
				return cache.NewDefault(rAddr.IP), nil
			} else {
				logrus.Warnf("resolver.hosts.%s=%s is not ip address", name, userIP)
			}
		}
		// S2: IP地址，直接返回
		if addr.IsIP() {
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
	return ipv.(stdnet.IP), nil
}

func (d *CacheResolver) Set(name string, ip stdnet.IP) {
	_ = d.cached.Set(name, ip)
}

func InitResolverWith(opts Options) *CacheResolver {
	resolverOnce.Do(func() {
		resolverInst = &CacheResolver{
			cached: cache.New(opts.CacheSize).
				LRU().
				Expiration(opts.CacheTTL).
				Build(),
		}
	})
	return resolverInst
}

func UseResolver() *CacheResolver {
	assert.MustNotNil(resolverInst, "resolver not initialized")
	return resolverInst
}
