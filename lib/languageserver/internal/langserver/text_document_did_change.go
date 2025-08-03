package langserver

import (
	"context"

	ictx "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/contexts"
	lsp "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/protocol"
	"github.com/glennsarti/sentinel-utils/lib/languageserver/internal/queues"
)

func (svc *service) TextDocumentDidChange(ctx context.Context, params lsp.DidChangeTextDocumentParams) error {
	ds, err := ictx.DocumentStore(ctx)
	if err != nil {
		return err
	}

	lq, err := ictx.LintQueue(ctx)
	if err != nil {
		return err
	}

	sv, err := ictx.SentinelVersion(ctx)
	if err != nil {
		return err
	}

	// Process the changes
	for _, ce := range params.ContentChanges {
		// No range means the whole document
		if ce.Range == nil {
			if err := ds.UpdateDocument(
				params.TextDocument.URI,
				int(params.TextDocument.Version),
				[]byte(ce.Text),
			); err != nil {
				return err
			}
		}
	}

	req := queues.LintQueueRequest{
		DocId:           string(params.TextDocument.URI),
		DocVersion:      int(params.TextDocument.Version),
		SentinelVersion: sv,
	}
	if err := lq.Enqueue(req); err != nil {
		return err
	}

	return nil
}
