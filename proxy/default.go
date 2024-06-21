package proxy

import (
	"context"
	"vanity"
	"vanity/net"
)

var (
	_ vanity.Inbound = (*RawInbound)(nil)
)

type RawInbound struct {
}

func (d *RawInbound) Process(ctx context.Context, link *net.Connection) error {
	return nil
}

////

var (
	_ vanity.Outbound = (*DirectOutbound)(nil)
)

type DirectOutbound struct {
}

func (d *DirectOutbound) DailServe(ctx context.Context, conn *net.Link) error {
	return nil
}
