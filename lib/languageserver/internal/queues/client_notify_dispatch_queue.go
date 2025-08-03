package queues

import (
	"context"
	"log"
)

type ClientNotifyDispatchRequest struct {
	Method string
	Params any
}

type ClientNotifyDispatchQueue interface {
	Enqueue(req ClientNotifyDispatchRequest) error

	Start(context.Context) error
	StartAsync(context.Context) error
	Stop()
	Name() string
	Logger() *log.Logger
}
