package vanity

import (
	"context"
	"vanity/common"
	"vanity/net"
)

const (
	ctxKeyInstance uint32 = iota
	ctxKeyID
	ctxKeyConnection
	ctxKeyLink
	ctxKeyDestination
)

//// Semaphore

func FromContext(ctx context.Context) *Instance {
	if i, ok := ctx.Value(ctxKeyInstance).(*Instance); ok {
		return i
	}
	return nil
}

func MustFromContext(ctx context.Context) *Instance {
	if i, ok := ctx.Value(ctxKeyInstance).(*Instance); ok {
		return i
	}
	panic("Semaphore is not in context.")
}

func contextWith(ctx context.Context, i *Instance) context.Context {
	if FromContext(ctx) != i {
		ctx = context.WithValue(ctx, ctxKeyInstance, i)
	}
	return ctx
}

// ID

func contextWithID(ctx context.Context, id common.ID) context.Context {
	return context.WithValue(ctx, ctxKeyID, id)
}

func IDFromContext(ctx context.Context) common.ID {
	if id, ok := ctx.Value(ctxKeyID).(common.ID); ok {
		return id
	}
	panic("ID is not in context.")
}

// Connection

func contextWithConnection(ctx context.Context, v *net.Connection) context.Context {
	return context.WithValue(ctx, ctxKeyConnection, v)
}

func ConnectionFromContext(ctx context.Context) *net.Connection {
	if id, ok := ctx.Value(ctxKeyConnection).(*net.Connection); ok {
		return id
	}
	panic("Connection is not in context.")
}
