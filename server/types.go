package server

import (
	"strconv"
)

//// CAddress

type CAddress struct {
	Address string `yaml:"address"`
	Port    int    `yaml:"port"`
}

func (c CAddress) String() string {
	return c.Address + ":" + strconv.Itoa(c.Port)
}
