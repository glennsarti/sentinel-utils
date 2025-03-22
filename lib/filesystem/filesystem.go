package filesystem

import (
	"io/fs"

	"github.com/glennsarti/sentinel-parser/filetypes"
)

type FS interface {
	fs.FS
	fs.ReadDirFS
	fs.ReadFileFS
	fs.StatFS

	// Joins any number of path elements into a single path,
	// separating them with a filesystem specific separator
	PathJoin(...string) string

	// Returns all but the last element of path, typically the path's directory.
	// Trailing slashes are removed.
	// If the path is empty, ParentPath returns ".".
	// If the path consists entirely of separators, ParentPath returns a single separator.
	ParentPath(string) string

	// BasePath returns the last element of path.
	// Trailing path separators are removed before extracting the last element.
	// If the path is empty, BasePath returns ".".
	// If the path consists entirely of separators, BasePath returns a single separator.
	BasePath(string) string
}

type File struct {
	Path    string
	Name    string
	Type    filetypes.FileType
	ID      string
	Content *[]byte
}

func (f File) String() string {
	switch f.Type {
	case filetypes.ConfigOverrideFileType:
		return "configuration override"
	case filetypes.ConfigPrimaryFileType:
		return "configuration"
	case filetypes.ConfigTestFileType:
		return "test"
	case filetypes.ModuleFileType:
		return "module"
	case filetypes.PolicyFileType:
		return "policy"
	default:
		return "unknown"
	}
}
