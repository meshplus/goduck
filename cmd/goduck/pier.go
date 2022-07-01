package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fatih/color"
	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/cmd/goduck/pier"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/urfave/cli/v2"
)

var pierConfigMap = map[string]string{
	"v1.6.1":  "v1.6.1",
	"v1.6.5":  "v1.6.5", // v1.6.1
	"v1.7.0":  "v1.7.0", // v1.6.1
	"v1.8.0":  "v1.8.0",
	"v1.9.0":  "v1.9.0", // v1.8.0
	"v1.11.0": "v1.11.0",
	"v1.11.1": "v1.11.0",
	"v1.11.3": "v1.11.3",
	"v1.23.0": "v1.23.0",
}

var pierCMD = &cli.Command{
	Name:  "pier",
	Usage: "Operation about pier",
	Subcommands: []*cli.Command{
		{
			Name:  "start",
			Usage: "Start pier with its appchain up",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "appchain",
					Usage: "Specify appchain type, one of ethereum or fabric",
					Value: types.ChainTypeEther,
				},
				&cli.StringFlag{
					Name:  "target",
					Usage: "Specify the startup path of the pier (default:$repo/pier/.pier_$chainType)",
				},
				&cli.StringFlag{
					Name:  "upType",
					Usage: "Specify the startup type, one of binary or docker",
					Value: types.TypeBinary,
				},
				&cli.StringFlag{
					Name:  "configPath",
					Usage: "Specify the configuration file path for the configuration to be modified, default: $repo/pier_config/$VERSION/pier_modify_config.toml",
				},
				&cli.StringFlag{
					Aliases: []string{"version", "v"},
					Value:   "v1.23.0",
					Usage:   "Pier version",
				},
			},
			Action: pierStart,
		},
		{
			Name:  "register",
			Usage: "Register pier to BitXHub",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "appchain",
					Usage: "Specify appchain type, one of ethereum or fabric",
					Value: types.ChainTypeEther,
				},
				&cli.StringFlag{
					Name:  "target",
					Usage: "Specify the startup path of the pier (default:$repo/pier/.pier_$chainType)",
				},
				&cli.StringFlag{
					Name:  "upType",
					Usage: "Specify the startup type, one of binary or docker",
					Value: types.TypeBinary,
				},
				&cli.StringFlag{
					Name:  "method",
					Usage: "Specify appchain method, only useful for v1.8.0+",
					Value: "appchain",
				},
				&cli.StringFlag{
					Name:  "cid",
					Usage: "Specify the contanierID of the pier, only useful for docker",
				},
				&cli.StringFlag{
					Aliases: []string{"version", "v"},
					Value:   "v1.23.0",
					Usage:   "Pier version",
				},
			},
			Action: pierRegister,
		},
		{
			Name:  "rule",
			Usage: "deploy rule to BitXHub",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "appchain",
					Usage: "Specify appchain type, one of ethereum or fabric",
					Value: types.ChainTypeEther,
				},
				&cli.StringFlag{
					Name:  "target",
					Usage: "Specify the startup path of the pier, only useful for binary (default:$repo/pier/.pier_$chainType)",
				},
				&cli.StringFlag{
					Name:  "cid",
					Usage: "Specify the contanierID of the pier, only useful for docker",
				},
				&cli.StringFlag{
					Name:  "ruleRepo",
					Usage: "Specify the path of the rule (default:$repo/pier/.pier_$chainType/$chainType/validating.wasm)",
				},
				&cli.StringFlag{
					Name:  "upType",
					Usage: "Specify the startup type, one of binary or docker",
					Value: types.TypeBinary,
				},
				&cli.StringFlag{
					Name:  "method",
					Usage: "Specify appchain method, only useful for v1.8.0+",
					Value: "appchain",
				},
				&cli.StringFlag{
					Aliases: []string{"version", "v"},
					Value:   "v1.23.0",
					Usage:   "Pier version",
				},
			},
			Action: pierRuleDeploy,
		},
		{
			Name:  "stop",
			Usage: "Stop pier with its appchain down",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "appchain",
					Usage: "Specify appchain type, one of ethereum or fabric",
					Value: types.ChainTypeEther,
				},
			},
			Action: pierStop,
		},
		{
			Name:  "clean",
			Usage: "Clean pier with its appchain",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "appchain",
					Usage: "Specify appchain type, one of ethereum or fabric",
					Value: types.ChainTypeEther,
				},
			},
			Action: pierClean,
		},
		{
			Name:  "config",
			Usage: "Generate configuration for Pier",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name:  "appchain",
					Usage: "Specify appchain type, one of ethereum or fabric",
					Value: types.ChainTypeEther,
				},
				&cli.StringFlag{
					Name:  "target",
					Usage: "Specify the directory to where to put the generated configuration files, default: $repo/pier/.pier_$APPCHAINTYPE/",
				},
				&cli.StringFlag{
					Name:  "configPath",
					Usage: "Specify the configuration file path for the configuration to be modified, default: $repo/pier_config/$VERSION/pier_modify_config.toml",
				},
				&cli.StringFlag{
					Name:  "upType",
					Usage: "Specify the startup type, one of binary or docker",
					Value: types.TypeBinary,
				},
				&cli.StringFlag{
					Aliases: []string{"version", "v"},
					Value:   "v1.6.5",
					Usage:   "Pier version",
				},
			},
			Action: generatePierConfig,
		},
	},
}

