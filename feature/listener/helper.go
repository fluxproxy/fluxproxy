package listener

import (
	"github.com/bytepowered/assert"
	"github.com/rocket-proxy/rocket-proxy/net"
	stdnet "net"
)

func parseRemoteAddress(remoteAddr string) net.Address {
	host, _, hpErr := stdnet.SplitHostPort(remoteAddr)
	assert.MustNil(hpErr, "http: parse host port error: %s", hpErr)
	srcAddr, _ := net.ParseAddress(net.NetworkTCP, host)
	assert.MustTrue(srcAddr.IsIP(), "http: srcAddr is not ip")
	return srcAddr
}
