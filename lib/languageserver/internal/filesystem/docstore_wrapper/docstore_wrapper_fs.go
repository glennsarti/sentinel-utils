package docstorewrapper

import (
	"errors"
	"fmt"
	"io/fs"
	"time"

	"github.com/glennsarti/sentinel-utils/lib/languageserver/internal/filesystem"
	lsp "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/protocol"
	"github.com/glennsarti/sentinel-utils/lib/languageserver/internal/stores"
)

func NewDocumentStoreWrapperFS(docstore stores.DocumentStore, baseFs filesystem.SessionFS) (filesystem.SessionFS, error) {
	if docstore == nil {
		return nil, fmt.Errorf("missing document store")
	}
	if baseFs == nil {
		return nil, fmt.Errorf("missing filesystem")
	}

	return wrappedFileSystem{
		baseFs:   baseFs,
		docstore: docstore,
	}, nil
}

type wrappedFileSystem struct {
	baseFs   filesystem.SessionFS
	docstore stores.DocumentStore
}

func (dsw wrappedFileSystem) Name() string {
	return fmt.Sprintf("docstore wrapping %s", dsw.baseFs.Name())
}

func (dsw wrappedFileSystem) Open(name string) (fs.File, error) {
	return nil, errors.New("not implemented")
}

func (dsw wrappedFileSystem) Stat(name string) (fs.FileInfo, error) {
	if uri, err := dsw.PathToUri(name); err == nil {
		if doc, err := dsw.docstore.GetDocument(uri); err == nil {
			return fauxFileInfo{
				name:    dsw.BasePath(name),
				size:    int64(len(doc.Text)),
				fmode:   fs.FileMode(0777),
				modTime: time.Now(),
				isDir:   false,
			}, nil
		}
	}
	return dsw.baseFs.Stat(name)
}

func (dsw wrappedFileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	// TODO: Won't include new items in the directory
	return dsw.baseFs.ReadDir(name)
}

func (dsw wrappedFileSystem) ReadFile(name string) ([]byte, error) {
	// Use the document store if it exists.
	if uri, err := dsw.PathToUri(name); err == nil {
		if doc, err := dsw.docstore.GetDocument(uri); err == nil {
			return doc.Text, nil
		}
	}

	return dsw.baseFs.ReadFile(name)
}

func (dsw wrappedFileSystem) PathJoin(elem ...string) string {
	return dsw.baseFs.PathJoin(elem...)
}

func (dsw wrappedFileSystem) ParentPath(item string) string {
	return dsw.baseFs.ParentPath(item)
}

func (dsw wrappedFileSystem) BasePath(item string) string {
	return dsw.baseFs.BasePath(item)
}

func (dsw wrappedFileSystem) UriToPath(uri lsp.DocumentURI) (string, error) {
	return dsw.baseFs.UriToPath(uri)
}

func (dsw wrappedFileSystem) PathToUri(path string) (lsp.DocumentURI, error) {
	return dsw.baseFs.PathToUri(path)
}
