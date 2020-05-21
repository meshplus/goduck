package repo

import (
	"fmt"

	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/mitchellh/go-homedir"
)

const (
	// DefaultPathName is the default config dir name
	DefaultPathName = ".goduck"

	// DefaultPathRoot is the path to the default config dir location.
	DefaultPathRoot = "~/" + DefaultPathName
)

var RootPath string

// PathRoot returns root path (default .pier)
func PathRoot() (string, error) {
	if RootPath != "" {
		return RootPath, nil
	}

	return homedir.Expand(DefaultPathRoot)
}

// PathRootWithDefault gets current config path with default value
func PathRootWithDefault(path string) (string, error) {
	var err error
	if len(path) == 0 {
		path, err = PathRoot()
		if err != nil {
			return "", err
		}
	}

	if !fileutil.Exist(path) {
		return "", fmt.Errorf("please run `goduck --repo=%s init` first", path)
	}

	return path, nil
}
