package ruleset

import (
	"context"
	"fmt"
	"github.com/rocketmanapp/rocket-proxy"
	"github.com/rocketmanapp/rocket-proxy/net"
	stdnet "net"
	"sync"
)

var (
	_ rocket.Ruleset = (*Loopback)(nil)
)

var (
	lookbackOnce sync.Once
	lookbackInst *Loopback
)

type Loopback struct {
	localAddrs []net.Destination
}

func NewLoopbackWith(ctx context.Context) *Loopback {
	lookbackOnce.Do(func() {
		// 内部端口
		locals := make([]net.Destination, 0)
		localAddrs := lookupLocalAddrs(ctx)
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
		// 外部端口（如公网地址）
		lookbackInst = &Loopback{localAddrs: locals}
	})
	return lookbackInst
}

func (l *Loopback) Allow(ctx context.Context, permit rocket.Permit) (context.Context, error) {
	// 禁止回环访问
	for _, local := range l.localAddrs {
		if local.Address.Equal(permit.Destination.Address) &&
			local.Port == permit.Destination.Port {
			return ctx, fmt.Errorf("loopback address %s:%d", local.Address, local.Port)
		}
	}
	return ctx, nil
}

func lookupServerPorts(ctx context.Context) []portmeta {
	output := make([]portmeta, 0, 3+21)
	configer := rocket.Configer(ctx)
	// http/https
	if false == configer.Bool("https.disabled") {
		http := configer.Int("server.http_port")
		https := configer.Int("server.https_port")
		if http > 0 {
			output = append(output, portmeta{Port: http, Network: net.NetworkTCP})
		}
		if https > 0 {
			output = append(output, portmeta{Port: https, Network: net.NetworkTCP})
		}
	}
	// socks
	if false == configer.Bool("socks.disabled") {
		socks := configer.Int("server.socks_port")
		if socks > 0 {
			output = append(output, portmeta{Port: socks, Network: net.NetworkTCP})
		}
	}
	// forward
	type Item struct {
		Network  string `json:"network"`
		Disabled bool   `yaml:"disabled"`
		Port     int    `json:"port"`
	}
	type Rules struct {
		Rules []Item `json:"rules"`
	}
	var rules Rules
	_ = rocket.ConfigUnmarshalWith(ctx, "forward.rules", &rules)
	for _, rule := range rules.Rules {
		if rule.Disabled {
			continue
		}
		if rule.Port > 0 {
			output = append(output, portmeta{Port: rule.Port, Network: net.ParseNetwork(rule.Network)})
		}
	}
	return output
}

func lookupLocalAddrs(_ context.Context) []net.IP {
	ifaces, err := stdnet.Interfaces()
	if err != nil {
		return make([]net.IP, 0)
	}
	output := make([]net.IP, 0, len(ifaces))
	if len(ifaces) == 0 {
		return output
	}
	for _, iface := range ifaces {
		if iface.Flags&stdnet.FlagUp == 0 {
			continue // interface down
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *stdnet.IPNet:
				output = append(output, v.IP)
			case *stdnet.IPAddr:
				output = append(output, v.IP)
			}
		}
	}
	return output
}

type portmeta struct {
	Port    int
	Network net.Network
}
