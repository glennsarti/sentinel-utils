package os

import (
	"fmt"
	"io/fs"

	basefs "github.com/glennsarti/sentinel-utils/lib/filesystem"
	baseosfs "github.com/glennsarti/sentinel-utils/lib/filesystem/os"

	"github.com/glennsarti/sentinel-utils/lib/languageserver/internal/filesystem"
	lsp "github.com/glennsarti/sentinel-utils/lib/languageserver/internal/protocol"
)

func NewOSFileSystem(rootUri lsp.DocumentURI) (filesystem.SessionFS, error) {
	newFs := osFileSystem{}

	rootPath, err := newFs.UriToPath(rootUri)
	if err != nil {
		return nil, err
	}

	if f, err := baseosfs.NewOSFileSystem(rootPath); err != nil {
		return nil, err
	} else {
		newFs.baseFs = f
		newFs.root = rootPath
	}

	return newFs, nil
}

type osFileSystem struct {
	baseFs basefs.FS
	root   string
}

func (ofs osFileSystem) Name() string {
	return fmt.Sprintf("os (%q)", ofs.root)
}

func (ofs osFileSystem) Open(name string) (fs.File, error) {
	return ofs.baseFs.Open(name)
}

func (ofs osFileSystem) Stat(name string) (fs.FileInfo, error) {
	return ofs.baseFs.Stat(name)
}

func (ofs osFileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	return ofs.baseFs.ReadDir(name)
}

func (ofs osFileSystem) ReadFile(name string) ([]byte, error) {
	return ofs.baseFs.ReadFile(name)
}

func (ofs osFileSystem) PathJoin(elem ...string) string {
	return ofs.baseFs.PathJoin(elem...)
}

func (ofs osFileSystem) ParentPath(item string) string {
	return ofs.baseFs.ParentPath(item)
}

func (ofs osFileSystem) BasePath(item string) string {
	return ofs.baseFs.BasePath(item)
}
