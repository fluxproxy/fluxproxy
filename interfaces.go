package proxy

import (
	"context"
	"github.com/fluxproxy/fluxproxy/net"
	"io"
	stdnet "net"
)

type Authenticate string

const (
	AuthenticateUnknown Authenticate = ""
	AuthenticateAllow   Authenticate = "ALLOW"
	AuthenticateBasic   Authenticate = "BASIC"
	AuthenticateBearer  Authenticate = "BEARER"
	AuthenticateSource  Authenticate = "SOURCE"
	AuthenticateToken   Authenticate = "TOKEN"
)

// Authentication 身份认证信息
type Authentication struct {
	Source         net.Address  // 客户端源地址
	Authenticate   Authenticate // 指定获取身份认证的方式
	Authentication string       // 用于身份验证的凭证
}

// Listener 监听器，监听服务端口，完成与客户端的连接握手。
type Listener interface {
	// Listen 以阻塞态监听服务端，接收客户端连接
	Listen(ctx context.Context) error
}

// Dispatcher 管理通道连接请求及路由
type Dispatcher interface {
	// Authenticate 对客户端进行身份认证
	Authenticate(ctx context.Context, auth Authentication) error

	// Dispatch 执行通道连接（同步执行）
	Dispatch(Connector)
}

// Connection 表示与目标服务器建立的网络连接
type Connection interface {
	io.Closer
	// Conn 返回建立的连接
	Conn() stdnet.Conn
}

// Connector 连接源客户端与目标服务器的通道
type Connector interface {
	io.Closer

	// Destination 返回目标服务器的地址
	Destination() net.Address

	// Source 返回源客户端的地址
	Source() net.Address

	// Connect 连接到目标服务器
	Connect(remote Connection) error

	// HookFunc 根据指定 Key 获取 Hook 函数
	HookFunc(any) (HookFunc, bool)

	// Context 返回通道的 Context
	Context() context.Context
}

// Dialer 建立与目标地址的连接
type Dialer interface {
	// Name 名称
	Name() string
	// Dial 建立与目标地址的连接
	Dial(ctx context.Context, remote net.Address) (Connection, error)
}

// Resolver 域名解析器
type Resolver interface {
	// Resolve 将域名解析成 IP 地址
	Resolve(ctx context.Context, addr net.Address) (stdnet.IP, error)
}

// Authenticator 身份认证
type Authenticator interface {
	Authenticate(context.Context, Authentication) error
}

type Permit struct {
	Source      net.Address
	Destination net.Address
}

type Ruleset interface {
	Allow(context.Context, Permit) error
}

type AuthenticationProvideFunc func(ctx context.Context) Authentication

// HookFunc 注册到Context中的Hook函数
type HookFunc func(ctx context.Context, s error, v ...any) error
