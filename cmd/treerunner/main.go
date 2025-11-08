package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/emicklei/gi"
)

var dir = flag.String("dir", ".", "directory to run gi on")
var dry = flag.Bool("dry", false, "dry run - do not run")

func main() {
	flag.Parse()

	success, failed := 0, 0

	filepath.WalkDir(*dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() {
			if safeRun(path) {
				success++
			} else {
				failed++
			}
		}
		return nil
	})
	fmt.Printf("summary: %d succeeded, %d failed\n", success, failed)
}

func safeRun(path string) (ok bool) {
	fmt.Println("gi run", path)
	if *dry {
		return true
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "[treerunner] recovered from panic: %v\n", r)
			ok = false
		}
	}()
	gi.Run(path)
	ok = true
	return
}
