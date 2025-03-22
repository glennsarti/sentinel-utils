package os

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"unicode"

	lsp "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/protocol"
)

func (dsw osFileSystem) UriToPath(uri lsp.DocumentURI) (string, error) {
	if uri == "" {
		return "", nil
	}

	u, err := url.ParseRequestURI(string(uri))
	if err != nil {
		return "", err
	}

	// This filesystem only accepts URIs with the file:// scheme
	if u.Scheme != `file` {
		return "", fmt.Errorf("documentURI scheme is not invalid got %q from %q", u.Scheme, uri)
	}

	if isWindowsUriPath(u.Path) {
		// Trim the leading slash and uppercase Windows drive letters
		u.Path = strings.ToUpper(string(u.Path[1])) + u.Path[2:]
	}

	return filepath.FromSlash(u.Path), nil
}

func (dsw osFileSystem) PathToUri(path string) (lsp.DocumentURI, error) {
	if path == "" {
		return lsp.DocumentURI(""), nil
	}

	path = filepath.Clean(path)

	if !isWindowsDrivePath(path) {
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
	}

	if isWindowsDrivePath(path) {
		path = "/" + strings.ToUpper(string(path[0])) + path[1:]
	}

	u := url.URL{
		Scheme: `file`,
		Path:   filepath.ToSlash(path),
	}
	return lsp.DocumentURI(u.String()), nil
}

func isWindowsUriPath(path string) bool {
	if len(path) < 4 {
		return false
	}
	return path[0] == '/' && unicode.IsLetter(rune(path[1])) && path[2] == ':'
}

func isWindowsDrivePath(path string) bool {
	if len(path) < 3 {
		return false
	}
	return unicode.IsLetter(rune(path[0])) && path[1] == ':'
}
