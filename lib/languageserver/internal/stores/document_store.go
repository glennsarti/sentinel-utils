package stores

import (
	"errors"

	lsp "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/protocol"
)

type UriNormaliserFunc func(lsp.DocumentURI) lsp.DocumentURI

type DocumentStore interface {
	UpdateDocument(uri lsp.DocumentURI, version int, text []byte) error
	SetDocument(uri lsp.DocumentURI, languageId string, version int, text []byte) error

	GetDocument(uri lsp.DocumentURI) (*Document, error)
	GetDocumentVersion(uri lsp.DocumentURI, version int) (*Document, error)
}

type Document struct {
	Id string

	DocumentURI lsp.DocumentURI
	LanguageID  string
	Version     int

	Text []byte
}

var (
	ErrDocumentNotExist     = errors.New("document does not exist")
	ErrDocumentWrongVersion = errors.New("document is not the correct version")
)
