package concrete

import (
	"log"

	"github.com/glennsarti/sentinel-utils/lib/languageserver/internal/stores"
)

var _ stores.StateStore = &stateStore{}

type stateStore struct {
	logger        *log.Logger
	documentStore stores.DocumentStore
}

func NewStateStore(
	normaliser stores.UriNormaliserFunc,
	logger *log.Logger,
) (*stateStore, error) {
	return &stateStore{
		logger:        logger,
		documentStore: newDocumentStore(normaliser, logger),
	}, nil
}

func (ds *stateStore) DocumentStore() stores.DocumentStore {
	return ds.documentStore
}
