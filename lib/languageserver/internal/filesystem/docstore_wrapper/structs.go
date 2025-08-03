package docstorewrapper

import (
	"io/fs"
	"time"
)

var _ fs.FileInfo = fauxFileInfo{}

type fauxFileInfo struct {
	name    string
	size    int64
	fmode   fs.FileMode
	modTime time.Time
	isDir   bool
}

func (ffi fauxFileInfo) Name() string       { return ffi.name }
func (ffi fauxFileInfo) Size() int64        { return ffi.size }
func (ffi fauxFileInfo) Mode() fs.FileMode  { return ffi.fmode }
func (ffi fauxFileInfo) ModTime() time.Time { return ffi.modTime }
func (ffi fauxFileInfo) IsDir() bool        { return ffi.isDir }
func (ffi fauxFileInfo) Sys() any           { return nil }
