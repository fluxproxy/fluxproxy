package rocket

//// Authenticate Types

const (
	AuthenticateSource = "Source"
	AuthenticateBasic  = "Basic"
	AuthenticateBearer = "Bearer"
	AuthenticateToken  = "Token"
)

// ListenerOptions 监听器的网络参数
type ListenerOptions struct {
	Address string
	Port    int
}
