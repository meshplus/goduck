package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/codeskyblue/go-sh"
	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/cmd/goduck/pier"
	"github.com/meshplus/goduck/internal/download"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/urfave/cli/v2"
)

var pierCMD = &cli.Command{
	Name:  "pier",
	Usage: "Operation about pier",
	Subcommands: []*cli.Command{
		{
			Name:  "start",
			Usage: "Start pier with its appchain up",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "chain",
					Usage:    "specify appchain type, ethereum(default) or fabric",
					Required: false,
					Value:    types.Ethereum,
				},
				&cli.StringFlag{
					Name:     "cryptoPath",
					Usage:    "path of crypto-config, only useful for fabric chain, e.g $HOME/crypto-config",
					Required: false,
				},
				&cli.StringFlag{
					Name:     "pier-type",
					Usage:    "specify pier up type, docker(default) or binary",
					Required: false,
					Value:    types.TypeDocker,
				},
				&cli.StringFlag{
					Name:  "version,v",
					Value: "v1.5.0",
					Usage: "pier version",
				},
				&cli.StringFlag{
					Name:  "tls",
					Value: "false",
					Usage: "whether to enable TLS, true or false, only useful for v1.4.0+",
				},
				&cli.StringFlag{
					Name:  "http-port",
					Value: "44544",
					Usage: "peer's http port, only useful for v1.4.0+",
				},
				&cli.StringFlag{
					Name:  "pprof-port",
					Value: "44550",
					Usage: "pier pprof port, only useful for binary",
				},
				&cli.StringFlag{
					Name:  "api-port",
					Value: "8080",
					Usage: "pier api port, only useful for binary",
				},
				&cli.StringFlag{
					Name:  "overwrite",
					Value: "true",
					Usage: "whether to overwrite the configuration if the pier configuration file exists locally",
				},
				&cli.StringFlag{
					Name:  "appchainIP",
					Value: "127.0.0.1",
					Usage: "the application chain IP that pier connects to",
				},
			},
			Action: pierStart,
		},
		{
			Name:  "stop",
			Usage: "Stop pier with its appchain down",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "chain",
					Usage:    "specify appchain type, ethereum(default) or fabric",
					Required: false,
					Value:    types.Ethereum,
				},
			},
			Action: pierStop,
		},
		{
			Name:  "clean",
			Usage: "Clean pier with its appchain",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "chain",
					Usage:    "specify appchain type, ethereum(default) or fabric",
					Required: false,
					Value:    types.Ethereum,
				},
			},
			Action: pierClean,
		},
		{
			Name:  "config",
			Usage: "Generate configuration for Pier",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "mode",
					Value: types.PierModeDirect,
					Usage: "configuration mode, one of direct or relay",
				},
				&cli.StringFlag{
					Name:  "type",
					Value: types.TypeBinary,
					Usage: "configuration type, one of binary or docker",
				},
				&cli.StringFlag{
					Name:  "bitxhub",
					Usage: "BitXHub's address, only useful in relay mode",
				},
				&cli.StringSliceFlag{
					Name:  "validators",
					Usage: "BitXHub's validators, only useful in relay mode, e.g. --validators \"0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013\" --validators \"0x79a1215469FaB6f9c63c1816b45183AD3624bE34\" --validators \"0x97c8B516D19edBf575D72a172Af7F418BE498C37\" --validators \"0x97c8B516D19edBf575D72a172Af7F418BE498C37\"",
				},
				&cli.StringFlag{
					Name:  "port",
					Value: "5001",
					Usage: "pier's port, only useful in direct mode",
				},
				&cli.StringSliceFlag{
					Name:  "peers",
					Usage: "peers' address, only useful in direct mode, e.g. --peers \"/ip4/127.0.0.1/tcp/4001/p2p/Qma1oh5JtrV24gfP9bFrVv4miGKz7AABpfJhZ4F2Z5ngmL\"",
				},
				&cli.StringSliceFlag{
					Name:  "connectors",
					Usage: "address of peers which need to connect, only useful in union mode for v1.4.0+, e.g. --connectors \"/ip4/127.0.0.1/tcp/4001/p2p/Qma1oh5JtrV24gfP9bFrVv4miGKz7AABpfJhZ4F2Z5ngmL\" --connectors \"/ip4/127.0.0.1/tcp/4002/p2p/Qma1oh5JtrV24gfP9bFrVv4miGKz7AABpfJhZ4F2Z5abcD\"",
				},
				&cli.StringFlag{
					Name:  "providers",
					Value: "1",
					Usage: "the minimum number of cross-chain gateways that need to be found in a large-scale network, only useful in union mode for v1.4.0+",
				},
				&cli.StringFlag{
					Name:  "appchain-type",
					Value: "ethereum",
					Usage: "appchain type, one of ethereum or fabric",
				},
				&cli.StringFlag{
					Name:  "appchain-IP",
					Value: "127.0.0.1",
					Usage: "appchain IP address",
				},
				&cli.StringFlag{
					Name:  "target",
					Value: ".",
					Usage: "where to put the generated configuration files",
				},
				&cli.StringFlag{
					Name:  "tls",
					Value: "false",
					Usage: "whether to enable TLS, only useful for v1.4.0+",
				},
				&cli.StringFlag{
					Name:  "http-port",
					Value: "44544",
					Usage: "peer's http port, only useful for v1.4.0+",
				},
				&cli.StringFlag{
					Name:  "pprof-port",
					Value: "44550",
					Usage: "peer's pprof port",
				},
				&cli.StringFlag{
					Name:  "api-port",
					Value: "8080",
					Usage: "peer's api port",
				},
				&cli.StringFlag{
					Name:     "cryptoPath",
					Usage:    "path of crypto-config, only useful for fabric chain, e.g $HOME/crypto-config",
					Value:    "$HOME/.goduck/crypto-config",
					Required: false,
				},
				&cli.StringFlag{
					Name:  "version",
					Value: "v1.5.0",
					Usage: "pier version",
				},
			},
			Action: generatePierConfig,
		},
	},
}

