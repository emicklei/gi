package main

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	goroot := build.Default.GOROOT
	if goroot == "" {
		fmt.Println("GOROOT is not set. Please set GOROOT or run with a Go installation in PATH.")
		os.Exit(1)
	}

	srcDir := filepath.Join(goroot, "src")

	// fmt.Printf("Searching for package paths in: %s\n", srcDir)

	err := filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Skip special directories
			dirName := info.Name()
			if strings.HasPrefix(dirName, ".") || strings.HasPrefix(dirName, "_") || dirName == "vendor" {
				return filepath.SkipDir
			}

			// Check if this directory contains any .go files (excluding _test.go)
			hasGoFiles := false
			entries, err := os.ReadDir(path)
			if err != nil {
				return err
			}
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".go") && !strings.HasSuffix(entry.Name(), "_test.go") {
					hasGoFiles = true
					break
				}
			}

			if hasGoFiles {
				// Calculate the package path relative to GOROOT/src
				relPath, err := filepath.Rel(srcDir, path)
				if err != nil {
					return err
				}
				// Use forward slashes for package paths
				packagePath := filepath.ToSlash(relPath)
				if !strings.HasPrefix(packagePath, "cmd/") {
					if !strings.Contains(packagePath, "internal") {
						fmt.Printf(`"%s",`+"\n", packagePath)
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		fmt.Printf("Error walking the directory: %v\n", err)
	}
}
