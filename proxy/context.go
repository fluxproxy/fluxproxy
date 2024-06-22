package proxy

import (
	"context"
	"vanity/common"
	"vanity/net"
)

const (
	ctxKeyID uint32 = iota
	ctxKeyConnection
)

// ID

func ContextWithID(ctx context.Context, id common.ID) context.Context {
	return context.WithValue(ctx, ctxKeyID, id)
}

func IDFromContext(ctx context.Context) common.ID {
	if id, ok := ctx.Value(ctxKeyID).(common.ID); ok {
		return id
	}
	panic("ID is not in context.")
}

// Connection

func ContextWithConnection(ctx context.Context, v *net.Connection) context.Context {
	return context.WithValue(ctx, ctxKeyConnection, v)
}

func ConnectionFromContext(ctx context.Context) *net.Connection {
	if id, ok := ctx.Value(ctxKeyConnection).(*net.Connection); ok {
		return id
	}
	panic("Connection is not in context.")
}
