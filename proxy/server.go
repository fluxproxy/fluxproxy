package proxy

import "context"

type Server interface {
	Init(context.Context) error
	Serve(context.Context) error
}
