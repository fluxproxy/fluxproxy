package internal

import (
	"context"
	"github.com/rocket-proxy/rocket-proxy"
)

type hookCtxKey struct {
	key string
}

var (
	CtxHookAfterDialed = hookCtxKey{key: "ctx:hook-func:after-dialed"}
	CtxHookAfterAuthed = hookCtxKey{key: "ctx:hook-func:after-authed"}
)

func ContextWithHook(ctx context.Context, k any, v rocket.HookFunc) context.Context {
	return context.WithValue(ctx, k, v)
}

func LookupHook(ctx context.Context, k any) (f rocket.HookFunc, ok bool) {
	f, ok = ctx.Value(k).(rocket.HookFunc)
	return
}
