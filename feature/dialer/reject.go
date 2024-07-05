package dialer

import (
	"context"
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

func NewRejectDialer() *Reject {
	return &Reject{}
}

func (r *Reject) Name() string {
	return REJECT
}

func (r *Reject) Dial(srcConnCtx context.Context, address net.Address) (rocket.Connection, error) {
	return rocket.NewRejectConnection(), nil
}
