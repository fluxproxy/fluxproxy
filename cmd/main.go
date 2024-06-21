package main

import (
	"github.com/bytepowered/assert-go"
	"vanity"
)

func main() {
	inst := vanity.NewInstance()
	assert.MustNil(inst.Start(), "instance start error")
	defer func() {
		assert.MustNil(inst.Stop(), "instance stop error")
	}()
	assert.MustNil(inst.Serve(), "instance serve error")
}
