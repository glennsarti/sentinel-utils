package contexts

import (
	"context"
	"errors"

	"github.com/glennsarti/sentinel-utils/lib/languageserver/internal/stores"
)

type DocumentStoreContextKey struct{}

func WithDocumentStore(ctx context.Context, value stores.DocumentStore) context.Context {
	return context.WithValue(ctx, DocumentStoreContextKey{}, value)
}

func DocumentStore(ctx context.Context) (stores.DocumentStore, error) {
	if value, ok := ctx.Value(DocumentStoreContextKey{}).(stores.DocumentStore); !ok {
		return nil, errors.New("document store not found")
	} else {
		return value, nil
	}
}
