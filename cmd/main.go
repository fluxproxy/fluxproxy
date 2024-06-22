package main

import (
	"fluxway"
	"github.com/bytepowered/assert-go"
)

func main() {
	inst := fluxway.NewInstance()
	assert.MustNil(inst.Start(), "instance start error")
	defer func() {
		assert.MustNil(inst.Stop(), "instance stop error")
	}()
	assert.MustNil(inst.Serve(), "instance serve error")
}
