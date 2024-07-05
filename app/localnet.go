package app

import (
	"context"
	"github.com/fluxproxy/fluxproxy/net"
	stdnet "net"
)

func loadLocalAddrs(ctx context.Context) []net.Address {
	locals := make([]net.Address, 0)
	localAddrs := lookupLocalIPAddrs(ctx)
	serverPorts := lookupServerPorts(ctx)
	for _, localAddr := range localAddrs {
		for _, sp := range serverPorts {
			// IP Addrs
			locals = append(locals, net.Address{
				Family:  net.ToAddressFamily(localAddr),
				IP:      localAddr,
				Port:    sp.Port,
				Network: sp.Network,
			})
			// Localhost
			locals = append(locals, net.Address{
				Family:  net.AddressFamilyDomain,
				Domain:  "localhost",
				Port:    sp.Port,
				Network: sp.Network,
			})
		}
	}
	return locals
}

func lookupServerPorts(ctx context.Context) []Netport {
	output := make([]Netport, 0, 3+21)
	// http
	var httpConfig HttpConfig
	_ = unmarshalWith(ctx, configPathServerHttp, &httpConfig)
	if httpConfig.Disabled == false && httpConfig.Port > 0 {
		output = append(output, Netport{Port: httpConfig.Port, Network: net.NetworkTCP})
	}
	// socks
	var socksConfig SocksConfig
	_ = unmarshalWith(ctx, configPathServerSocks, &socksConfig)
	if socksConfig.Disabled == false && socksConfig.Port > 0 {
		output = append(output, Netport{Port: socksConfig.Port, Network: net.NetworkTCP})
	}
	return output
}

func lookupLocalIPAddrs(_ context.Context) []stdnet.IP {
	ifaces, err := stdnet.Interfaces()
	if err != nil {
		return make([]stdnet.IP, 0)
	}
	output := make([]stdnet.IP, 0, len(ifaces))
	if len(ifaces) == 0 {
		return output
	}
	for _, iface := range ifaces {
		if iface.Flags&stdnet.FlagUp == 0 {
			continue // down
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

type Netport struct {
	Port    int
	Network net.Network
}
