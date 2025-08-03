package queues

import (
	"context"
	"log"
)

type LintQueueRequest struct {
	DocId           string
	DocVersion      int
	SentinelVersion string
}

type LintQueue interface {
	Enqueue(req LintQueueRequest) error

	Start(context.Context) error
	StartAsync(context.Context) error
	Stop()
	Name() string
	Logger() *log.Logger
}
