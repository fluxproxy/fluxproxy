package rocket

import (
	"context"
	"github.com/rocket-proxy/rocket-proxy/net"
	"io"
	stdnet "net"
)

// Listener 监听器，监听服务端口，完成与客户端的连接握手。
type Listener interface {
	// Init 执行初始化操作
	Init(ctx context.Context) error

	// Listen 以阻塞态监听服务端，接收客户端连接
	Listen(ctx context.Context, dispatcher Dispatcher) error
}

// Dispatcher 管理通道连接请求及路由
type Dispatcher interface {
	// Init 执行初始化操作
	Init(ctx context.Context) error

	// Serve 以阻塞状态运行，处理 Submit 提交的通道连接请求
	Serve(ctx context.Context) error

	// Submit 提交通道连接请求
	Submit(Tunnel)
}

// Connection 表示与目标服务器建立的网络连接
type Connection interface {
	// ReadWriter 返回连接的读写接口
	ReadWriter() io.ReadWriter

	// Conn 返回建立的连接
	Conn() stdnet.Conn

	// Close 关闭连接
	Close() error
}

// Tunnel 连接源客户端与目标服务器的通道
type Tunnel interface {
	// Address 源客户端地址
	Address() net.Address

	// Connect 连接到目标服务器
	Connect(remote Connection)

	// Close 关闭通道
	Close() error

	Context() context.Context
}

// Dialer 建立与目标地址的连接
type Dialer interface {
	Name() string
	Dial(remote net.Address) (Connection, error)
}

//// Hook func

type HookFunc func(ctx context.Context) error
