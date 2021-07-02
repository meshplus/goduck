package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

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
			Name:  "register",
			Usage: "Register pier to BitXHub",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "chain",
					Usage:    "specify appchain type, ethereum(default) or fabric",
					Required: false,
					Value:    types.ChainTypeEther,
				},
				&cli.StringFlag{
					Name:     "cryptoPath",
					Usage:    "path of crypto-config, only useful for fabric chain, e.g $HOME/crypto-config",
					Required: false,
				},
				&cli.StringFlag{
					Name:     "pierType",
					Usage:    "specify pier up type, docker(default) or binary",
					Required: false,
					Value:    types.TypeDocker,
				},
				&cli.StringFlag{
					Aliases: []string{"version", "v"},
					Value:   "v1.0.0-rc1",
					Usage:   "pier version",
				},
				&cli.StringFlag{
					Name:  "tls",
					Value: "false",
					Usage: "whether to enable TLS, true or false, only useful for v1.4.0+",
				},
				&cli.StringFlag{
					Name:  "httpPort",
					Value: "44544",
					Usage: "peer's http port, only useful for v1.4.0+",
				},
				&cli.StringFlag{
					Name:  "pprofPort",
					Value: "44550",
					Usage: "pier pprof port, only useful for binary",
				},
				&cli.StringFlag{
					Name:  "apiPort",
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
					Usage: "IP of appchain which pier connects to (default: \"127.0.0.1\")",
				},
				&cli.StringFlag{
					Name:  "appchainAddr",
					Usage: "address of appchain which pier connects to, e.g. ethereum \"ws://127.0.0.1:8546\", fabric \"127.0.0.1:7053\" ",
				},
				&cli.StringFlag{
					Name:  "appchainPorts",
					Usage: "ports of appchain which pier connects to. e.g. ethereum \"8546\", fabric \"7050, 7051, 7053, 8051, 8053, 9051, 9053, 10051, 10053\"",
				},
				&cli.StringFlag{
					Name:  "contractAddr",
					Value: "0xD3880ea40670eD51C3e3C0ea089fDbDc9e3FBBb4",
					Usage: "address of the contract on the appChain. Only works on Ethereum",
				},
				&cli.StringFlag{
					Name:  "pierRepo",
					Usage: "the startup path of the pier (default:$repo/pier/.pier_$chainType)",
				},
			},
			Action: pierRegister,
		},
		{
			Name:  "start",
			Usage: "Start pier with its appchain up",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:     "chain",
					Usage:    "specify appchain type, ethereum(default) or fabric",
					Required: false,
					Value:    types.ChainTypeEther,
				},
				&cli.StringFlag{
					Name:     "cryptoPath",
					Usage:    "path of crypto-config, only useful for fabric chain, e.g $HOME/crypto-config",
					Required: false,
				},
				&cli.StringFlag{
					Name:     "pierType",
					Usage:    "specify pier up type, docker(default) or binary",
					Required: false,
					Value:    types.TypeDocker,
				},
				&cli.StringFlag{
					Aliases: []string{"version", "v"},
					Value:   "v1.0.0-rc1",
					Usage:   "pier version",
				},
				&cli.StringFlag{
					Name:  "tls",
					Value: "false",
					Usage: "whether to enable TLS, true or false, only useful for v1.4.0+",
				},
				&cli.StringFlag{
					Name:  "httpPort",
					Value: "44544",
					Usage: "peer's http port, only useful for v1.4.0+",
				},
				&cli.StringFlag{
					Name:  "pprofPort",
					Value: "44550",
					Usage: "pier pprof port, only useful for binary",
				},
				&cli.StringFlag{
					Name:  "apiPort",
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
					Usage: "IP of appchain which pier connects to (default: \"127.0.0.1\")",
				},
				&cli.StringFlag{
					Name:  "appchainAddr",
					Usage: "address of appchain which pier connects to, e.g. ethereum \"ws://127.0.0.1:8546\", fabric \"127.0.0.1:7053\" ",
				},
				&cli.StringFlag{
					Name:  "appchainPorts",
					Usage: "ports of appchain which pier connects to. e.g. ethereum \"8546\", fabric \"7050, 7051, 7053, 8051, 8053, 9051, 9053, 10051, 10053\"",
				},
				&cli.StringFlag{
					Name:  "contractAddr",
					Value: "0xD3880ea40670eD51C3e3C0ea089fDbDc9e3FBBb4",
					Usage: "address of the contract on the appChain. Only works on Ethereum",
				},
				&cli.StringFlag{
					Name:  "pierRepo",
					Usage: "the startup path of the pier (default:$repo/pier/.pier_$chainType)",
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
					Value:    types.ChainTypeEther,
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
					Value:    types.ChainTypeEther,
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
					Usage: "configuration mode, one of direct, relay or union",
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
					Name:  "appchainType",
					Value: "ethereum",
					Usage: "appchain type, one of ethereum or fabric",
				},
				&cli.StringFlag{
					Name:  "appchainIP",
					Usage: "IP of appchain which pier connects to (default: \"127.0.0.1\")",
				},
				&cli.StringFlag{
					Name:  "appchainAddr",
					Usage: "address of appchain which pier connects to, e.g. ethereum \"ws://127.0.0.1:8546\", fabric \"127.0.0.1:7053\"",
				},
				&cli.StringFlag{
					Name:  "appchainPorts",
					Usage: "ports of appchain which pier connects to. e.g. ethereum \"8546\", fabric \"7050, 7051, 7053, 8051, 8053, 9051, 9053, 10051, 10053\"(The first one is port of orderer. The remaining, in turn, are the first node's urlSubstitutionExp port and eventUrlSubstitutionExp port, and the second node's urlSubstitutionExp port and eventUrlSubstitutionExp port...)",
				},
				&cli.StringFlag{
					Name:  "contractAddr",
					Value: "0xD3880ea40670eD51C3e3C0ea089fDbDc9e3FBBb4",
					Usage: "address of the contract on the appChain. Only works on Ethereum",
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
					Name:  "httpPort",
					Value: "44544",
					Usage: "peer's http port, only useful for v1.4.0+",
				},
				&cli.StringFlag{
					Name:  "pprofPort",
					Value: "44550",
					Usage: "peer's pprof port",
				},
				&cli.StringFlag{
					Name:  "apiPort",
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
					Aliases: []string{"version", "v"},
					Value:   "v1.0.0-rc1",
					Usage:   "pier version",
				},
			},
			Action: generatePierConfig,
		},
	},
}

