package pier

import (
	"os"
	"path/filepath"
	"runtime"

	"github.com/meshplus/bitxhub-kit/fileutil"

	"github.com/meshplus/goduck/internal/download"
	"github.com/meshplus/goduck/internal/types"
)

func DownloadPierBinary(repoPath string) error {
	root := filepath.Join(repoPath, "bin")
	if !fileutil.Exist(root) {
		err := os.Mkdir(root, 0755)
		if err != nil {
			return err
		}
	}

	if runtime.GOOS == "linux" {
		if !fileutil.Exist(filepath.Join(root, "pier")) {
			err := download.Download(root, types.PierUrlLinux)
			if err != nil {
				return err
			}

			if !fileutil.Exist(filepath.Join(root, "libwasmer.so")) {
				err := download.Download(root, types.LinuxWasmLibUrl)
				if err != nil {
					return err
				}
			}
		}
	}
	if runtime.GOOS == "darwin" {
		if !fileutil.Exist(filepath.Join(root, "pier")) {
			err := download.Download(root, types.PierUrlMacOS)
			if err != nil {
				return err
			}
		}

		if !fileutil.Exist(filepath.Join(root, "libwasmer.dylib")) {
			err := download.Download(root, types.MacOSWasmLibUrl)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
