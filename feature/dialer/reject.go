package dialer

import (
	"github.com/rocket-proxy/rocket-proxy"
	"github.com/rocket-proxy/rocket-proxy/net"
)

var (
	_ rocket.Dialer = (*Reject)(nil)
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

func (r *Reject) Dial(address net.Address) (rocket.Connection, error) {
	return rocket.NewRejectConnection(), nil
}