func pierStart(ctx *cli.Context) error {
	chainType := ctx.String("chain")
	cryptoPath := ctx.String("cryptoPath")
	pierUpType := ctx.String("pier-type")
	version := ctx.String("version")
	tls := ctx.String("tls")
	http := ctx.String("http-port")
	pport := ctx.String("pprof-port")
	aport := ctx.String("api-port")
	overwrite := ctx.String("overwrite")
	appchainIP := ctx.String("appchainIP")

	if chainType == "fabric" && cryptoPath == "" {
		return fmt.Errorf("start fabric pier need crypto-config path")
	}

	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	data, err := ioutil.ReadFile(filepath.Join(repoRoot, "release.json"))
	if err != nil {
		return err
	}

	var release *Release
	if err := json.Unmarshal(data, &release); err != nil {
		return err
	}

	if !AdjustVersion(version, release.Pier) {
		return fmt.Errorf("unsupport pier verison")
	}

	if pierUpType == types.TypeBinary {
		if !fileutil.Exist(filepath.Join(repoRoot, fmt.Sprintf("bin/pier_%s_%s/pier", runtime.GOOS, version))) {
			if err := downloadPierBinary(repoRoot, version); err != nil {
				return fmt.Errorf("download pier binary error:%w", err)
			}
		}
	}

	return pier.StartPier(repoRoot, chainType, cryptoPath, pierUpType, version, tls, http, pport, aport, overwrite, appchainIP)
}

func pierStop(ctx *cli.Context) error {
	chainType := ctx.String("chain")

	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	return pier.StopPier(repoRoot, chainType)
}

func pierClean(ctx *cli.Context) error {
	chainType := ctx.String("chain")

	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	return pier.CleanPier(repoRoot, chainType)
}

func downloadPierBinary(repoPath string, version string) error {
	path := fmt.Sprintf("pier_%s_%s", runtime.GOOS, version)
	root := filepath.Join(repoPath, "bin", path)
	if !fileutil.Exist(root) {
		err := os.MkdirAll(root, 0755)
		if err != nil {
			return err
		}
	}

	if runtime.GOOS == "linux" {
		if !fileutil.Exist(filepath.Join(root, "pier")) {
			url := fmt.Sprintf(types.PierUrlLinux, version, version)
			err := download.Download(root, url)
			if err != nil {
				return err
			}

			err = sh.Command("/bin/bash", "-c", fmt.Sprintf("cd %s && tar xf pier_linux-amd64_%s.tar.gz -C %s --strip-components 1 && export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:%s", root, version, root, root)).Run()
			if err != nil {
				return fmt.Errorf("extract pier binary: %s", err)
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
			url := fmt.Sprintf(types.PierUrlMacOS, version, version)
			err := download.Download(root, url)
			if err != nil {
				return err
			}

			err = sh.Command("/bin/bash", "-c", fmt.Sprintf("cd %s && tar xf pier_darwin_x86_64_%s.tar.gz -C %s --strip-components 1 && install_name_tool -change @rpath/libwasmer.dylib %s/libwasmer.dylib %s/pier", root, version, root, root, root)).Run()
			if err != nil {
				return fmt.Errorf("extract pier binary: %s", err)
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
