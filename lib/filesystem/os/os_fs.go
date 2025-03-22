package os

import (
	"io/fs"
	"os"
	"path/filepath"

	"github.com/glennsarti/sentinel-utils/lib/filesystem"
)

func NewOSFileSystem(root string) (filesystem.FS, error) {
	return &osFileSystem{
		FS: os.DirFS(root),
	}, nil
}

type osFileSystem struct {
	fs.FS
}

func (d osFileSystem) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (d osFileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

func (d osFileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

func (d osFileSystem) PathJoin(elem ...string) string {
	return filepath.Join(elem...)
}

func (d osFileSystem) ParentPath(item string) string {
	return filepath.Dir(item)
}

func (d osFileSystem) BasePath(item string) string {
	return filepath.Base(item)
}
