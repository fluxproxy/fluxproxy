package proxy

import (
	"github.com/rocket-proxy/rocket-proxy"
	"github.com/rocket-proxy/rocket-proxy/modules/connector"
	"github.com/rocket-proxy/rocket-proxy/net"
)

var (
	_ rocket.Proxy = (*Reject)(nil)
)

const (
	REJECT = "REJECT"
)

type Reject struct {
}

func NewReject() *Reject {
	return &Reject{}
}

func (r *Reject) Name() string {
	return REJECT
}

func (r *Reject) Generate(address net.Address) (rocket.Connector, error) {
	return connector.NewReject(), nil
}
