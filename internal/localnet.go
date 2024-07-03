package internal

import (
	"context"
	"github.com/rocketmanapp/rocket-proxy/net"
	stdnet "net"
)

func LookupLocalIPAddrs(_ context.Context) []net.IP {
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
