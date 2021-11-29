package pier

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codeskyblue/go-sh"
	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/internal/download"
	"github.com/meshplus/goduck/internal/types"
)

func DownloadPierBinary(repoPath string, version string, system string) error {
	path := fmt.Sprintf("pier_%s_%s", system, version)
	root := filepath.Join(repoPath, "bin", path)
	if !fileutil.Exist(root) {
		err := os.MkdirAll(root, 0755)
		if err != nil {
			return err
		}
	}

	if !fileutil.Exist(filepath.Join(root, "pier")) {
		if system == "linux" {
			url := fmt.Sprintf(types.PierUrlLinux, version, version)
			err := download.Download(root, url)
			if err != nil {
				return err
			}

			err = sh.Command("/bin/bash", "-c", fmt.Sprintf("cd %s && tar xf pier_linux-amd64_%s.tar.gz  && export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:%s", root, version, root)).Run()
			if err != nil {
				return fmt.Errorf("extract pier binary: %s", err)
			}
		}
		if system == "darwin" {
			url := fmt.Sprintf(types.PierUrlMacOS, version, version)
			err := download.Download(root, url)
			if err != nil {
				return err
			}

			err = sh.Command("/bin/bash", "-c", fmt.Sprintf("cd %s && tar xf pier_darwin_x86_64_%s.tar.gz  && install_name_tool -change @rpath/libwasmer.dylib %s/libwasmer.dylib %s/pier", root, version, root, root)).Run()
			if err != nil {
				return fmt.Errorf("extract pier binary: %s", err)
			}
		}
	}

	return nil
}

func DownloadPierPlugin(repoPath string, chain string, version string, system string) error {
	path := fmt.Sprintf("pier_%s_%s", system, version)
	root := filepath.Join(repoPath, "bin", path)
	pluginName := fmt.Sprintf(types.PierPlugin, chain)
	if !fileutil.Exist(root) {
		err := os.MkdirAll(root, 0755)
		if err != nil {
			return err
		}
	}
	if !fileutil.Exist(filepath.Join(root, pluginName)) {
		if system == "linux" {
			switch chain {
			case types.ChainTypeFabric:
				url := fmt.Sprintf(types.PierFabricClientUrlLinux, version, version)
				err := download.Download(root, url)
				if err != nil {
					return err
				}

				err = sh.Command("/bin/bash", "-c", fmt.Sprintf("cd %s && mv fabric-client-%s-Linux %s && chmod +x %s", root, version, pluginName, pluginName)).Run()
				if err != nil {
					return fmt.Errorf("rename fabric client error: %s", err)
				}
			case types.ChainTypeEther:
				url := fmt.Sprintf(types.PierEthereumClientUrlLinux, version, version)
				err := download.Download(root, url)
				if err != nil {
					return err
				}

				err = sh.Command("/bin/bash", "-c", fmt.Sprintf("cd %s && mv eth-client-%s-Linux %s && chmod +x %s", root, version, pluginName, pluginName)).Run()
				if err != nil {
					return fmt.Errorf("rename eth client error: %s", err)
				}

			}

		}
		if system == "darwin" {
			switch chain {
			case types.ChainTypeFabric:
				url := fmt.Sprintf(types.PierFabricClientUrlMacOS, version, version)
				err := download.Download(root, url)
				if err != nil {
					return err
				}

				err = sh.Command("/bin/bash", "-c", fmt.Sprintf("cd %s && mv fabric-client-%s-Darwin %s && chmod +x %s", root, version, pluginName, pluginName)).Run()
				if err != nil {
					return fmt.Errorf("rename fabric client error: %s", err)
				}

			case types.ChainTypeEther:
				url := fmt.Sprintf(types.PierEthereumClientUrlMacOS, version, version)
				err := download.Download(root, url)
				if err != nil {
					return err
				}

				err = sh.Command("/bin/bash", "-c", fmt.Sprintf("cd %s && mv eth-client-%s-Darwin %s && chmod +x %s", root, version, pluginName, pluginName)).Run()
				if err != nil {
					return fmt.Errorf("rename eth client error: %s", err)
				}
			}
		}
	}

	return nil
}
