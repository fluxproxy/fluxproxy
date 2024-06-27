package proxy

import (
	"context"
	"fluxway/net"
)

type HookFunc func(ctx context.Context, conn *net.Connection) error
