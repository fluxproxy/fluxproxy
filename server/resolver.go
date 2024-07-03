package server

import (
	"context"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/modules/resolver"
	"github.com/rocketmanapp/rocket-proxy/net"
	"github.com/sirupsen/logrus"
	"sync"
	"time"
)

var (
	resolverOnce = sync.Once{}
	resolverInst *resolver.CacheResolver
)

func NewResolverWith(ctx context.Context) *resolver.CacheResolver {
	resolverOnce.Do(func() {
		var config ResolverConfig
		_ = rocket.ConfigUnmarshalWith(ctx, "resolver", &config)
		if config.CacheSize <= 0 {
			config.CacheSize = 1024 * 10
		}
		if config.CacheTTL <= 0 {
			config.CacheTTL = 60
		}
		resolverInst = resolver.NewResolverWith(resolver.Options{
			CacheSize: config.CacheSize,
			CacheTTL:  time.Duration(config.CacheTTL) * time.Second,
			Hosts:     config.Hosts,
		})
		// prepare
		for name, ip := range config.Hosts {
			if rsv := net.ParseAddress(ip); rsv.Family().IsIP() {
				resolverInst.Set(name, rsv.IP())
			} else {
				logrus.Warnf("resolver.hosts.%s=%s is not ip address", name, ip)
			}
		}
	})
	return resolverInst
}
