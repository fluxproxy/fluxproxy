package feature

import (
	"context"
	"errors"
	"github.com/bytepowered/assert"
	"github.com/rocket-proxy/rocket-proxy"
	"sync"
)

var (
	_ rocket.Ruleset = (*MultiRuleset)(nil)
)

var (
	rulesetOnce = sync.Once{}
	rulesetInst *MultiRuleset
)

type MultiRuleset struct {
	rulesets []rocket.Ruleset
}

func (c *MultiRuleset) Allow(ctx context.Context, permit rocket.Permit) error {
	for _, ruleset := range c.rulesets {
		err := ruleset.Allow(ctx, permit)
		if err != nil {
			if errors.Is(err, rocket.ErrNoRulesetMatched) {
				continue
			}
			return err
		}
	}
	return rocket.ErrNoRulesetMatched
}

func InitMultiRuleset(ruleset []rocket.Ruleset) *MultiRuleset {
	rulesetOnce.Do(func() {
		rulesetInst = &MultiRuleset{rulesets: ruleset}
	})
	return rulesetInst
}

func UseRuleset() *MultiRuleset {
	assert.MustNotNil(rulesetInst, "ruleset not initialized")
	return rulesetInst
}
