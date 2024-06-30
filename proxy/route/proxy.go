package route

import (
	"context"
	"fmt"
	"github.com/bytepowered/assert-go"
	"github.com/rocketmanapp/rocket-proxy/net"
	"github.com/rocketmanapp/rocket-proxy/proxy"
)

//// 由客户端指定代理目标地址的路由器

var (
	_ proxy.Router = (*ProxyRouter)(nil)
)

type ProxyRouter struct {
}

func NewProxyRouter() *ProxyRouter {
	return &ProxyRouter{}
}

func (d *ProxyRouter) Route(ctx context.Context, income *net.Connection) (target net.Connection, err error) {
	proxyType := proxy.RequiredProxyType(ctx)
	switch proxyType {
	case proxy.ServerType_SOCKS5, proxy.ServerType_HTTPS:
		assert.MustTrue(income.Destination.IsValid(), "destination must be valid")
		return *income, nil
	default:
		return *income, fmt.Errorf("unsupported proxy type: %d", proxyType)
	}
}
