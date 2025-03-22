package txtar_fs

import (
	"cmp"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/glennsarti/sentinel-utils/lib/filesystem"

	"golang.org/x/tools/txtar"
)

var _ fs.FS = &txtarFileSystem{}

func NewTxtarFileSystem(archive *txtar.Archive) filesystem.FS {

	newFS := &txtarFileSystem{
		files:      make(map[string]txtar.File, 0),
		modTime:    time.Now(),
		roFileMode: 0o0111,
	}

	for _, f := range archive.Files {
		resolvedPath := "/" + newFS.resolvePath(f.Name)
		newFS.files[resolvedPath] = f
	}

	return newFS
}

type txtarFileSystem struct {
	files      map[string]txtar.File
	modTime    time.Time
	roFileMode uint32
}

func (d *txtarFileSystem) resolvePath(name string) string {
	return path.Clean(name)
}

func (d *txtarFileSystem) Open(name string) (fs.File, error) {
	return nil, fmt.Errorf("not implemented")
}

func (d *txtarFileSystem) Stat(name string) (fs.FileInfo, error) {
	// Special case the root
	if name == "/" || name == "." {
		return fileInfo{
			name:    "/",
			size:    0,
			mode:    fs.ModeDir + fs.FileMode(d.roFileMode),
			modTime: d.modTime,
			isDir:   true,
			raw:     nil,
		}, nil
	}
	actualName := d.resolvePath(name)

	// Is it a specific file?
	if file, ok := d.files[actualName]; ok {
		return fileInfo{
			name:    actualName,
			size:    int64(len(file.Data)),
			mode:    fs.FileMode(d.roFileMode),
			modTime: d.modTime,
			isDir:   false,
			raw:     nil,
		}, nil
	}

	// Is it a directory?
	for filename := range d.files {
		// Is there a file that is in, or in a sub directory of the requested path
		if strings.HasPrefix(filename, actualName+"/") {
			return fileInfo{
				name:    actualName,
				size:    0,
				mode:    fs.FileMode(d.roFileMode),
				modTime: d.modTime,
				isDir:   false,
				raw:     nil,
			}, nil
		}
	}

	return nil, fs.ErrNotExist
}

func cmpDirEntry(a, b fs.DirEntry) int {
	return cmp.Compare(a.Name(), b.Name())
}

func (d *txtarFileSystem) ReadDir(name string) ([]fs.DirEntry, error) {
	entries := make(map[string]fs.DirEntry, 0)
	actualDir := d.resolvePath(name)
	// Special case the root as we add a trailing "/" when comparing paths later
	if actualDir == "/" {
		actualDir = ""
	}

	for filename, file := range d.files {
		if strings.HasPrefix(filename, actualDir+"/") {
			subName := strings.TrimPrefix(filename, actualDir+"/")
			if !strings.Contains(subName, "/") {
				if _, ok := entries[subName]; !ok {
					entries[subName] = dirEntry{
						name:  subName,
						isDir: false,
						mode:  fs.FileMode(d.roFileMode),
						fInfo: fileInfo{
							name:    subName,
							size:    int64(len(file.Data)),
							mode:    fs.FileMode(d.roFileMode),
							modTime: d.modTime,
							isDir:   false,
							raw:     nil,
						},
					}
				}
			} else {
				parts := strings.SplitN(subName, "/", 2)
				if _, ok := entries[parts[0]]; !ok {
					entries[parts[0]] = dirEntry{
						name:  parts[0],
						isDir: true,
						mode:  fs.FileMode(d.roFileMode),
						fInfo: fileInfo{
							name:    subName,
							size:    0,
							mode:    fs.FileMode(d.roFileMode),
							modTime: d.modTime,
							isDir:   true,
							raw:     nil,
						},
					}
				}
			}
		}
	}

	sortedEntries := make([]fs.DirEntry, 0)
	for _, entry := range entries {
		sortedEntries = append(sortedEntries, entry)
	}
	slices.SortFunc(sortedEntries, cmpDirEntry)

	return sortedEntries, nil
}

func (d *txtarFileSystem) ReadFile(name string) ([]byte, error) {
	actualName := d.resolvePath(name)

	if file, ok := d.files[actualName]; ok {
		return file.Data, nil
	}

	return nil, fs.ErrNotExist
}

func (d *txtarFileSystem) PathJoin(elem ...string) string {
	return path.Join(elem...)
}

func (d *txtarFileSystem) BasePath(item string) string {
	return path.Base(item)
}

func (d *txtarFileSystem) ParentPath(item string) string {
	return path.Dir(item)
}
