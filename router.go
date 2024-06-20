package avoidy

import "avoidy/net"

type Router interface {
	Network() []net.Network
}
