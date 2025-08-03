package langserver

import (
	"context"
	"errors"
	"fmt"

	"github.com/creachadair/jrpc2"
	"github.com/glennsarti/sentinel-parser/features"
	ictx "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/contexts"
	lsp "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/protocol"
	"github.com/glennsarti/sentinel-utils/lib/languageserver/internal/queues"
)

func (svc *service) SentinelSetVersion(ctx context.Context, params lsp.SentinelSetVersionRequest) (any, error) {
	response := lsp.SetSentinelVersionResponse{
		AvailableVersions: features.SentinelVersions,
	}

	ver, err := ictx.SentinelVersion(ctx)
	if err != nil {
		return response, errors.New("could not determine Sentinel Version")
	}

	lq, err := ictx.LintQueue(ctx)
	if err != nil {
		return nil, err
	}

	// Did it actually change?
	if ver == params.Version {
		response.SentinelVersion = ver
		return response, nil
	}
	// Is the new version valid?
	versionOk, actualVersion := features.ValidateSentinelVersion(params.Version)

	rpcServer := jrpc2.ServerFromContext(ctx)
	if rpcServer == nil {
		return response, errors.New("missing RPC server from context")
	}

	if !versionOk {
		_ = rpcServer.Notify(ctx, "window/showMessage", &lsp.ShowMessageParams{
			Type:    lsp.Warning,
			Message: fmt.Sprintf("%q is not a valid Sentinel vesion.", params.Version),
		})
		return response, fmt.Errorf("%q is not a valid Sentinel vesion", params.Version)
	}

	// Set the new version
	if err := ictx.SetSentinelVersion(ctx, &actualVersion); err != nil {
		return response, fmt.Errorf("failed to set Sentinel version: %w", err)
	}
	response.SentinelVersion = params.Version

	// Enqueue new diagnostics
	req := queues.LintQueueRequest{
		SentinelVersion: actualVersion,
	}
	if err := lq.Enqueue(req); err != nil {
		return nil, err
	}

	return response, nil
}
