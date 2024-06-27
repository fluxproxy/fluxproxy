package proxy

import (
	"context"
	"fluxway/net"
)

// ListenerOptions 监听器的网络参数
type ListenerOptions struct {
	// Common
	Address string
	Port    int
	// TLS
	TLSCertFile string
	TLSKeyFile  string
}

// ListenerHandler 监听器的回调处理函数
type ListenerHandler func(ctx context.Context, conn net.Connection) error

// Listener 监听器，用于建立端口监听，实现网络代理协议的服务端。
type Listener interface {
	// Network 返回当前监听器网络接口的协议类型
	Network() net.Network

	// ServerType 返回当前监听器的代理协议类型
	ServerType() ServerType

	// Init 执行监听器实例的初始化操作
	Init(options ListenerOptions) error

	// Serve 以阻塞状态运行监听器，接收客户端连接。当客户端成功建立连接后，通过 handler 函数回调，进行下一步流量路由。
	Serve(ctx context.Context, handler ListenerHandler) error
}

// Server 代理服务端
type Server interface {
	// Init 初始化服务端
	Init(context.Context) error

	// Serve 以阻塞状态运行服务端
	Serve(context.Context) error
}

// ConnectorSelector 根据连接选择连接至目标地址的Connector
type ConnectorSelector func(*net.Connection) (Connector, bool)

// Connector 远程地址连接器
type Connector interface {
	// DialServe 以阻塞状态建立远程地址连接，进行双向数据读写。
	DialServe(ctx context.Context, link *net.Connection) (err error)
}

// Router 代理路由器
type Router interface {
	// Route 根据监听器建立的连接和代理类型，选择代理请求的远程目标地址。
	Route(ctx context.Context, income *net.Connection) (target net.Connection, err error)
}

// Resolver 域名解析器
type Resolver interface {
	Resolve(ctx context.Context, name string) (net.IP, error)
}
