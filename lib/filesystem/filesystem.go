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

	PathJoin(...string) string
	ParentPath(string) string
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
