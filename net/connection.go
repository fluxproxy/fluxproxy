package net

import (
	"context"
	"io"
	"net"
)

type Connection struct {
	Context context.Context
	Address net.Addr
	Network Network
	io.ReadWriter
}
