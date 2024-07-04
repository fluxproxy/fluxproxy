package rocket

import (
	"context"
	"fmt"
	"github.com/knadh/koanf/v2"
	"github.com/sirupsen/logrus"
)

type contextKey struct {
	key string
}

var (
	CtxKeyID         = contextKey{key: "ctx-key-id"}
	CtxKeySource     = contextKey{key: "ctx-key-source"}
	CtxKeyConfiger   = contextKey{key: "ctx-key-configer"}
	CtxKeyServerType = contextKey{key: "ctx-key-server-type"}
)

var (
	CtxHookFuncOnDialer = contextKey{key: "ctx:hook-func:on-dialer"}
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

//// Configer

func Configer(ctx context.Context) *koanf.Koanf {
	if v, ok := ctx.Value(CtxKeyConfiger).(*koanf.Koanf); ok {
		return v
	}
	panic("Configer is not in context.")
}

func ConfigerUnmarshal(ctx context.Context, path string, out any) error {
	if err := Configer(ctx).UnmarshalWithConf(path, out, koanf.UnmarshalConf{Tag: "yaml"}); err != nil {
		return fmt.Errorf("config unmarshal %s. %w", path, err)
	}
	return nil
}

// Hooks

func ContextWithHookFunc(ctx context.Context, k any, v HookFunc) context.Context {
	return context.WithValue(ctx, k, v)
}

func LookupHookFunc(ctx context.Context, k any) (f HookFunc, ok bool) {
	f, ok = ctx.Value(k).(HookFunc)
	return
}
