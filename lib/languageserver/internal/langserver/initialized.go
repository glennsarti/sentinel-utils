package langserver

import (
	"context"
	"errors"

	"github.com/creachadair/jrpc2"
	"github.com/glennsarti/sentinel-parser/features"
	ictx "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/contexts"
	lsp "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/protocol"
)

func Initialized(ctx context.Context, params lsp.InitializedParams) error {
	rpcServer := jrpc2.ServerFromContext(ctx)
	if rpcServer == nil {
		return errors.New("missing RPC server from context")
	}

	// Send the version notification
	if ver, err := ictx.SentinelVersion(ctx); err != nil {
		return errors.New("could not determine Sentinel Version")
	} else {
		_ = rpcServer.Notify(ctx, lsp.SentinelVersionCommand, &lsp.SentinelVersionParams{
			SentinelVersion:   ver,
			AvailableVersions: features.SentinelVersions,
		})
	}

	return nil
}
