package rocket

import (
	"context"
	"github.com/sirupsen/logrus"
)

type contextKey struct {
	key string
}

var (
	CtxKeyID                 = contextKey{key: "ctx-key-id"}
	CtxKeyLogger             = contextKey{key: "ctx-key-logger"}
	CtxKeyConfiger           = contextKey{key: "ctx-key-configer"}
	CtxKeyServerType         = contextKey{key: "ctx-key-server-type"}
	CtxKeyHookFuncDialPhased = contextKey{key: "ctx-key-hook-func-dial-phased"}
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
	return context.WithValue(ctx, CtxKeyHookFuncDialPhased, v)
}

func HookFuncDialPhased(ctx context.Context) HookFunc {
	if v, ok := ctx.Value(CtxKeyHookFuncDialPhased).(HookFunc); ok {
		return v
	}
	return nil
}
