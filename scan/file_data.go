package scan

import (
	cp "github.com/otiai10/copy"
	"os"
	"path/filepath"
)

type FileData struct {
	Parent   *FileData
	dir      string
	Info     os.FileInfo
	size     int64
	Children []*FileData
	count    int
}

func newRootFileData(dir string) *FileData {
	return &FileData{dir: dir, size: 0, count: 0}
}

func newFileData(parant *FileData, file os.FileInfo) *FileData {
	var size int64 = -1
	count := -1
	if !file.IsDir() {
		size = file.Size()
		count = 0
	}
	return &FileData{Parent: parant, dir: parant.Path(), Info: file, size: size, count: count}
}

func (d FileData) Root() bool {
	return d.Info == nil
}

func (d FileData) Label() string {
	if d.Root() {
		return "/.."
	}

	if d.Info.IsDir() {
		return d.Info.Name() + "/"
	}

	return d.Info.Name()
}

func (d FileData) Path() string {
	if d.Root() {
		return d.dir
	}

	return filepath.Join(d.dir, d.Info.Name())
}

func (d FileData) String() string {
	return d.Path()
}

func (d *FileData) Count() int {
	if d.count != -1 {
		return d.count
	}
	c := len(d.Children)
	for _, f := range d.Children {
		c += f.Count()
	}
	d.count = c
	return c
}

func (d *FileData) Size() int64 {
	if d.size != -1 {
		return d.size
	}

	var s int64
	if d.Info != nil {
		s += d.Info.Size()
	}
	for _, f := range d.Children {
		s += f.Size()
	}
	d.size = s
	return s
}

func (d *FileData) SetChildren(children []*FileData) {
	d.Children = children
	d.size = -1
	d.count = -1
	d.Size()
	d.Count()
}

func (d *FileData) Delete() error {
	return os.RemoveAll(d.Path())
}

func (d *FileData) Move(dstDirectoryPath string) error {
	dstFilePath := filepath.Join(dstDirectoryPath, d.Info.Name())
	return cp.Copy(d.Path(), dstFilePath)
}

func (file *FileData) SubtractSizeFromAncestors() {
	parent := file.Parent
	for parent != nil {
		file.size -= parent.size
		parent = parent.Parent // Move up in the hierarchy
	}
}

func hasDir(files []os.FileInfo) bool {
	for _, file := range files {
		if file.IsDir() {
			return true
		}
	}
	return false
}
