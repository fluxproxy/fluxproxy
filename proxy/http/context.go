package http

import (
	"context"
	"net/http"
)

const (
	ctxKeyHttpResponseWriter = "http.response-writer"
	ctxKeyHttpHttpRequest    = "http.response-request"
)

func contextWithResponseWriter(ctx context.Context, w http.ResponseWriter) context.Context {
	return context.WithValue(ctx, ctxKeyHttpResponseWriter, w)
}

func requiredResponseWriter(ctx context.Context) http.ResponseWriter {
	v, ok := ctx.Value(ctxKeyHttpResponseWriter).(http.ResponseWriter)
	if ok {
		return v
	}
	panic("ResponseWriter not in context")
}

func contextWithHttpRequest(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, ctxKeyHttpHttpRequest, r)
}

func requiredHttpRequest(ctx context.Context) *http.Request {
	v, ok := ctx.Value(ctxKeyHttpHttpRequest).(*http.Request)
	if ok {
		return v
	}
	panic("*http.Request not in context")
}
