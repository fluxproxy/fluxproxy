package proxy

import (
	"github.com/rocket-proxy/rocket-proxy"
	"github.com/rocket-proxy/rocket-proxy/modules/connector"
	"github.com/rocket-proxy/rocket-proxy/net"
	stdnet "net"
)

const (
	DIRECT = "DIRECT"
)

var (
	_ rocket.Proxy = (*Direct)(nil)
)

type Direct struct {
}

func NewDirect() *Direct {
	return &Direct{}
}

func (d *Direct) Name() string {
	return DIRECT
}

func (d *Direct) Generate(address net.Address) (rocket.Connector, error) {
	conn, err := stdnet.Dial("tcp", address.String())
	if err != nil {
		return nil, err
	}
	_ = (conn.(*stdnet.TCPConn)).SetKeepAlive(true)
	return connector.NewDirect(conn), nil
}
