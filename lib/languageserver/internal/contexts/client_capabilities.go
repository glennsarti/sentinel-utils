package contexts

import (
	"context"
	"errors"

	lsp "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/protocol"
)

type clientCapabilitiesContextKey struct{}

func SetClientCapabilities(ctx context.Context, value *lsp.ClientCapabilities) error {
	if ret, ok := ctx.Value(clientCapabilitiesContextKey{}).(*lsp.ClientCapabilities); !ok {
		return errors.New("client capabilities not found")
	} else {
		*ret = *value
	}
	return nil
}

func WithClientCapabilities(ctx context.Context, value *lsp.ClientCapabilities) context.Context {
	return context.WithValue(ctx, clientCapabilitiesContextKey{}, value)
}

func ClientCapabilities(ctx context.Context) (lsp.ClientCapabilities, error) {
	if value, ok := ctx.Value(clientCapabilitiesContextKey{}).(*lsp.ClientCapabilities); !ok {
		return lsp.ClientCapabilities{}, errors.New("client capabilities not found")
	} else {
		return *value, nil
	}
}