func pierStart(ctx *cli.Context) error {
	chainType := ctx.String("appchain")
	target := ctx.String("target")
	upType := ctx.String("upType")
	configPath := ctx.String("configPath")
	version := ctx.String("version")

	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}
	if !fileutil.Exist(filepath.Join(repoRoot, types.ReleaseJson)) {
		return fmt.Errorf("please `goduck init` first")
	}

	data, err := ioutil.ReadFile(filepath.Join(repoRoot, types.ReleaseJson))
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

	if target == "" {
		target = filepath.Join(repoRoot, fmt.Sprintf("pier/.pier_%s", chainType))
	} else {
		target, err = filepath.Abs(target)
		if err != nil {
			return fmt.Errorf("get absolute pier repo path: %w", err)
		}
	}

	if upType == types.TypeBinary && !fileutil.Exist(target) {
		if err := os.MkdirAll(target, 0755); err != nil {
			return err
		}
	}

	if configPath == "" {
		configPath = filepath.Join(repoRoot, fmt.Sprintf("%s/%s/%s", types.PierConfigRepo, pierConfigMap[version], types.PierModifyConfig))
	} else {
		configPath, err = filepath.Abs(configPath)
		if err != nil {
			return fmt.Errorf("get absolute config path: %w", err)
		}
	}

	if err := pier.DownloadPierBinary(repoRoot, version, runtime.GOOS); err != nil {
		return fmt.Errorf("download pier binary error:%w", err)
	}
	pluginSys := runtime.GOOS
	if upType == types.TypeDocker {
		pluginSys = types.LinuxSystem
	}
	if err := pier.DownloadPierPlugin(repoRoot, chainType, version, pluginSys); err != nil {
		return fmt.Errorf("download pier binary error:%w", err)
	}

	return pier.StartPier(repoRoot, chainType, target, upType, configPath, version)
}

func pierRegister(ctx *cli.Context) error {
	chainType := ctx.String("appchain")
	target := ctx.String("target")
	upType := ctx.String("upType")
	method := ctx.String("method")
	version := ctx.String("version")
	cid := ctx.String("cid")

	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}
	if !fileutil.Exist(filepath.Join(repoRoot, types.ReleaseJson)) {
		return fmt.Errorf("please `goduck init` first")
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

	if target == "" {
		target = filepath.Join(repoRoot, fmt.Sprintf("pier/.pier_%s", chainType))
	} else {
		target, err = filepath.Abs(target)
		if err != nil {
			return fmt.Errorf("get absolute pier repo path: %w", err)
		}
	}

	if upType == types.TypeBinary && !fileutil.Exist(target) {
		return fmt.Errorf("the pier startup path(%s) does not have a startup binary", target)
	}

	if upType == types.TypeDocker && cid == "" {
		return fmt.Errorf("Docker mode needs to specify CID (you can find it by using the conmand `goduck status list`)")
	}

	if upType == types.TypeBinary {
		if err := pier.DownloadPierBinary(repoRoot, version, runtime.GOOS); err != nil {
			return fmt.Errorf("download pier binary error:%w", err)
		}
		binPath := filepath.Join(repoRoot, fmt.Sprintf("bin/%s", fmt.Sprintf("pier_%s_%s", runtime.GOOS, version)))
		color.Blue("pier binary path: %s", binPath)
	}

	return pier.RegisterPier(repoRoot, target, chainType, upType, method, version, cid)
	//return pier.RegisterPier(repoRoot, chainType, cryptoPath, pierUpType, version, tls, http, pport, aport, overwrite, appchainIP, appchainAddr, appchainPorts, appchainContractAddr, pierRepo, adminKey, method)
}

