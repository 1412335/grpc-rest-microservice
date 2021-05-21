package utils

import "context"

type contextKey string

func (c contextKey) String() string {
	return string(c)
}

// GetContextValue gets value from the context.
func GetContextValue(ctx context.Context, key string) (string, bool) {
	value, ok := ctx.Value(contextKey(key)).(string)
	return value, ok
}

// SetContextValue set context with value.
func SetContextValue(ctx context.Context, key string, value interface{}) context.Context {
	return context.WithValue(ctx, contextKey(key), value)
}
