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
	CtxKeySource             = contextKey{key: "ctx-key-source"}
	CtxKeyConfiger           = contextKey{key: "ctx-key-configer"}
	CtxKeyServerType         = contextKey{key: "ctx-key-server-type"}
	CtxKeyHookFuncDialPhased = contextKey{key: "ctx-key-hook-func-dial-phased"}
)

func SetContextLogID(ctx context.Context, id string, source string) context.Context {
	ctx = context.WithValue(ctx, CtxKeyID, id)
	return context.WithValue(ctx, CtxKeySource, source)
}

func Logger(ctx context.Context) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"source": ctx.Value(CtxKeySource),
		"id":     ctx.Value(CtxKeyID),
	})
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
