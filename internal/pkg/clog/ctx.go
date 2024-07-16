package clog

import "context"

type key int

const (
	clogKey key = 0
)

func NewContext(ctx context.Context, xl *Logger) context.Context {
	return context.WithValue(ctx, clogKey, xl)
}

func FromContext(ctx context.Context) (xl *Logger, ok bool) {
	xl, ok = ctx.Value(clogKey).(*Logger)
	return
}

func FromContextSafe(ctx context.Context) *Logger {
	xl, ok := ctx.Value(clogKey).(*Logger)
	if !ok {
		xl = New()
	}
	return xl
}
