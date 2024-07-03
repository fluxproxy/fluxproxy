package server

import (
	"context"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/internal"
	"github.com/rocketmanapp/rocket-proxy/modules/ruleset"
	"github.com/rocketmanapp/rocket-proxy/net"
	"github.com/sirupsen/logrus"
	stdnet "net"
	"strings"
	"sync"
)

var (
	combinedOnce sync.Once
	combinedInst *ruleset.Combined
)

func NewRulesetWith(ctx context.Context) *ruleset.Combined {
	combinedOnce.Do(func() {
		// 最高优先级：禁止回环访问
		rulesets := []rocket.Ruleset{
			ruleset.NewLoopback(loadLocalAddrs(ctx)),
		}
		// 第二优化级：根据配置文件
		rulesets = append(rulesets, loadRuleset(ctx)...)
		combinedInst = ruleset.NewCombinedWith(rulesets)
	})
	return combinedInst
}

func loadRuleset(ctx context.Context) []rocket.Ruleset {
	var rules []RulesetConfig
	_ = rocket.ConfigUnmarshalWith(ctx, "ruleset", &rules)
	var rulesets = make([]rocket.Ruleset, 0, len(rules))
	for _, config := range rules {
		switch strings.ToLower(config.Type) {
		case "ipnet":
			rulesets = append(rulesets, newIPNetRuleset(config))
		}
	}
	return rulesets
}

func newIPNetRuleset(config RulesetConfig) rocket.Ruleset {
	nets := make([]stdnet.IPNet, 0, len(config.Address))
	for _, sAddr := range config.Address {
		if _, ipNet, err := stdnet.ParseCIDR(sAddr); err == nil {
			nets = append(nets, *ipNet)
		} else {
			logrus.Warnf("invalid ruleset(ipnet) address: %s", sAddr)
		}
	}
	return ruleset.NewIPNet(strings.EqualFold(config.Access, "allow"), strings.EqualFold(config.Origin, "source"), nets)
}

func loadLocalAddrs(ctx context.Context) []net.Destination {
	locals := make([]net.Destination, 0)
	localAddrs := internal.LookupLocalIPAddrs(ctx)
	serverPorts := lookupServerPorts(ctx)
	for _, localAddr := range localAddrs {
		for _, sp := range serverPorts {
			locals = append(locals, net.Destination{
				Address: net.IPAddress(localAddr),
				Port:    net.Port(sp.Port),
				Network: sp.Network,
			})
		}
	}
	return locals
}

func lookupServerPorts(ctx context.Context) []CNetport {
	output := make([]CNetport, 0, 3+21)
	configer := rocket.Configer(ctx)
	var serverConfig ServerConfig
	_ = rocket.ConfigUnmarshalWith(ctx, "server", &serverConfig)
	// http/https
	if false == configer.Bool("https.disabled") {
		if serverConfig.HttpPort > 0 {
			output = append(output, CNetport{Port: serverConfig.HttpPort, Network: net.NetworkTCP})
		}
		if serverConfig.HttpsPort > 0 {
			output = append(output, CNetport{Port: serverConfig.HttpsPort, Network: net.NetworkTCP})
		}
	}
	// socks
	if false == configer.Bool("socks.disabled") {
		if serverConfig.SocksPort > 0 {
			output = append(output, CNetport{Port: serverConfig.SocksPort, Network: net.NetworkTCP})
		}
	}
	// forward
	var forwardConfig ForwardConfig
	_ = rocket.ConfigUnmarshalWith(ctx, "forward", &forwardConfig)
	for _, rule := range forwardConfig.Rules {
		if rule.Disabled {
			continue
		}
		if rule.Port > 0 {
			output = append(output, CNetport{Port: rule.Port, Network: net.ParseNetwork(rule.Network)})
		}
	}
	return output
}
