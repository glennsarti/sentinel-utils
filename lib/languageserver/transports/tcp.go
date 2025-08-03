package transports

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/creachadair/jrpc2/channel"
	"github.com/creachadair/jrpc2/server"
)

func NewTCPTransport(address string, logger *log.Logger) Transport {
	return &tcpTransport{
		address: address,
		logger:  logger,
	}
}

type tcpTransport struct {
	address       string
	actualAddress string
	listener      net.Listener
	accepter      server.Accepter
	logger        *log.Logger
}

func (t *tcpTransport) Accepter() (server.Accepter, error) {
	if t.accepter != nil {
		return t.accepter, nil
	}

	t.logger.Printf("Starting TCP server at %q ...", t.address)
	if lst, err := net.Listen("tcp", t.address); err != nil {
		return nil, fmt.Errorf("TCP Server failed to start: %s", err)
	} else {
		t.listener = lst
	}
	t.actualAddress = t.listener.Addr().String()
	t.logger.Printf("TCP server running at %q", t.actualAddress)
	t.accepter = server.NetAccepter(t.listener, channel.LSP)

	return t.accepter, nil
}

func (t *tcpTransport) StartAndWait(ctx context.Context) error {
	<-ctx.Done()
	t.logger.Printf("Stopping TCP server %q ...", t.address)
	if err := t.listener.Close(); err != nil {
		t.logger.Printf("TCP server at %q failed to stop: %s", t.actualAddress, err)
		return err
	}

	t.logger.Printf("TCP server at %q stopped.", t.actualAddress)
	return nil
}
