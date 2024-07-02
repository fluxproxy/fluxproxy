package rocket

import (
	"context"
	"github.com/sirupsen/logrus"
)

const (
	CtxKeyID uint32 = iota
	CtxKeyLogger
	CtxKeyConfiger
	CtxKeyServerType
	CtxKeyHookDialPhased
)

func SetContextLogger(ctx context.Context, id string, logger *logrus.Entry) context.Context {
	return context.WithValue(context.WithValue(ctx, CtxKeyID, id), CtxKeyLogger, logger)
}

func Logger(ctx context.Context) *logrus.Entry {
	if v, ok := ctx.Value(CtxKeyLogger).(*logrus.Entry); ok {
		return v
	}
	panic("Logger is not in context.")
}

func RequiredServerType(ctx context.Context) ServerType {
	if v, ok := ctx.Value(CtxKeyServerType).(ServerType); ok {
		return v
	}
	panic("ServerType is not in context.")
}

func RequiredID(ctx context.Context) string {
	if v, ok := ctx.Value(CtxKeyID).(string); ok {
		return v
	}
	panic("ID is not in context.")
}

// Hooks

func ContextWithHookFuncDialPhased(ctx context.Context, v HookFunc) context.Context {
	return context.WithValue(ctx, CtxKeyHookDialPhased, v)
}

func HookFuncDialPhased(ctx context.Context) HookFunc {
	if v, ok := ctx.Value(CtxKeyHookDialPhased).(HookFunc); ok {
		return v
	}
	return nil
}
