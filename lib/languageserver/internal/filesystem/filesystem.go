package filesystem

import (
	"github.com/glennsarti/sentinel-utils/lib/filesystem"
	lsp "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/protocol"
)

type SessionFS interface {
	filesystem.FS

	Name() string

	// Converts an LSP Document URI into a file system path.
	UriToPath(uri lsp.DocumentURI) (string, error)

	// Converts a file system path into an LSP Document URI.
	PathToUri(path string) (lsp.DocumentURI, error)
}
