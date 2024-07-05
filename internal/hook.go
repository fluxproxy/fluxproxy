package internal

import (
	"context"
)

type hookCtxKey struct {
	key string
}

var (
	CtxHookAfterDialed  = hookCtxKey{key: "ctx:hook-func:after-dialed"}
	CtxHookAfterRuleset = hookCtxKey{key: "ctx:hook-func:after-ruleset"}
)

func ContextWithHook(ctx context.Context, k any, v proxy.HookFunc) context.Context {
	return context.WithValue(ctx, k, v)
}

func ContextWithHooks(ctx context.Context, hooks map[any]proxy.HookFunc) context.Context {
	for k, f := range hooks {
		ctx = context.WithValue(ctx, k, f)
	}
	return ctx
}

func LookupHook(ctx context.Context, k any) (f proxy.HookFunc, ok bool) {
	f, ok = ctx.Value(k).(proxy.HookFunc)
	return
}
