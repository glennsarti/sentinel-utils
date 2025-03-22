package languageserver

import (
	"context"
	"log"

	"github.com/creachadair/jrpc2"
	"github.com/glennsarti/sentinel-utils/lib/languageserver/internal"
	"github.com/glennsarti/sentinel-utils/lib/languageserver/transports"

	"github.com/creachadair/jrpc2/server"
	"github.com/glennsarti/sentinel-utils/lib/languageserver/internal/langserver"
)

type langServer struct {
	ctx        context.Context
	logger     *log.Logger
	srvOptions *jrpc2.ServerOptions
}

func NewLangServer(logger *log.Logger, ctx context.Context) *langServer {
	opts := &jrpc2.ServerOptions{
		AllowPush:   true,
		Concurrency: 0,
		Logger:      jrpc2.StdLogger(logger),
		RPCLog:      internal.NewRpcLogger(logger),
	}

	return &langServer{
		ctx:        ctx,
		logger:     logger,
		srvOptions: opts,
	}
}

func (ls *langServer) newService() server.Service {
	svc := langserver.NewService(ls.logger, ls.ctx)
	return svc
}

func (ls *langServer) StartAndWait(t transports.Transport) error {
	accepter, err := t.Accepter()
	if err != nil {
		return err
	}

	go func() {
		ls.logger.Println("Starting loop server ...")
		err = server.Loop(ls.ctx, accepter, ls.newService, &server.LoopOptions{
			ServerOptions: ls.srvOptions,
		})
		if err != nil {
			ls.logger.Printf("Loop server failed to start: %s", err)
		}
	}()

	err = t.StartAndWait(ls.ctx)
	return err
}
