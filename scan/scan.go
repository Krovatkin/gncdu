package scan

import (
	"encoding/json"
	"io/ioutil"
	"runtime"
	"sync"
)

func ScanDirConcurrent(dir string, concurrency int) ([]*FileData, error) {
	root := newRootFileData(dir)

	if concurrency == 0 {
		concurrency = DefaultConcurrency()
	}

	ch := make(chan *FileData)
	closeWait := &sync.WaitGroup{}

	var wait sync.WaitGroup
	wait.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			for file := range ch {
				scanDir(file, ch, closeWait)
				closeWait.Done()
			}
			wait.Done()
		}()
	}

	err := scanDir(root, ch, closeWait)
	if err != nil {
		return nil, err
	}

	go func() {
		closeWait.Wait()
		close(ch)
	}()

	wait.Wait()

	return root.Children, nil
}

func DefaultConcurrency() int {
	maxProcs := runtime.GOMAXPROCS(0)
	numCPU := runtime.NumCPU()
	if maxProcs < numCPU {
		return maxProcs
	}

	return numCPU
}

func scanDir(parent *FileData, ch chan *FileData, closeWait *sync.WaitGroup) error {
	if !parent.Root() && (parent.JsonSize != -1 || !parent.IsDir) {
		return nil
	}

	files, err := ioutil.ReadDir(parent.Path())
	if err != nil {
		return err
	}

	children := []*FileData{}
	closeWait.Add(len(files))
	for _, file := range files {
		f := newFileData(parent, file)
		go func() {
			ch <- f
		}()

		children = append(children, f)
	}

	parent.Children = children
	return nil
}

// LoadFromJSON loads FileData from JSON file
func LoadFromJSON(filename string) ([]*FileData, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var root *FileData
	err = json.Unmarshal(data, &root)
	if err != nil {
		return nil, err
	}

	// Set parent references for all root files and their children
	for _, file := range root.Children {
		file.Parent = root
		file.SetParents()
	}

	return root.Children, nil
}
