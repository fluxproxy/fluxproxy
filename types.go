package proxy

import (
	"errors"
)

var (
	ErrNoRulesetMatched = errors.New("no-ruleset-matched")
)

// ListenerOptions 监听器的网络参数
type ListenerOptions struct {
	Address string
	Port    int
	Verbose bool
	Auth    bool
}
