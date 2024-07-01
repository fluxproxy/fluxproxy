package proxy

import (
	"context"
	"fmt"
	"github.com/knadh/koanf/v2"
)

func Configer(ctx context.Context) *koanf.Koanf {
	if v, ok := ctx.Value(CtxKeyConfiger).(*koanf.Koanf); ok {
		return v
	}
	panic("Configer is not in context.")
}

func ConfigUnmarshalWith(ctx context.Context, path string, out any) error {
	if err := Configer(ctx).UnmarshalWithConf(path, out, koanf.UnmarshalConf{Tag: "yaml"}); err != nil {
		return fmt.Errorf("unmarshal %s config: %w", path, err)
	}
	return nil
}
