package concrete

import (
	"log"
	"sync"

	lsp "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/protocol"
	"github.com/glennsarti/sentinel-utils/lib/languageserver/internal/stores"
)

var _ stores.DocumentStore = &DocumentStore{}

func newDocumentStore(
	normaliser stores.UriNormaliserFunc,
	logger *log.Logger,
) *DocumentStore {
	return &DocumentStore{
		normaliser: normaliser,
		docs:       make(map[string]*stores.Document, 0),
		logger:     logger,
	}
}

type DocumentStore struct {
	muWrite    sync.RWMutex
	docs       map[string]*stores.Document
	normaliser stores.UriNormaliserFunc
	logger     *log.Logger
}

func (ds *DocumentStore) UpdateDocument(rawUri lsp.DocumentURI, version int, text []byte) error {
	uri := ds.normaliser(rawUri)

	ds.muWrite.Lock()
	defer ds.muWrite.Unlock()

	// Can't update documents that don't exist
	doc, err := ds.getDocumentUnsafe(uri)
	if err != nil {
		return err
	}

	// Can't update documents that are in the future
	if doc.Version > version {
		return stores.ErrDocumentWrongVersion
	}

	id := string(uri)
	ds.docs[id].Version = version
	ds.docs[id].Text = text

	return nil
}

func (ds *DocumentStore) SetDocument(rawUri lsp.DocumentURI, languageId string, version int, text []byte) error {
	uri := ds.normaliser(rawUri)

	ds.muWrite.Lock()
	defer ds.muWrite.Unlock()

	if d, err := ds.getDocumentUnsafe(uri); err == nil {
		d.Text = text
		d.Version = version
		return nil
	}

	id := string(uri)
	newDoc := &stores.Document{
		Id:          id,
		DocumentURI: uri,
		LanguageID:  languageId,
		Version:     version,
		Text:        text,
	}

	ds.docs[id] = newDoc
	return nil
}

func (ds *DocumentStore) GetDocument(rawUri lsp.DocumentURI) (*stores.Document, error) {
	uri := ds.normaliser(rawUri)

	ds.muWrite.RLock()
	defer ds.muWrite.RUnlock()

	if d, err := ds.getDocumentUnsafe(uri); err != nil {
		return nil, err
	} else {
		return &stores.Document{
			Id:          d.Id,
			DocumentURI: d.DocumentURI,
			LanguageID:  d.LanguageID,
			Version:     d.Version,
			Text:        d.Text,
		}, nil
	}
}

func (ds *DocumentStore) GetDocumentVersion(rawUri lsp.DocumentURI, version int) (*stores.Document, error) {
	uri := ds.normaliser(rawUri)

	ds.muWrite.RLock()
	defer ds.muWrite.RUnlock()

	if d, err := ds.getDocumentUnsafe(uri); err != nil {
		return nil, err
	} else if d.Version != version {
		return nil, stores.ErrDocumentWrongVersion
	} else {
		return &stores.Document{
			Id:          d.Id,
			DocumentURI: d.DocumentURI,
			LanguageID:  d.LanguageID,
			Version:     d.Version,
			Text:        d.Text,
		}, nil
	}
}

// Non-threadsafe operations below

func (ds *DocumentStore) getDocumentUnsafe(uri lsp.DocumentURI) (*stores.Document, error) {
	id := string(uri)
	if d, ok := ds.docs[id]; ok {
		return d, nil
	}
	return nil, stores.ErrDocumentNotExist
}