func pierRegister(ctx *cli.Context) error {
	chainType := ctx.String("chain")
	cryptoPath := ctx.String("cryptoPath")
	pierUpType := ctx.String("pierType")
	version := ctx.String("version")
	tls := ctx.String("tls")
	http := ctx.String("httpPort")
	pport := ctx.String("pprofPort")
	aport := ctx.String("apiPort")
	overwrite := ctx.String("overwrite")
	appchainIP := ctx.String("appchainIP")
	appchainAddr := ctx.String("appchainAddr")
	appchainPortsTmp := ctx.String("appchainPorts")
	appchainPorts := strings.Replace(appchainPortsTmp, " ", "", -1)
	appchainContractAddr := ctx.String("contractAddr")
	pierRepo := ctx.String("pierRepo")

	appPorts, appchainAddr, appchainIP, err := getAppchainParams(chainType, appchainIP, appchainPorts, appchainAddr, cryptoPath)
	if err != nil {
		return err
	}
	appchainPorts = strings.Join(appPorts, ",")

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

	if pierRepo == "" {
		pierRepo = filepath.Join(repoRoot, fmt.Sprintf("pier/.pier_%s", chainType))
	}

	if !fileutil.Exist(filepath.Join(repoRoot, fmt.Sprintf("bin/pier_%s_%s/pier", runtime.GOOS, version))) {
		if err := downloadPierBinary(repoRoot, version, runtime.GOOS); err != nil {
			return fmt.Errorf("download pier binary error:%w", err)
		}
	}
	if pierUpType == types.TypeDocker && !fileutil.Exist(filepath.Join(repoRoot, fmt.Sprintf("bin/pier_linux_%s/pier", version))) {
		if err := downloadPierBinary(repoRoot, version, "linux"); err != nil {
			return fmt.Errorf("download pier binary error:%w", err)
		}
	}

	return pier.RegisterPier(repoRoot, chainType, cryptoPath, pierUpType, version, tls, http, pport, aport, overwrite, appchainIP, appchainAddr, appchainPorts, appchainContractAddr, pierRepo)
}

