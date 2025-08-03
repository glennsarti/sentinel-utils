package langserver

import (
	"context"

	ictx "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/contexts"
	lsp "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/protocol"
	"github.com/glennsarti/sentinel-utils/lib/languageserver/internal/queues"
)

func (svc *service) TextDocumentDidOpen(ctx context.Context, params lsp.DidOpenTextDocumentParams) error {
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

	if err := ds.SetDocument(
		params.TextDocument.URI,
		string(params.TextDocument.LanguageID),
		int(params.TextDocument.Version),
		[]byte(params.TextDocument.Text),
	); err != nil {
		return err
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
