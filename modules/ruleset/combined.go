package ruleset

import (
	"context"
	"errors"
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
			if errors.Is(err, rocket.ErrRulesetNotMatched) {
				continue
			}
			return ctx, err
		}
	}
	return ctx, rocket.ErrRulesetNotMatched
}
