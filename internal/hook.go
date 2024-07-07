package internal

import (
	"context"
	"github.com/fluxproxy/fluxproxy"
)

type hookCtxKey struct {
	key string
}

var (
	CtxHookAfterResolve = hookCtxKey{key: "ctx:hook-func:after-resolve"}
	CtxHookAfterDial    = hookCtxKey{key: "ctx:hook-func:after-dial"}
	CtxHookAfterRuleset = hookCtxKey{key: "ctx:hook-func:after-ruleset"}
	CtxHookAfterConnect = hookCtxKey{key: "ctx:hook-func:after-connect"}
)

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
