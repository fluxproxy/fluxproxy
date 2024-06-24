package route

import (
	"context"
	"fluxway/net"
	"fluxway/proxy"
	"fmt"
	"github.com/bytepowered/assert-go"
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
	proxyType := proxy.ProxyTypeFromContext(ctx)
	switch proxyType {
	case proxy.ProxyType_SOCKS5, proxy.ProxyType_HTTPS:
		assert.MustTrue(income.Destination.IsValid(), "proxy-type: socks/https, income destination must be valid")
		return *income, nil
	default:
		return *income, fmt.Errorf("unsupported proxy-type: %d", proxyType)
	}
}
