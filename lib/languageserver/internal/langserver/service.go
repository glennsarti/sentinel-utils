package langserver

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/creachadair/jrpc2"
	rpch "github.com/creachadair/jrpc2/handler"
	jserver "github.com/creachadair/jrpc2/server"
	ictx "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/contexts"
	"github.com/glennsarti/sentinel-utils/lib/languageserver/internal/filesystem"
	wrapFS "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/filesystem/docstore_wrapper"
	osFS "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/filesystem/os"
	lsp "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/protocol"
	"github.com/glennsarti/sentinel-utils/lib/languageserver/internal/session"
	"github.com/glennsarti/sentinel-utils/lib/languageserver/internal/stores"
	storesImpl "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/stores/concrete"

	"github.com/glennsarti/sentinel-utils/lib/languageserver/internal/queues"
	dispatchImpl "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/queues/client_dispatch"
	lintImpl "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/queues/lint"
)

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"log"

// 	jserver "github.com/creachadair/jrpc2/server"

// 	"github.com/creachadair/jrpc2"
// 	"github.com/creachadair/jrpc2/code"
// 	rpch "github.com/creachadair/jrpc2/handler"

// 	lsctx "github.com/glennsarti/vscode-sentinel/lsp/internal/context"
// 	"github.com/glennsarti/vscode-sentinel/lsp/internal/filesystem"
// 	"github.com/glennsarti/vscode-sentinel/lsp/internal/langserver/notifiers"
// 	"github.com/glennsarti/vscode-sentinel/lsp/internal/langserver/session"
// 	ilsp "github.com/glennsarti/vscode-sentinel/lsp/internal/lsp"
// 	lsp "github.com/glennsarti/vscode-sentinel/lsp/internal/protocol"
// 	"github.com/glennsarti/vscode-sentinel/lsp/internal/settings"
// 	"github.com/glennsarti/vscode-sentinel/lsp/internal/telemetry"
// 	"github.com/hashicorp/hcl-lang/decoder"
// 	"github.com/hashicorp/hcl-lang/lang"
// )

func NewService(logger *log.Logger, ctx context.Context) jserver.Service {
	//srvCtx, stopService := context.WithCancel(ctx)
	srv := service{
		logger: logger,
		srvCtx: ctx,
		//stopService: stopService,
	}
	return &srv
}

type service struct {
	logger     *log.Logger
	srvCtx     context.Context
	stateStore stores.StateStore
	sessionFS  filesystem.SessionFS

	lintQueue         queues.LintQueue
	clientNotifyQueue queues.ClientNotifyDispatchQueue
}

func (svc *service) setupService(rootUri lsp.DocumentURI, ctx context.Context) error {
	rpcServer := jrpc2.ServerFromContext(ctx)
	if rpcServer == nil {
		return errors.New("missing RPC server from context")
	}

	baseFS, err := osFS.NewOSFileSystem(rootUri)
	if err != nil {
		return err
	}
	svc.logger.Printf("Created new OS file system called: %s", baseFS.Name())

	// VSCode can do strange things to URIs. So we need to normalise
	// the names so we can do comparisons easier.
	normaliser := func(rawUri lsp.DocumentURI) lsp.DocumentURI {
		if p, err := baseFS.UriToPath(rawUri); err == nil {
			if u, err := baseFS.PathToUri(p); err == nil {
				return u
			}
		}
		return rawUri
	}

	if s, err := storesImpl.NewStateStore(
		normaliser,
		svc.logger,
	); err != nil {
		return err
	} else {
		svc.stateStore = s
	}

	if f, err := wrapFS.NewDocumentStoreWrapperFS(svc.stateStore.DocumentStore(), baseFS); err != nil {
		return err
	} else {
		svc.sessionFS = f
	}
	svc.logger.Printf("Wrapped the file system with a document store")

	if q, err := dispatchImpl.NewQueue(50, rpcServer, svc.logger); err != nil {
		return err
	} else {
		svc.clientNotifyQueue = q
	}

	if q, err := lintImpl.NewLintQueue(
		1,
		rootUri,
		svc.sessionFS,
		svc.clientNotifyQueue,
		svc.logger,
	); err != nil {
		return err
	} else {
		svc.lintQueue = q
	}
	svc.logger.Printf("Created new lint queue: %s", svc.lintQueue.Name())

	return nil
}

