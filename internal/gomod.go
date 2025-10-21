package internal

import (
	"fmt"
	"os"

	"golang.org/x/mod/modfile"
)

func LoadGoMod(filename string) (*modfile.File, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", filename, err)
	}
	modFile, err := modfile.Parse(filename, data, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", filename, err)
	}
	return modFile, nil
}
