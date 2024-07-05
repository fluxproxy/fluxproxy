package rocket

//// Authenticate Types

const (
	AuthenticateAllow  = "ALLOW"
	AuthenticateBasic  = "BASIC"
	AuthenticateBearer = "BEARER"
	AuthenticateSource = "SOURCE"
	AuthenticateToken  = "TOKEN"
)

// ListenerOptions 监听器的网络参数
type ListenerOptions struct {
	Address string
	Port    int
}
