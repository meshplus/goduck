package bitxhub

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/codeskyblue/go-sh"
	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/internal/download"
	"github.com/meshplus/goduck/internal/types"
)

func DownloadBitxhubBinary(repoPath string, version string) error {
	path := fmt.Sprintf("bitxhub_%s_%s", runtime.GOOS, version)
	root := filepath.Join(repoPath, "bin", path)
	if !fileutil.Exist(root) {
		err := os.MkdirAll(root, 0755)
		if err != nil {
			return err
		}
	}

	if runtime.GOOS == "linux" {
		if !fileutil.Exist(filepath.Join(root, "bitxhub")) {
			url := fmt.Sprintf(types.BitxhubUrlLinux, version, version)
			err := download.Download(root, url)
			if err != nil {
				return err
			}
		}
	}
	if runtime.GOOS == "darwin" {
		if !fileutil.Exist(filepath.Join(root, "bitxhub")) {
			url := fmt.Sprintf(types.BitxhubUrlMacOS, version, version)
			err := download.Download(root, url)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func ExtractBitxhubBinary(repoPath string, version string) error {
	path := filepath.Join(repoPath, "bin", fmt.Sprintf("bitxhub_%s_%s", runtime.GOOS, version))
	var file string
	if runtime.GOOS == "linux" {
		file = fmt.Sprintf(types.BitxhubTarNameLinux, version)
	} else if runtime.GOOS == "darwin" {
		file = fmt.Sprintf(types.BitxhubTarNameMacOS, version)
	}

	if !fileutil.Exist(filepath.Join(path, "bitxhub")) {
		err := sh.Command("/bin/bash", "-c", fmt.Sprintf("cd %s && tar xzf %s", path, file)).Run()
		if err != nil {
			return err
		}
	}

	return nil
}

func DownloadBitxhubConfig(repoPath string, version string) error {
	if !fileutil.Exist(repoPath) {
		err := os.MkdirAll(repoPath, 0755)
		if err != nil {
			return err
		}
	}

	if !fileutil.Exist(filepath.Join(repoPath, types.BitxhubConfigName)) {
		url := fmt.Sprintf(types.BitxhubConfigUrl, version)
		err := download.Download(repoPath, url)
		if err != nil {
			return err
		}
	}

	return nil
}
