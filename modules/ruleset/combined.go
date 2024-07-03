package ruleset

import (
	"context"
	"github.com/rocketmanapp/rocket-proxy"
)

var (
	_ rocket.Ruleset = (*Combined)(nil)
)

type Combined struct {
	rulesets []rocket.Ruleset
}

func NewCombinedWith(rulesets []rocket.Ruleset) *Combined {
	return &Combined{rulesets: rulesets}
}

func (c *Combined) Allow(ctx context.Context, permit rocket.Permit) (context.Context, error) {
	for _, ruleset := range c.rulesets {
		ctx, err := ruleset.Allow(ctx, permit)
		if err != nil {
			return ctx, err
		}
	}
	return ctx, nil
}
