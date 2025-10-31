package internal

import (
	"fmt"
	"os"
	"path"

	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
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

// https://stackoverflow.com/questions/67211875/how-to-get-the-path-to-a-go-module-dependency
func GetModulePath(name, version string) (string, error) {
	// first we need GOMODCACHE
	cache, ok := os.LookupEnv("GOMODCACHE")
	if !ok {
		cache = path.Join(os.Getenv("GOPATH"), "pkg", "mod")
	}

	// then we need to escape path
	escapedPath, err := module.EscapePath(name)
	if err != nil {
		return "", err
	}

	// version also
	escapedVersion, err := module.EscapeVersion(version)
	if err != nil {
		return "", err
	}

	return path.Join(cache, escapedPath+"@"+escapedVersion), nil
}
