package client_dispatch

import (
	"context"
	"log"

	"github.com/glennsarti/sentinel-utils/lib/languageserver/internal/queues"
	"github.com/glennsarti/sentinel-utils/lib/languageserver/internal/queues/generic"
)

var _ queues.ClientNotifyDispatchQueue = &dipatchQueue{}

type ClientNotifier interface {
	Notify(ctx context.Context, method string, params interface{}) error
}

func NewQueue(
	queueSize int,
	disptacher ClientNotifier,
	logger *log.Logger,
) (queues.ClientNotifyDispatchQueue, error) {
	dq := &dipatchQueue{
		disptacher: disptacher,
		logger:     logger,
	}
	dq.baseq = generic.NewGenericQueue(1, queueSize, dq.process)

	return dq, nil
}

type dipatchQueue struct {
	logger     *log.Logger
	baseq      *generic.GenericQueue[queues.ClientNotifyDispatchRequest]
	disptacher ClientNotifier
}

func (dq *dipatchQueue) Enqueue(req queues.ClientNotifyDispatchRequest) error {
	dq.baseq.Enqueue(req)
	return nil
}
func (dq *dipatchQueue) Start(ctx context.Context) error      { return dq.baseq.Start(ctx) }
func (dq *dipatchQueue) StartAsync(ctx context.Context) error { return dq.baseq.StartAsync(ctx) }
func (dq *dipatchQueue) Stop()                                {}
func (dq *dipatchQueue) Name() string                         { return "clientNotifyDisptachQueue" }
func (dq *dipatchQueue) Logger() *log.Logger                  { return dq.logger }

func (dq *dipatchQueue) process(job queues.ClientNotifyDispatchRequest) error {
	return dq.disptacher.Notify(context.TODO(), job.Method, job.Params)
}
