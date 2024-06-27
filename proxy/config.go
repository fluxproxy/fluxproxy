package proxy

import (
	"context"
	"fmt"
	"github.com/knadh/koanf"
)

// Utils

func UnmarshalConfig(ctx context.Context, path string, out any) error {
	k, ok := ctx.Value(CtxKeyConfiger).(*koanf.Koanf)
	if !ok {
		panic("Configure 'Koanf' is not in context.")
	}
	if err := k.UnmarshalWithConf(path, out, koanf.UnmarshalConf{Tag: "yaml"}); err != nil {
		return fmt.Errorf("unmarshal %s options: %w", path, err)
	}
	return nil
}
