package rocket

import (
	"errors"
)

var (
	ErrNoRulesetMatched = errors.New("no-ruleset-matched")
)

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
	Verbose bool
	Auth    bool
}