func pierRuleDeploy(ctx *cli.Context) error {
	chainType := ctx.String("appchain")
	target := ctx.String("target")
	ruleRepo := ctx.String("ruleRepo")
	upType := ctx.String("upType")
	method := ctx.String("method")
	version := ctx.String("version")
	cid := ctx.String("cid")

	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}
	if !fileutil.Exist(filepath.Join(repoRoot, types.ReleaseJson)) {
		return fmt.Errorf("please `goduck init` first")
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

	if target == "" {
		target = filepath.Join(repoRoot, fmt.Sprintf("pier/.pier_%s", chainType))
	} else {
		target, err = filepath.Abs(target)
		if err != nil {
			return fmt.Errorf("get absolute pier repo path: %w", err)
		}
	}

	if upType == types.TypeBinary && !fileutil.Exist(target) {
		return fmt.Errorf("the pier startup path(%s) does not have a startup binary", target)
	}

	if upType == types.TypeDocker && cid == "" {
		return fmt.Errorf("Docker mode needs to specify CID (you can find it by using the conmand `goduck status list`)")
	}

	if ruleRepo == "" {
		ruleRepo = filepath.Join(repoRoot, fmt.Sprintf("pier/.pier_%s/%s/validating.wasm", chainType, chainType))
	}

	return pier.DeployRule(repoRoot, chainType, target, ruleRepo, upType, method, version, cid)
}

func pierStop(ctx *cli.Context) error {
	chainType := ctx.String("appchain")

	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}
	if !fileutil.Exist(filepath.Join(repoRoot, types.ReleaseJson)) {
		return fmt.Errorf("please `goduck init` first")
	}

	return pier.StopPier(repoRoot, chainType)
}

func pierClean(ctx *cli.Context) error {
	chainType := ctx.String("appchain")

	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}
	if !fileutil.Exist(filepath.Join(repoRoot, types.ReleaseJson)) {
		return fmt.Errorf("please `goduck init` first")
	}

	return pier.CleanPier(repoRoot, chainType)
}

func generatePierConfig(ctx *cli.Context) error {
	chainType := ctx.String("appchain")
	target := ctx.String("target")
	configPath := ctx.String("configPath")
	version := ctx.String("version")
	upType := ctx.String("upType")

	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return fmt.Errorf("parse repo path error:%w", err)
	}
	if !fileutil.Exist(filepath.Join(repoRoot, types.ReleaseJson)) {
		return fmt.Errorf("please `goduck init` first")
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
		return fmt.Errorf("unsupport Pier verison")
	}

	if target == "" {
		target = filepath.Join(repoRoot, fmt.Sprintf("pier/.pier_%s", chainType))
	} else {
		target, err = filepath.Abs(target)
		if err != nil {
			return fmt.Errorf("get absolute target path: %w", err)
		}
	}

	if _, err := os.Stat(target); os.IsNotExist(err) {
		if err := os.MkdirAll(target, 0755); err != nil {
			return err
		}
	}

	if configPath == "" {
		configPath = filepath.Join(repoRoot, fmt.Sprintf("%s/%s/%s", types.PierConfigRepo, pierConfigMap[version], types.PierModifyConfig))
	}

	if err := pier.DownloadPierBinary(repoRoot, version, runtime.GOOS); err != nil {
		return fmt.Errorf("download pier binary error:%w", err)
	}
	pluginSys := runtime.GOOS
	if upType == types.TypeDocker {
		pluginSys = types.LinuxSystem
	}
	if err := pier.DownloadPierPlugin(repoRoot, chainType, version, pluginSys); err != nil {
		return fmt.Errorf("download pier binary error:%w", err)
	}
	binPath := filepath.Join(repoRoot, fmt.Sprintf("bin/%s", fmt.Sprintf("pier_%s_%s", runtime.GOOS, version)))
	pluginPath := filepath.Join(repoRoot, fmt.Sprintf("bin/%s", fmt.Sprintf("pier_%s_%s", pluginSys, version)))
	color.Blue("pier binary path: %s", binPath)
	appchainConfigPath := filepath.Join(repoRoot, fmt.Sprintf("pier/%s", chainType))
	if chainType == types.ChainTypeEther {
		appchainConfigPath = filepath.Join(repoRoot, fmt.Sprintf("pier/%s/%s", chainType, EthConfigMap[version]))
	}

	return pier.GeneratePier(filepath.Join(repoRoot, types.PierConfigRepo, pierConfigMap[version], types.PierConfigScript), repoRoot, target, configPath, chainType, binPath, pluginPath, appchainConfigPath)
}

// TODO: delete
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