// Assigner builds out the jrpc2.Map according to the LSP protocol
// and passes related dependencies to handlers via context
func (svc *service) Assigner() (jrpc2.Assigner, error) {
	svc.logger.Println("Preparing new session ...")

	clientSession := session.NewSession()

	m := rpch.Map{
		"initialize": func(ctx context.Context, req *jrpc2.Request) (any, error) {
			ctx = ictx.WithClientCapabilities(ctx, clientSession.ClientCapabilities)
			return handle(ctx, req, svc.Initialize)
		},
		"initialized": func(ctx context.Context, req *jrpc2.Request) (any, error) {
			clientSession.Ready = true

			ctx = ictx.WithSentinelVersion(ctx, clientSession.SentinelVersion)
			svc.clientNotifyQueue.StartAsync(context.Background())
			svc.lintQueue.StartAsync(context.Background())

			return handle(ctx, req, Initialized)
		},

		"textDocument/didChange": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			if !clientSession.Ready {
				return nil, newClientNotReadyError()
			}

			ctx = ictx.WithDocumentStore(ctx, svc.stateStore.DocumentStore())
			ctx = ictx.WithSentinelVersion(ctx, clientSession.SentinelVersion)
			ctx = ictx.WithLintQueue(ctx, svc.lintQueue)

			return handle(ctx, req, svc.TextDocumentDidChange)
		},
		"textDocument/didOpen": func(ctx context.Context, req *jrpc2.Request) (any, error) {
			if !clientSession.Ready {
				return nil, newClientNotReadyError()
			}

			ctx = ictx.WithDocumentStore(ctx, svc.stateStore.DocumentStore())
			ctx = ictx.WithSentinelVersion(ctx, clientSession.SentinelVersion)
			ctx = ictx.WithLintQueue(ctx, svc.lintQueue)

			return handle(ctx, req, svc.TextDocumentDidOpen)
		},
		"textDocument/didClose": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			return nil, nil
		},
		"textDocument/didSave": func(ctx context.Context, req *jrpc2.Request) (interface{}, error) {
			return nil, nil
		},

		// Unfortunately the ServerCapabilities serialization doesn't turn these off. So for now
		// Just ignore them and return a noop respsonse
		"textDocument/codeLens": func(ctx context.Context, req *jrpc2.Request) (any, error) {
			return make([]lsp.CodeLens, 0), nil
		},
		"textDocument/documentLink": func(ctx context.Context, req *jrpc2.Request) (any, error) {
			return make([]lsp.DocumentLink, 0), nil
		},
		"textDocument/completion": func(ctx context.Context, req *jrpc2.Request) (any, error) {
			return make([]lsp.CompletionItem, 0), nil
		},

		"$/setTrace": func(ctx context.Context, req *jrpc2.Request) (any, error) {
			return nil, nil // TODO: Ignore these for now
		},
		"$/cancelRequest": func(ctx context.Context, req *jrpc2.Request) (any, error) {
			return nil, nil // TODO: Ignore these for now
		},

		"shutdown": func(ctx context.Context, req *jrpc2.Request) (any, error) {
			svc.shutdown()
			return nil, nil
		},

		// Custom messages

		// Set Sentinel Version
		lsp.SetSentinelVersionCommand: func(ctx context.Context, req *jrpc2.Request) (any, error) {
			if !clientSession.Ready {
				return nil, newClientNotReadyError()
			}

			ctx = ictx.WithSentinelVersion(ctx, clientSession.SentinelVersion)
			ctx = ictx.WithLintQueue(ctx, svc.lintQueue)

			return handle(ctx, req, svc.SentinelSetVersion)
		},
	}

	return m, nil
}

func (svc *service) Finish(_ jrpc2.Assigner, status jrpc2.ServerStatus) {
	if status.Closed || status.Err != nil {
		svc.logger.Printf("session stopped unexpectedly (err: %v)", status.Err)
	}

	svc.shutdown()
}

func (svc *service) shutdown() {
	svc.srvCtx.Done()
	svc.lintQueue.Stop()
}

func handle(ctx context.Context, req *jrpc2.Request, fn any) (any, error) {
	f := rpch.New(fn)
	result, err := f(ctx, req)
	if ctx.Err() != nil && errors.Is(ctx.Err(), context.Canceled) {
		err = fmt.Errorf("%w: %s", requestCancelledCode.Err(), err)
	}
	return result, err
}
