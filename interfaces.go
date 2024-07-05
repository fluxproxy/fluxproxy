package rocket

import (
	"context"
	"github.com/rocket-proxy/rocket-proxy/net"
	"io"
	stdnet "net"
)

// Authentication 身份认证信息
type Authentication struct {
	Source         net.Address
	Authenticate   string
	Authentication string
}

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
	// Destination 返回目标服务器的地址
	Destination() net.Address

	// Authentication 返回源客户端的身份认证信息
	Authentication() Authentication

	// Connect 连接到目标服务器
	Connect(remote Connection) error

	// Context 返回通道的 Context
	Context() context.Context

	// Close 关闭通道
	Close() error
}

// Dialer 建立与目标地址的连接
type Dialer interface {
	Name() string
	Dial(ctx context.Context, remote net.Address) (Connection, error)
}

// Resolver 域名解析器
type Resolver interface {
	Resolve(ctx context.Context, addr net.Address) (stdnet.IP, error)
}

// Authenticator 身份认证
type Authenticator interface {
	Authenticate(context.Context, Authentication) error
}

// HookFunc 注册到Context中的Hook函数
type HookFunc func(ctx context.Context, s error, v ...any) error
