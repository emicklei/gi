package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/emicklei/gi"
)

var dir = flag.String("dir", ".", "directory to run gi on")
var dry = flag.Bool("dry", false, "dry run - do not run")
var report = flag.String("report", "treerunner-report.json", "generate report for the run summary")
var skip = flag.String("skip", "", "comma-separated list of directories to skip")

func main() {
	flag.Parse()

	pathsToSkip := strings.Split(*skip, ",")

	wr := walkReport{
		Name: "treerunner",
		Runs: make(map[string]runReport),
	}

	filepath.WalkDir(*dir, func(path string, d fs.DirEntry, err error) error {
		if d.IsDir() && path != "." {
			rr := runReport{}
			if slices.Contains(pathsToSkip, path) {
				fmt.Printf("skipping directory: %s\n", path)
				rr.Skipped = true
				wr.Runs[path] = rr
				return nil
			}
			if msg, ok := safeRun(path); ok {
				wr.Success++
				rr.Pass = true
			} else {
				wr.Failed++
				rr.Pass = false
				rr.Error = msg
			}
			wr.Runs[path] = rr
		}
		return nil
	})
	fmt.Printf("summary: %d succeeded, %d failed\n", wr.Success, wr.Failed)
	if *report != "" {
		generateReport(wr)
	}
}

type walkReport struct {
	Name    string               `json:"name"`
	Success int                  `json:"success"`
	Failed  int                  `json:"failed"`
	Label   string               `json:"label"`
	Runs    map[string]runReport `json:"runs"`
}

type runReport struct {
	Pass    bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Skipped bool   `json:"skipped,omitempty"`
}

func generateReport(wr walkReport) {
	wr.Label = fmt.Sprintf("%d/%d", wr.Success, wr.Success+wr.Failed)
	data, _ := json.MarshalIndent(wr, "", "  ")
	os.WriteFile(*report, data, 0644)
}

func safeRun(path string) (msg string, ok bool) {
	fmt.Println("gi run", path)
	if *dry {
		return "dry run", true
	}
	defer func() {
		if r := recover(); r != nil {
			msg = fmt.Sprintf("panic: %v", r)
			fmt.Fprintf(os.Stderr, "[treerunner] recovered from panic: %v\n", r)
			ok = false
		}
	}()
	gi.Run(path)
	ok = true
	return
}
