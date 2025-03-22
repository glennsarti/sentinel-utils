package contexts

import (
	"context"
	"errors"

	"github.com/glennsarti/sentinel-utils/lib/languageserver/internal/queues"
)

type LintQueueContextKey struct{}

func WithLintQueue(ctx context.Context, value queues.LintQueue) context.Context {
	return context.WithValue(ctx, LintQueueContextKey{}, value)
}

func LintQueue(ctx context.Context) (queues.LintQueue, error) {
	if value, ok := ctx.Value(LintQueueContextKey{}).(queues.LintQueue); !ok {
		return nil, errors.New("lint queue not found")
	} else {
		return value, nil
	}
}