func pierStart(ctx *cli.Context) error {
	chainType := ctx.String("chain")
	cryptoPath := ctx.String("cryptoPath")
	pierUpType := ctx.String("pierType")
	version := ctx.String("version")
	tls := ctx.String("tls")
	http := ctx.String("httpPort")
	pport := ctx.String("pprofPort")
	aport := ctx.String("apiPort")
	overwrite := ctx.String("overwrite")
	appchainIP := ctx.String("appchainIP")
	appchainAddr := ctx.String("appchainAddr")
	appchainPorts := strings.Replace(ctx.String("appchainPorts"), " ", "", -1)
	appchainContractAddr := ctx.String("contractAddr")
	pierRepo := ctx.String("pierRepo")

	appPorts, appchainAddr, appchainIP, err := getAppchainParams(chainType, appchainIP, appchainPorts, appchainAddr, cryptoPath)
	if err != nil {
		return err
	}
	appchainPorts = strings.Join(appPorts, ",")

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

	if pierRepo == "" {
		pierRepo = filepath.Join(repoRoot, fmt.Sprintf("pier/.pier_%s", chainType))
	}

	if pierUpType == types.TypeBinary && !fileutil.Exist(pierRepo) {
		return fmt.Errorf("the pier startup path(%s) does not have a startup binary", pierRepo)
	}

	return pier.StartPier(repoRoot, chainType, cryptoPath, pierUpType, version, tls, http, pport, aport, overwrite, appchainIP, appchainAddr, appchainPorts, appchainContractAddr, pierRepo)
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

func downloadPierBinary(repoPath string, version string, system string) error {
	path := fmt.Sprintf("pier_%s_%s", system, version)
	root := filepath.Join(repoPath, "bin", path)
	if !fileutil.Exist(root) {
		err := os.MkdirAll(root, 0755)
		if err != nil {
			return err
		}
	}

	if system == "linux" {
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
				err := download.Download(root, fmt.Sprintf(types.LinuxWasmLibUrl, version))
				if err != nil {
					return err
				}
			}
		}
		if !fileutil.Exist(filepath.Join(root, types.FabricClient)) && !fileutil.Exist(filepath.Join(root, types.FabricClientSo)) {
			url := fmt.Sprintf(types.PierFabricClientUrlLinux, version, version)
			err := download.Download(root, url)
			if err != nil {
				return err
			}

			err = sh.Command("/bin/bash", "-c", fmt.Sprintf("cd %s && mv fabric-client-%s-Darwin %s && chmod +x %s", root, version, types.FabricClient, types.FabricClient)).Run()
			if err != nil {
				return fmt.Errorf("rename fabric client error: %s", err)
			}
		}

		if !fileutil.Exist(filepath.Join(root, types.EthClient)) && !fileutil.Exist(filepath.Join(root, types.EthClientSo)) {
			url := fmt.Sprintf(types.PierEthereumClientUrlLinux, version, version)
			err := download.Download(root, url)
			if err != nil {
				return err
			}

			err = sh.Command("/bin/bash", "-c", fmt.Sprintf("cd %s && mv eth-client-%s-Darwin %s && chmod +x %s", root, version, types.EthClient, types.EthClient)).Run()
			if err != nil {
				return fmt.Errorf("rename eth client error: %s", err)
			}
		}
	}
	if system == "darwin" {
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

		if !fileutil.Exist(filepath.Join(root, types.FabricClient)) && !fileutil.Exist(filepath.Join(root, types.FabricClientSo)) {
			url := fmt.Sprintf(types.PierFabricClientUrlMacOS, version, version)
			err := download.Download(root, url)
			if err != nil {
				return err
			}

			err = sh.Command("/bin/bash", "-c", fmt.Sprintf("cd %s && mv fabric-client-%s-Darwin %s && chmod +x %s", root, version, types.FabricClient, types.FabricClient)).Run()
			if err != nil {
				return fmt.Errorf("rename fabric client error: %s", err)
			}
		}

		if !fileutil.Exist(filepath.Join(root, types.EthClient)) && !fileutil.Exist(filepath.Join(root, types.EthClientSo)) {
			url := fmt.Sprintf(types.PierEthereumClientUrlMacOS, version, version)
			err := download.Download(root, url)
			if err != nil {
				return err
			}

			err = sh.Command("/bin/bash", "-c", fmt.Sprintf("cd %s && mv eth-client-%s-Darwin %s && chmod +x %s", root, version, types.EthClient, types.EthClient)).Run()
			if err != nil {
				return fmt.Errorf("rename eth client error: %s", err)
			}
		}

		if !fileutil.Exist(filepath.Join(root, "libwasmer.dylib")) {
			err := download.Download(root, fmt.Sprintf(types.MacOSWasmLibUrl, version))
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func generatePierConfig(ctx *cli.Context) error {
	repoRoot, err := repo.PathRoot()
	if err != nil {
		return err
	}

	mode := ctx.String("mode")
	startType := ctx.String("type")
	bitxhub := ctx.String("bitxhub")
	validators := ctx.StringSlice("validators")
	port := ctx.String("port")
	peers := ctx.StringSlice("peers")
	connectors := ctx.StringSlice("connectors")
	providers := ctx.String("providers")
	appchainType := ctx.String("appchainType")
	appchainIP := ctx.String("appchainIP")
	target := ctx.String("target")
	tls := ctx.String("tls")
	httpPort := ctx.String("httpPort")
	pprofPort := ctx.String("pprofPort")
	apiPort := ctx.String("apiPort")
	cryptoPath := ctx.String("cryptoPath")
	version := ctx.String("version")
	appchainAddr := ctx.String("appchainAddr")
	appchainPortsTmp := ctx.String("appchainPorts")
	appchainPorts := strings.Replace(appchainPortsTmp, " ", "", -1)
	appchainContractAddr := ctx.String("contractAddr")

	appPorts, appchainAddr, appchainIP, err := getAppchainParams(appchainType, appchainIP, appchainPorts, appchainAddr, cryptoPath)
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

	if !AdjustVersion(version, release.Bitxhub) {
		return fmt.Errorf("unsupport pier verison")
	}

	// generate key.json need pier binary file
	pierP := fmt.Sprintf("bin/pier_%s_%s/pier", runtime.GOOS, version)
	pierPath := filepath.Join(repoRoot, pierP)
	if !fileutil.Exist(pierPath) {
		if err := downloadPierBinary(repoRoot, version, runtime.GOOS); err != nil {
			return fmt.Errorf("download pier binary error:%w", err)
		}
	}
	if startType == types.TypeDocker && !fileutil.Exist(filepath.Join(repoRoot, fmt.Sprintf("bin/pier_linux_%s/pier", version))) {
		if err := downloadPierBinary(repoRoot, version, "linux"); err != nil {
			return fmt.Errorf("download pier binary error:%w", err)
		}
	}

	return InitPierConfig(mode, startType, bitxhub, validators, port, peers, connectors, providers, appchainType, appchainIP, appchainAddr, appPorts, appchainContractAddr, target, tls, httpPort, pprofPort, apiPort, version, pierPath, cryptoPath)
}

func getAppchainParams(chainType, appchainIP, appchainPorts, appchainAddr, cryptoPath string) ([]string, string, string, error) {
	var appPorts []string
	switch chainType {
	case types.ChainTypeFabric:
		if appchainPorts == "" {
			appPorts = append(appPorts, "7050", "7051", "7053", "8051", "8053", "9051", "9053", "10051", "10053")
		} else {
			appPorts = strings.Split(appchainPorts, ",")
			if len(appPorts) != 9 {
				return nil, "", "", fmt.Errorf("The specified number of application chain ports is incorrect. Fabric needs to specify 9 ports.")
			}
			if err := checkPorts(appPorts); err != nil {
				return nil, "", "", fmt.Errorf("The port cannot be repeated: %w", err)
			}
		}

		if appchainAddr == "" {
			if appchainIP == "" {
				appchainIP = "127.0.0.1"
			}
			appchainAddr = fmt.Sprintf("%s:%s", appchainIP, appPorts[2])
		} else {
			if appchainPorts != "" {
				if !strings.Contains(appchainAddr, appPorts[2]) && !strings.Contains(appchainAddr, appPorts[4]) && !strings.Contains(appchainAddr, appPorts[6]) && !strings.Contains(appchainAddr, appPorts[8]) {
					return nil, "", "", fmt.Errorf("AppchainAddr and appchainPorts are inconsistent. Please check the input parameters.\n 1. The port in appchainAddr should be the eventUrlSubstitutionExp port of a fabric node; \n 2. The order in which ports are specified is：the first one is port of orderer, the remaining, in turn, are the first node's urlSubstitutionExp port and eventUrlSubstitutionExp port, and the second node's urlSubstitutionExp port and eventUrlSubstitutionExp port...")
				}
			} else {
				return nil, "", "", fmt.Errorf("Please specify other ports for the Fabric chain.")
			}

			if appchainIP != "" {
				if !strings.Contains(appchainAddr, appchainIP) {
					return nil, "", "", fmt.Errorf("AppchainAddr and appchainIP are inconsistent. Please check the input parameters.")
				}
			} else {
				appchainIP = strings.Split(appchainAddr, ":")[0]
			}
		}

		if cryptoPath == "" {
			return nil, "", "", fmt.Errorf("Start fabric pier need crypto-config path.")
		}
	case types.ChainTypeEther:
		if appchainAddr == "" {
			if appchainIP == "" {
				appchainIP = "127.0.0.1"
			}

			if appchainPorts == "" {
				appPorts = append(appPorts, "8546")
			} else {
				appPorts = strings.Split(appchainPorts, ",")
				if len(appPorts) != 1 {
					return nil, "", "", fmt.Errorf("The specified number of application chain ports is incorrect. Ethereum needs to specify 1 port.")
				}
			}

			appchainAddr = fmt.Sprintf("ws://%s:%s", appchainIP, appPorts[0])
		} else {
			if appchainPorts != "" {
				if appchainPorts != "0000" {
					appPorts = strings.Split(appchainPorts, ",")
					if len(appPorts) != 1 {
						return nil, "", "", fmt.Errorf("The specified number of application chain ports is incorrect. Ethereum needs to specify 1 port.")
					}
					if !strings.Contains(appchainAddr, appPorts[0]) {
						return nil, "", "", fmt.Errorf("AppchainAddr(%s) and appchainPorts(%s) are inconsistent. Please check the input parameters.", appchainAddr, appchainPorts)
					}
				} else {
					appPorts = append(appPorts, "0000")
				}
			} else {
				appPorts = append(appPorts, "0000")
			}
			if appchainIP != "" {
				if appchainIP != "0.0.0.0" {
					if !strings.Contains(appchainAddr, appchainIP) {
						return nil, "", "", fmt.Errorf("AppchainAddr and appchainIP are inconsistent. Please check the input parameters.")
					}
				}
			} else {
				// In the case of Ethereum, if ADDR is given, then the IP parameter will be invalid and will just be assigned a default value
				appchainIP = "0.0.0.0"
			}
		}

	default:
		return nil, "", "", fmt.Errorf("unsupported appchain type")
	}

	return appPorts, appchainAddr, appchainIP, nil
}

func checkPorts(ports []string) error {
	portM := make(map[string]int, 0)
	for i, p := range ports {
		_, ok := portM[p]
		if ok {
			return fmt.Errorf("%s", p)
		}
		portM[p] = i
	}
	return nil
}
