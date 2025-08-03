package transports

import (
	"context"

	"github.com/creachadair/jrpc2/server"
)

type Transport interface {
	Accepter() (server.Accepter, error)
	StartAndWait(ctx context.Context) error
}
