package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/emicklei/gi"
)

var dir = flag.String("dir", ".", "directory to run gi on")
var dry = flag.Bool("dry", false, "dry run - do not run")
var report = flag.String("report", "treerunner-report.json", "generate report for the run summary")

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
	if *report != "" {
		generateReport(success, failed)
	}
}

func generateReport(success, failed int) {
	data, _ := json.Marshal(map[string]any{
		"name":    "gobyexample",
		"success": success,
		"failed":  failed,
		"label":   fmt.Sprintf("%d/%d", success, success+failed),
	})
	os.WriteFile(*report, data, 0644)
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
