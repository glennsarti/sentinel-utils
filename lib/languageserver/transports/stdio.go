package transports

import (
	"context"
	"log"
	"os"

	"github.com/creachadair/jrpc2/channel"
	"github.com/creachadair/jrpc2/server"
)

func NewSTDIOTransport(logger *log.Logger) Transport {
	return &stdioTransport{
		logger: logger,
	}
}

type stdioTransport struct {
	accepter server.Accepter
	logger   *log.Logger
}

func (t *stdioTransport) Accepter() (server.Accepter, error) {
	if t.accepter != nil {
		return t.accepter, nil
	}

	t.accepter = &stdioAccepter{}

	return t.accepter, nil
}

type stdioAccepter struct {
	stdioCh channel.Channel
}

var _ server.Accepter = &stdioAccepter{}

func (sa *stdioAccepter) Accept(ctx context.Context) (channel.Channel, error) {
	if sa.stdioCh == nil {
		sa.stdioCh = channel.LSP(os.Stdin, os.Stdout)
		return sa.stdioCh, nil
	}

	<-ctx.Done()
	return nil, nil
}

func (t *stdioTransport) StartAndWait(ctx context.Context) error {
	<-ctx.Done()

	t.logger.Println("STDIO server stopped.")
	return nil
}
