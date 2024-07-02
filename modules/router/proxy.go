package router

import (
	"context"
	"fmt"
	"github.com/bytepowered/assert"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/net"
)

//// 由客户端指定代理目标地址的路由器

var (
	_ rocket.Router = (*ProxyRouter)(nil)
)

type ProxyRouter struct {
}

func NewProxyRouter() *ProxyRouter {
	return &ProxyRouter{}
}

func (d *ProxyRouter) Route(ctx context.Context, income *net.Connection) (target net.Connection, err error) {
	serverType := rocket.RequiredServerType(ctx)
	switch serverType {
	case rocket.ServerTypeSOCKS, rocket.ServerTypeHTTPS:
		assert.MustTrue(income.Destination.IsValid(), "destination must be valid")
		return *income, nil
	default:
		return *income, fmt.Errorf("unsupported server type: %d", serverType)
	}
}
