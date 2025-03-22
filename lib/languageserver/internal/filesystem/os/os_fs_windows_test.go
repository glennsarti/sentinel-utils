//go:build windows
// +build windows

package os

import (
	"testing"

	lsp "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/protocol"
)

func TestUriToPath(t *testing.T) {
	subject := osFileSystem{}

	for _, testcase := range []struct {
		uri, expectPath string
	}{
		{
			uri:        ``,
			expectPath: ``,
		},
		{
			uri:        `file:///c:/Something/Foo%20Bar/baz.go`,
			expectPath: `C:\Something\Foo Bar\baz.go`,
		},
	} {
		inputUri := lsp.DocumentURI(testcase.uri)

		actualPath, err := subject.UriToPath(inputUri)
		if err != nil {
			t.Errorf("failed to convert URI (%q) to a path: %v", testcase.uri, err)
		}
		if actualPath != testcase.expectPath {
			t.Errorf("URI %q expected path %q, got %q", testcase.uri, testcase.expectPath, actualPath)
		}
	}
}

func TestPathToUri(t *testing.T) {
	subject := osFileSystem{}

	for _, testcase := range []struct {
		path      string
		expectUri lsp.DocumentURI
	}{
		{
			path:      ``,
			expectUri: lsp.DocumentURI(``),
		},
		{
			path:      `C:\Something\Foo Bar\baz.go`,
			expectUri: lsp.DocumentURI(`file:///C:/Something/Foo%20Bar/baz.go`),
		},
	} {
		actualUri, err := subject.PathToUri(testcase.path)
		if err != nil {
			t.Errorf("failed to convert path (%q) to a uri: %v", testcase.path, err)
		}
		if actualUri != testcase.expectUri {
			t.Errorf("Path %q expected uri %q, got %q", testcase.path, testcase.expectUri, actualUri)
		}
	}
}
