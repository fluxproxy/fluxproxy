package vanity

import "context"

const (
	ctxKeyServer uint32 = iota
	ctxKeyId
)

func FromContext(ctx context.Context) *Server {
	if s, ok := ctx.Value(ctxKeyServer).(*Server); ok {
		return s
	}
	return nil
}

func MustFromContext(ctx context.Context) *Server {
	if s, ok := ctx.Value(ctxKeyServer).(*Server); ok {
		return s
	}
	panic("Server is not in context.")
}

func toContext(ctx context.Context, v *Server) context.Context {
	if FromContext(ctx) != v {
		ctx = context.WithValue(ctx, ctxKeyServer, v)
	}
	return ctx
}
