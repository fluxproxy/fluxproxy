package dialer

import (
	"context"
	"github.com/fluxproxy/fluxproxy"
	"github.com/fluxproxy/fluxproxy/net"
)

var (
	_ proxy.Dialer = (*Reject)(nil)
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

func (r *Reject) Dial(srcConnCtx context.Context, address net.Address) (proxy.Connection, error) {
	return proxy.NewRejectConnection(), nil
}
