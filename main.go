package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"github.com/bastengao/gncdu/config"
	"github.com/bastengao/gncdu/debug"
	"github.com/bastengao/gncdu/scan"
	"github.com/bastengao/gncdu/ui"
)

var cFlag = flag.Int("c", scan.DefaultConcurrency(), "the number of concurrent scanners, default is number of CPU")
var leFlag = flag.Bool("log", false, "record deletions and other activity")
var helpFlag = flag.Bool("help", false, "help")
var loadFlag = flag.String("load", "", "Load FileData from saved JSON file")

func main() {
	flag.Parse()

	config.EnableLog = *leFlag
	if helpFlag != nil && *helpFlag {
		fmt.Printf("gncdu %s\n", ui.Version)
		flag.Usage()
		return
	}

	if *loadFlag != "" {
		debug.Info(fmt.Sprintf("Loading from %s", *loadFlag))

		ui.ShowUI(func() ([]*scan.FileData, error) {
			files, err := scan.LoadFromJSON(*loadFlag)
			if err != nil {
				return nil, fmt.Errorf("failed to load JSON: %v", err)
			}
			return files, nil
		})
		return
	}

	dir := "."
	args := flag.Args()
	if len(args) > 0 {
		dir = args[0]
	}
	dir, err := filepath.Abs(dir)
	debug.Info(fmt.Sprintf("Scanning %s", dir))
	if err != nil {
		fmt.Println(err)
		return
	}

	ui.ShowUI(func() ([]*scan.FileData, error) {
		files, err := scan.ScanDirConcurrent(dir, *cFlag)

		if err != nil {
			return nil, err
		}

		return files, nil
	})
}
