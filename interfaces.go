package rocket

import (
	"context"
	"github.com/rocket-proxy/rocket-proxy/net"
	"io"
	stdnet "net"
)

// ListenerOptions 监听器的网络参数
type ListenerOptions struct {
	Address string
	Port    int
}

// Listener 监听器，监听服务端口，完成与客户端的连接握手。
type Listener interface {
	// Init 执行初始化操作
	Init(ctx context.Context) error
	// Listen 以阻塞态监听服务端，接收客户端连接
	Listen(ctx context.Context, dispatcher Dispatcher) error
}

type Dispatcher interface {
	Init(ctx context.Context) error
	Serve(ctx context.Context) error
	Submit(Server)
}

type Connector interface {
	ReadWriter() io.ReadWriter
	Conn() stdnet.Conn
	Close() error
}

type Server interface {
	Address() net.Address
	Connect(Connector)
	Close() error
	Done() <-chan struct{}
}

type Proxy interface {
	Name() string
	Generate(address net.Address) (Connector, error)
}

//// Hook func

type HookFunc func(ctx context.Context) error
