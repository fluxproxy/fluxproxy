package rocket

import (
	"context"
	"github.com/knadh/koanf/v2"
	"github.com/sirupsen/logrus"
)

type contextKey struct {
	key string
}

var (
	CtxKeyID       = contextKey{key: "ctx-key-id"}
	CtxKeySource   = contextKey{key: "ctx-key-source"}
	CtxKeyConfiger = contextKey{key: "ctx-key-configer"}
)

func Logger(ctx context.Context) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"id": ctx.Value(CtxKeyID),
	})
}

func Configer(ctx context.Context) *koanf.Koanf {
	if v, ok := ctx.Value(CtxKeyConfiger).(*koanf.Koanf); ok {
		return v
	}
	panic("Configer is not in context.")
}
