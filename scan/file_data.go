package scan

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	cp "github.com/otiai10/copy"
)

type FileData struct {
	Parent   *FileData   `json:"-"` // Exclude parent (no cycles)
	dir      string      `json:"dir"`
	Name     string      `json:"name"`
	size     int64       `json:"size"`
	IsDir    bool        `json:"isDir"`
	Children []*FileData `json:"children,omitempty"`
	count    int         `json:"count"`
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
	return &FileData{Parent: parant, dir: parant.Path(), Name: file.Name(), size: size, IsDir: file.IsDir(), count: count}
}

func (d FileData) Root() bool {
	return d.Name == ""
}

func (d FileData) Label() string {
	if d.Root() {
		return "/.."
	}

	if d.IsDir {
		return d.Name + "/"
	}

	return d.Name
}

func (d FileData) Path() string {
	if d.Root() {
		return d.dir
	}

	return filepath.Join(d.dir, d.Name)
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

	var s int64 = 0

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

func (fd *FileData) Json() string {
	// pre-populate sizes and count
	fd.Size()
	fd.Count()
	jsonData, err := json.MarshalIndent(fd, "", "\t")
	if err != nil {
		return ""
	}
	return string(jsonData)
}

func (d *FileData) Move(dstDirectoryPath string) error {
	dstFilePath := filepath.Join(dstDirectoryPath, d.Name)

	err := cp.Copy(d.Path(), dstFilePath)
	if err != nil {
		return err
	}

	d.updateSizesOnMove(dstDirectoryPath)
	return nil
}

// SetParents recursively sets parent references for all children
func (fd *FileData) SetParents() {
	for _, child := range fd.Children {
		child.Parent = fd
		child.SetParents() // Recursive call on children
	}
}

func (file *FileData) SubtractSizeFromAncestors() {
	parent := file.Parent
	for parent != nil {
		file.size -= parent.size
		parent = parent.Parent // Move up in the hierarchy
	}
}

func (file *FileData) updateSizesOnMove(dst string) {
	// Subtracting size from ancestors
	file.SubtractSizeFromAncestors()

	// Getting the root
	root := file
	for root.Parent != nil {
		root = root.Parent
	}

	if !strings.HasPrefix(dst, root.Path()) {
		// If it isn't, then do nothing
		return
	}

	// Removing prefix from dst
	dst = strings.TrimPrefix(dst, root.Path())

	// Removing leading slash, if present
	dst = strings.TrimPrefix(dst, string(os.PathSeparator))

	// Splitting dst by path separator
	paths := strings.Split(dst, string(os.PathSeparator))

	// Update root node size first
	root.size += file.size

	// Iterating over elements
	for _, path := range paths {
		// Find the child of current iter starting from root
		var child *FileData
		for _, child = range root.Children {
			if child.Name == path {
				break
			}
		}

		// If child was found then increase the size by file.Size()
		if child != nil {
			child.size += file.size
			root = child // Make this node as new root for the next iteration
		} else {
			// Break upon no matching child
			break
		}
	}
}
