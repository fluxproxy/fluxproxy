package feature

import (
	"context"
	"errors"
	"github.com/bytepowered/assert"
	"sync"
)

var (
	_ proxy.Ruleset = (*MultiRuleset)(nil)
)

var (
	rulesetOnce = sync.Once{}
	rulesetInst *MultiRuleset
)

type MultiRuleset struct {
	rulesets []proxy.Ruleset
}

func (c *MultiRuleset) Allow(ctx context.Context, permit proxy.Permit) error {
	for _, ruleset := range c.rulesets {
		err := ruleset.Allow(ctx, permit)
		if err != nil {
			if errors.Is(err, proxy.ErrNoRulesetMatched) {
				continue
			}
			return err
		}
	}
	return proxy.ErrNoRulesetMatched
}

func InitMultiRuleset(ruleset []proxy.Ruleset) *MultiRuleset {
	rulesetOnce.Do(func() {
		rulesetInst = &MultiRuleset{rulesets: ruleset}
	})
	return rulesetInst
}

func UseRuleset() *MultiRuleset {
	assert.MustNotNil(rulesetInst, "ruleset not initialized")
	return rulesetInst
}
