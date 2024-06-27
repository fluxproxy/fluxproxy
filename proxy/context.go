package proxy

import (
	"context"
	"fmt"
	"github.com/knadh/koanf"
	"github.com/sirupsen/logrus"
)

const (
	// commons
	ctxKeyID uint32 = iota
	ctxKeyLogger
	ctxKeyConfiger
	ctxKeyConnection
	ctxKeyProxyType
	// hooks
	ctxKeyHookDailPhased
)

// ID

func ContextWithLogger(ctx context.Context, v *logrus.Entry) context.Context {
	return context.WithValue(ctx, ctxKeyLogger, v)
}

func RequiredLogger(ctx context.Context) *logrus.Entry {
	if v, ok := ctx.Value(ctxKeyLogger).(*logrus.Entry); ok {
		return v
	}
	panic("ctxKeyLogger is not in context.")
}

// Config

func ContextWithConfiger(ctx context.Context, v *koanf.Koanf) context.Context {
	return context.WithValue(ctx, ctxKeyConfiger, v)
}

func RequiredConfiger(ctx context.Context) *koanf.Koanf {
	if v, ok := ctx.Value(ctxKeyConfiger).(*koanf.Koanf); ok {
		return v
	}
	panic("Configure 'Koanf' is not in context.")
}

// ID

func ContextWithID(ctx context.Context, v string) context.Context {
	return context.WithValue(ctx, ctxKeyID, v)
}

func RequiredID(ctx context.Context) string {
	if v, ok := ctx.Value(ctxKeyID).(string); ok {
		return v
	}
	panic("ID is not in context.")
}

// ProxyType

func ContextWithProxyType(ctx context.Context, v ProxyType) context.Context {
	return context.WithValue(ctx, ctxKeyProxyType, v)
}

func RequiredProxyType(ctx context.Context) ProxyType {
	if v, ok := ctx.Value(ctxKeyProxyType).(ProxyType); ok {
		return v
	}
	panic("ProxyType is not in context.")
}

// Utils

func UnmarshalConfig(ctx context.Context, path string, out any) error {
	k := RequiredConfiger(ctx)
	if err := k.UnmarshalWithConf(path, out, koanf.UnmarshalConf{Tag: "yaml"}); err != nil {
		return fmt.Errorf("unmarshal %s options: %w", path, err)
	}
	return nil
}

// Hooks

func ContextWithHookDialPhased(ctx context.Context, v HookFunc) context.Context {
	return context.WithValue(ctx, ctxKeyHookDailPhased, v)
}

func LookupHookDialPhased(ctx context.Context) HookFunc {
	if v, ok := ctx.Value(ctxKeyHookDailPhased).(HookFunc); ok {
		return v
	}
	return nil
}
