package proxy

import (
	"context"
	"fluxway/net"
	"github.com/knadh/koanf"
	"github.com/sirupsen/logrus"
)

const (
	ctxKeyLogger uint32 = iota
	ctxKeyConfig
	ctxKeyID
	ctxKeyConnection
	ctxKeyProxyType
)

// ID

func ContextWithLogger(ctx context.Context, v *logrus.Entry) context.Context {
	return context.WithValue(ctx, ctxKeyLogger, v)
}

func LoggerFromContext(ctx context.Context) *logrus.Entry {
	if v, ok := ctx.Value(ctxKeyLogger).(*logrus.Entry); ok {
		return v
	}
	panic("ctxKeyLogger is not in context.")
}

// Config

func ContextWithConfig(ctx context.Context, v *koanf.Koanf) context.Context {
	return context.WithValue(ctx, ctxKeyConfig, v)
}

func ConfigFromContext(ctx context.Context) *koanf.Koanf {
	if v, ok := ctx.Value(ctxKeyConfig).(*koanf.Koanf); ok {
		return v
	}
	panic("Koanf is not in context.")
}

// ID

func ContextWithID(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, ctxKeyID, v)
}

func IDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyID).(string); ok {
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
