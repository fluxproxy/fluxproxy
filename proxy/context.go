package proxy

import (
	"context"
	"fluxway/common"
	"fluxway/net"
	"github.com/knadh/koanf"
)

const (
	ctxKeyConfig uint32 = iota
	ctxKeyID
	ctxKeyConnection
	ctxKeyProxyType
)

// ID

func ContextWithConfig(ctx context.Context, v *koanf.Koanf) context.Context {
	return context.WithValue(ctx, ctxKeyID, v)
}

func ConfigFromContext(ctx context.Context) *koanf.Koanf {
	if v, ok := ctx.Value(ctxKeyID).(*koanf.Koanf); ok {
		return v
	}
	panic("Koanf is not in context.")
}

// ID

func ContextWithID(ctx context.Context, id common.ID) context.Context {
	return context.WithValue(ctx, ctxKeyID, id)
}

func IDFromContext(ctx context.Context) common.ID {
	if v, ok := ctx.Value(ctxKeyID).(common.ID); ok {
		return v
	}
	panic("ID is not in context.")
}

// Connection

func ContextWithConnection(ctx context.Context, v *net.Connection) context.Context {
	return context.WithValue(ctx, ctxKeyConnection, v)
}

func ConnectionFromContext(ctx context.Context) *net.Connection {
	if v, ok := ctx.Value(ctxKeyConnection).(*net.Connection); ok {
		return v
	}
	panic("Connection is not in context.")
}

// ProxyType

func ContextWithProxyType(ctx context.Context, v ProxyType) context.Context {
	return context.WithValue(ctx, ctxKeyProxyType, v)
}

func ProxyTypeFromContext(ctx context.Context) ProxyType {
	if v, ok := ctx.Value(ctxKeyProxyType).(ProxyType); ok {
		return v
	}
	panic("ProxyType is not in context.")
}
