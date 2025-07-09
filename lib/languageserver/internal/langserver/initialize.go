package langserver

import (
	"context"

	ictx "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/contexts"
	lsp "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/protocol"
	ver "github.com/glennsarti/sentinel-utils/version"
)

func (svc *service) Initialize(ctx context.Context, params lsp.InitializeParams) (lsp.InitializeResult, error) {
	serverCaps := lsp.InitializeResult{
		Capabilities: lsp.ServerCapabilities{
			TextDocumentSync: lsp.TextDocumentSyncOptions{
				OpenClose: true,
				Change:    lsp.Full,
				Save: lsp.SaveOptions{
					IncludeText: false,
				},
			},
			Workspace: &lsp.Workspace6Gn{
				WorkspaceFolders: lsp.WorkspaceFolders5Gn{
					Supported: false,
				},
			},
		},
	}

	serverCaps.ServerInfo.Name = "sentinel-utils-lsp"
	serverCaps.ServerInfo.Version = ver.Version

	clientCaps := params.Capabilities

	if err := ictx.SetClientCapabilities(ctx, &clientCaps); err != nil {
		return serverCaps, err
	}

	if err := svc.setupService(params.RootURI, ctx); err != nil {
		return serverCaps, err
	}

	return serverCaps, nil
}
