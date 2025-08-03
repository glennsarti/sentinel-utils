package contexts

import (
	"context"
	"errors"
)

type sentinelVersionContextKey struct{}

func SetSentinelVersion(ctx context.Context, value *string) error {
	if ret, ok := ctx.Value(sentinelVersionContextKey{}).(*string); !ok {
		return errors.New("sentinel version not found")
	} else {
		*ret = *value
	}
	return nil
}

func WithSentinelVersion(ctx context.Context, value *string) context.Context {
	return context.WithValue(ctx, sentinelVersionContextKey{}, value)
}

func SentinelVersion(ctx context.Context) (string, error) {
	if value, ok := ctx.Value(sentinelVersionContextKey{}).(*string); !ok {
		return "", errors.New("sentinel version not found")
	} else {
		return *value, nil
	}
}
