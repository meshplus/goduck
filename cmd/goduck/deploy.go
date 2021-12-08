package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"

	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/cmd/goduck/bitxhub"
	"github.com/meshplus/goduck/cmd/goduck/pier"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
	"github.com/urfave/cli/v2"
)

func deployCMD() *cli.Command {
	return &cli.Command{
		Name:  "deploy",
		Usage: "Deploy BitXHub and pier",
		Subcommands: []*cli.Command{
			{
				Name:  "bitxhub",
				Usage: "Deploy BitXHub to remote server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "target",
						Usage:    "Specify the directory to where to put the generated configuration files",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "configPath",
						Usage: "Specify the configuration file path for the configuration to be modified, default: $repo/bxh_config/$version/bxh_modify_config.toml",
					},
					&cli.StringFlag{
						Aliases: []string{"version", "v"},
						Value:   "v1.11.0",
						Usage:   "BitXHub version",
					},
				},
				Action: deployBitXHub,
			},
			{
				Name:  "pier",
				Usage: "Deploy pier to remote server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "appchain",
						Usage: "Specify appchain type, one of ethereum or fabric",
						Value: types.ChainTypeEther,
					},
					&cli.StringFlag{
						Name:     "target",
						Usage:    "Specify the directory to where to put the generated configuration files",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "configPath",
						Usage: "Specify the configuration file path for the configuration to be modified, default: $repo/pier_config/$VERSION/pier_modify_config.toml",
					},
					&cli.StringFlag{
						Aliases: []string{"version", "v"},
						Value:   "v1.11.0",
						Usage:   "BitXHub version",
					},
				},
				Action: deployPier,
			},
		},
	}
}

func deployBitXHub(ctx *cli.Context) error {
	target := ctx.String("target")
	configPath := ctx.String("configPath")
	version := ctx.String("version")

	// 1. check goduck init
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return fmt.Errorf("parse repo path error:%w", err)
	}
	if !fileutil.Exist(filepath.Join(repoRoot, types.ReleaseJson)) {
		return fmt.Errorf("please `goduck init` first")
	}

	// 2. check versiojn
	data, err := ioutil.ReadFile(filepath.Join(repoRoot, "release.json"))
	if err != nil {
		return err
	}

	var release *Release
	if err := json.Unmarshal(data, &release); err != nil {
		return err
	}

	if !AdjustVersion(version, release.Bitxhub) {
		return fmt.Errorf("unsupport BitXHub verison")
	}

	// 3. get args and bin
	if configPath == "" {
		configPath = filepath.Join(repoRoot, fmt.Sprintf("bxh_config/%s/%s", bxhConfigMap[version], types.BxhModifyConfig))
	} else {
		configPath, err = filepath.Abs(configPath)
		if err != nil {
			return fmt.Errorf("get absolute config path: %w", err)
		}
	}

	err = bitxhub.DownloadBitxhubBinary(repoRoot, version, runtime.GOOS)
	if err != nil {
		return fmt.Errorf("download %s binary error:%w", runtime.GOOS, err)
	}

	// the remote system is usually Linux
	err = bitxhub.DownloadBitxhubBinary(repoRoot, version, types.LinuxSystem)
	if err != nil {
		return fmt.Errorf("download %s binary error:%w", types.LinuxSystem, err)
	}

	// 4. execute
	args := make([]string, 0)
	args = append(args, filepath.Join(repoRoot, types.DeployBxhScript), "up")
	args = append(args, version, configPath, target)
	return utils.ExecuteShell(args, repoRoot)
}

func deployPier(ctx *cli.Context) error {
	appchain := ctx.String("appchain")
	target := ctx.String("target")
	configPath := ctx.String("configPath")
	version := ctx.String("version")

	// 1. check goduck init
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return fmt.Errorf("parse repo path error:%w", err)
	}
	if !fileutil.Exist(filepath.Join(repoRoot, types.ReleaseJson)) {
		return fmt.Errorf("please `goduck init` first")
	}

	// 2. check versiojn
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

	// 3. get args and bin
	if configPath == "" {
		configPath = filepath.Join(repoRoot, fmt.Sprintf("pier_config/%s/%s", pierConfigMap[version], types.PierModifyConfig))
	} else {
		configPath, err = filepath.Abs(configPath)
		if err != nil {
			return fmt.Errorf("get absolute config path: %w", err)
		}
	}

	err = pier.DownloadPierBinary(repoRoot, version, runtime.GOOS)
	if err != nil {
		return fmt.Errorf("download %s binary error:%w", runtime.GOOS, err)
	}

	// the remote system is usually Linux
	err = pier.DownloadPierBinary(repoRoot, version, types.LinuxSystem)
	if err != nil {
		return fmt.Errorf("download %s binary error:%w", types.LinuxSystem, err)
	}

	if err := pier.DownloadPierPlugin(repoRoot, appchain, version, types.LinuxSystem); err != nil {
		return fmt.Errorf("download pier binary error:%w", err)
	}

	// 4. execute
	args := make([]string, 0)
	args = append(args, filepath.Join(repoRoot, types.DeployPierScript), "up")
	args = append(args, appchain, version, configPath, target)
	return utils.ExecuteShell(args, repoRoot)
}

//func pierCheck(who string, chain string, pprof string) error {
//	color.Blue("===> Check pier of %s\n", chain)
//	fmt.Println("You need to wait more than 15 seconds")
//
//	err := sh.Command("ssh", who, fmt.Sprintf("sleep 15 && echo `lsof -i:%s | grep LISTEN` | awk '{print $2}' > ~/.pier_%s/pier.PID", pprof, chain)).Run()
//	if err != nil {
//		return err
//	}
//
//	out, err := sh.Command("ssh", who, fmt.Sprintf("cat ~/.pier_%s/pier.PID", chain)).Output()
//	if err != nil {
//		return err
//	}
//
//	pid := strings.Replace(string(out), "\n", "", -1)
//
//	if pid != "" {
//		out, err = sh.Command("ssh", who, fmt.Sprintf("if [[ -n `ps aux | grep %s | grep -v grep | grep .pier_%s` ]]; then echo successful; else echo fail; fi", pid, chain)).Output()
//		if err != nil {
//			return err
//		}
//		res := strings.Replace(string(out), "\n", "", -1)
//		if res == "successful" {
//			color.Green("Start pier successful\n")
//		} else {
//			color.Red("Start pier fail\n")
//		}
//	} else {
//		color.Red("Start pier fail\n")
//	}
//
//	return nil
//}
//
//func pierStartRemote(who string, chain string) error {
//	color.Blue("===> Start pier of %s\n", chain)
//	err := sh.
//		Command("ssh", who, fmt.Sprintf("export LD_LIBRARY_PATH=$HOME/pier && export CONFIG_PATH=$HOME/.pier_fabric/fabric && nohup $HOME/pier/pier --repo $HOME/.pier_%s start >/dev/null 2>&1 &", chain)).Start()
//	if err != nil {
//		color.Red("Start pier fail\n")
//		return err
//	}
//
//	color.Blue("Start pier end\n")
//	return nil
//}
//
//func ruleDeploy(who string, chain string) error {
//	color.Blue("====> Deploy rule in bitxhub\n")
//	chainRepo := chain
//	if chain == types.ChainTypeEther {
//		chainRepo = "ether"
//	}
//	err := sh.Command("ssh", who,
//		fmt.Sprintf("export LD_LIBRARY_PATH=$HOME/pier && $HOME/pier/pier --repo $HOME/.pier_%s rule deploy --path $HOME/.pier_%s/%s/validating.wasm", chain, chain, chainRepo)).Run()
//	if err != nil {
//		return err
//	}
//	return nil
//}
//
//func appchainRegister(who, chain, version string) error {
//	color.Blue("====> Register pier(%s) to BitXHub \n", chain)
//
//	chainVersion := ""
//	if chain == "fabric" {
//		chainVersion = "1.4.3"
//	} else if chain == "ethereum" {
//		chainVersion = "1.9.13"
//	} else {
//		return fmt.Errorf("not support chain type")
//	}
//
//	if version < "v1.7.0" {
//		err := sh.Command("ssh", who,
//			fmt.Sprintf("export LD_LIBRARY_PATH=$HOME/pier && $HOME/pier/pier --repo $HOME/.pier_%s appchain register "+
//				"--name chain-%s "+
//				"--type %s "+
//				"--desc chain-%s-description "+
//				"--version %s "+
//				"--validators $HOME/.pier_%s/%s/%s.validators",
//				chain, chain, chain, chain, chainVersion, chain, chain, chain)).Run()
//		if err != nil {
//			return err
//		}
//	} else {
//		err := sh.Command("ssh", who,
//			fmt.Sprintf("export LD_LIBRARY_PATH=$HOME/pier && $HOME/pier/pier --repo $HOME/.pier_%s appchain register "+
//				"--name chain-%s "+
//				"--type %s "+
//				"--desc chain-%s-description "+
//				"--version %s "+
//				"--validators $HOME/.pier_%s/%s/%s.validators "+
//				"--consensusType consensusType",
//				chain, chain, chain, chain, chainVersion, chain, chain, chain)).Run()
//		if err != nil {
//			return err
//		}
//	}
//
//	return nil
//}
//
//func pierPrepare(repoRoot, version, target, who, mode, bitxhub, chain, ip string, validators []string, port string, peers, connectors []string, providers, tls, http, pprof, apiPort, cryptoPath, appchainIP, appchainAddr string, appPorts []string, appchainContractAddr, appchainDid string) error {
//	configPath := filepath.Join(repoRoot, "pier_deploy")
//	err := os.MkdirAll(configPath, os.ModePerm)
//	if err != nil {
//		return err
//	}
//
//	binPath := filepath.Join(repoRoot, fmt.Sprintf("bin/pier_linux_%s", version))
//	err = os.MkdirAll(binPath, os.ModePerm)
//	if err != nil {
//		return err
//	}
//
//	filename := fmt.Sprintf("pier_linux-amd64_%s.tar.gz", version)
//	filePath := filepath.Join(binPath, filename)
//	if !fileutil.Exist(filePath) {
//		url := fmt.Sprintf(types.PierUrlLinux, version, version)
//		err = download.Download(binPath, url)
//		if err != nil {
//			return err
//		}
//	}
//
//	if !fileutil.Exist(filePath) {
//		url := fmt.Sprintf(types.PierUrlLinux, version, version)
//		err = download.Download(binPath, url)
//		if err != nil {
//			return err
//		}
//	}
//	if chain == types.ChainTypeFabric {
//		if !fileutil.Exist(filepath.Join(binPath, types.FabricClient)) {
//			url := fmt.Sprintf(types.PierFabricClientUrlLinux, version, version)
//			err := download.Download(binPath, url)
//			if err != nil {
//				return err
//			}
//
//			err = sh.Command("/bin/bash", "-c", fmt.Sprintf("cd %s && mv fabric-client-%s-Linux %s && chmod +x %s", binPath, version, types.FabricClient, types.FabricClient)).Run()
//			if err != nil {
//				return fmt.Errorf("rename fabric client error: %s", err)
//			}
//		}
//	} else if chain == types.ChainTypeEther {
//		if !fileutil.Exist(filepath.Join(binPath, types.EthClient)) {
//			url := fmt.Sprintf(types.PierEthereumClientUrlLinux, version, version)
//			err := download.Download(binPath, url)
//			if err != nil {
//				return err
//			}
//
//			err = sh.Command("/bin/bash", "-c", fmt.Sprintf("cd %s && mv eth-client-%s-Linux %s && chmod +x %s", binPath, version, types.EthClient, types.EthClient)).Run()
//			if err != nil {
//				return fmt.Errorf("rename eth client error: %s", err)
//			}
//		}
//	}
//
//	libPath := filepath.Join(binPath, "libwasmer.so")
//	if !fileutil.Exist(libPath) {
//		err = download.Download(binPath, types.LinuxWasmLibUrl)
//		if err != nil {
//			return err
//		}
//	}
//
//	rulePath := filepath.Join(binPath, types.RuleName)
//	if !fileutil.Exist(rulePath) {
//		if chain == "fabric" {
//			err = download.Download(rulePath, types.FabricRuleUrl)
//		} else {
//			err = download.Download(rulePath, types.EthereumRuleUrl)
//		}
//		if err != nil {
//			return err
//		}
//	}
//
//	color.Blue("====> Generate pier configure locally\n")
//	pierPath := ""
//	err = InitPierConfig(mode, "binary", bitxhub, validators, port, peers, connectors, providers, chain, appchainIP, appchainAddr, appPorts, appchainContractAddr, configPath, tls, http, pprof, apiPort, version, pierPath, cryptoPath, appchainDid)
//	if err != nil {
//		return err
//	}
//
//	color.Blue("====> Uploading files to server %s\n", ip)
//	err = sh.Command("ssh", who, "if [ ! -d pier ]; then mkdir pier; fi && cd pier").Run()
//	if err != nil {
//		return err
//	}
//	err = sh.Command("ssh", who, fmt.Sprintf("if [ -d $HOME/pier/pier_deploy ]; then rm -r $HOME/pier/pier_deploy; fi")).Run()
//	if err != nil {
//		return err
//	}
//	// The files needed for deployment are placed in the ~/pier folder, and the configuration folder for actual deployment is ~/.pier_${chaintype}.
//	if version < "v1.7.0" {
//		err = sh.
//			Command("scp", libPath, fmt.Sprintf("%spier/", target)).
//			Command("scp", rulePath, fmt.Sprintf("%spier/", target)).
//			Command("scp", "-r", configPath, fmt.Sprintf("%spier/pier_deploy", target)).
//			Command("scp", filePath, fmt.Sprintf("%spier/", target)).
//			Run()
//		if err != nil {
//			return err
//		}
//	} else {
//		clientPath := ""
//		if chain == types.ChainTypeEther {
//			clientPath = filepath.Join(binPath, types.EthClient)
//		} else if chain == types.ChainTypeFabric {
//			clientPath = filepath.Join(binPath, types.FabricClient)
//		}
//		err = sh.
//			Command("scp", libPath, fmt.Sprintf("%spier/", target)).
//			Command("scp", rulePath, fmt.Sprintf("%spier/", target)).
//			Command("scp", "-r", configPath, fmt.Sprintf("%spier/pier_deploy", target)).
//			Command("scp", clientPath, fmt.Sprintf("%spier/", target)).
//			Command("scp", filePath, fmt.Sprintf("%spier/", target)).
//			Run()
//		if err != nil {
//			return err
//		}
//	}
//	err = sh.
//		Command("ssh", who, fmt.Sprintf("cd $HOME/pier && tar xf %s -C $HOME/pier --strip-components 1", filename)).
//		Run()
//	if err != nil {
//		return err
//	}
//
//	color.Blue("====> Update config\n")
//	err = sh.Command("ssh", who, fmt.Sprintf("export LD_LIBRARY_PATH=$HOME/pier && $HOME/pier/pier --repo $HOME/.pier_%s init", chain)).Run()
//	if err != nil {
//		return err
//	}
//
//	err = sh.Command("ssh", who, fmt.Sprintf("cp -r $HOME/pier/pier_deploy/. $HOME/.pier_%s", chain)).Run()
//	if err != nil {
//		return err
//	}
//
//	if mode == types.PierModeDirect {
//		out, err := sh.Command("ssh", who, fmt.Sprintf("export LD_LIBRARY_PATH=$HOME/pier && $HOME/pier/pier --repo $HOME/.pier_%s/tmp p2p id", chain)).Output()
//		if err != nil {
//			return fmt.Errorf("get pier id: %s", err)
//		}
//		pid := strings.Replace(string(out), "\n", "", -1)
//		err = sh.Command("ssh", who, fmt.Sprintf("sed -i \"s#\"/ip4/0.0.0.0/tcp/%s/p2p/\"#\"/ip4/0.0.0.0/tcp/%s/p2p/%s\"#g\" $HOME/.pier_%s/%s", port, port, pid, chain, repo.PierConfigName)).Run()
//		if err != nil {
//			return err
//		}
//	}
//
//	color.Blue("====> Copy appchain plugin\n")
//	chainPlugin := ""
//	if chain == "fabric" {
//		if version == "v1.0.0" || version == "v1.0.0-rc1" {
//			chainPlugin = "fabric-client-1.4.so"
//		} else {
//			chainPlugin = "fabric-client-1.4"
//		}
//	} else {
//		if version == "v1.0.0" || version == "v1.0.0-rc1" {
//			chainPlugin = "eth-client.so"
//		} else {
//			chainPlugin = "eth-client"
//		}
//	}
//
//	err = sh.
//		Command("ssh", who, fmt.Sprintf("mkdir -p $HOME/.pier_%s/plugins && cp $HOME/pier/%s $HOME/.pier_%s/plugins/", chain, chainPlugin, chain)).
//		Run()
//	if err != nil {
//		return err
//	}
//
//	if version != "v1.0.0" && version != "v1.0.0-rc1" {
//		err = sh.Command("ssh", who, fmt.Sprintf("mv $HOME/.pier_%s/plugins/%s $HOME/.pier_%s/plugins/appchain_plugin", chain, chainPlugin, chain)).Run()
//		if err != nil {
//			return err
//		}
//	}
//
//	if chain == "fabric" {
//		err = sh.
//			Command("ssh", who, fmt.Sprintf("cp -r %s $HOME/.pier_%s/fabric/", cryptoPath, chain)).
//			Run()
//		if err != nil {
//			return err
//		}
//		err = sh.Command("ssh", who, fmt.Sprintf("cp $HOME/.pier_%s/fabric/crypto-config/peerOrganizations/org2.example.com/peers/peer1.org2.example.com/msp/signcerts/peer1.org2.example.com-cert.pem $HOME/.pier_%s/fabric/fabric.validators", chain, chain)).Run()
//		if err != nil {
//			return err
//		}
//	} else if chain == "ethereum" {
//		err = sh.Command("ssh", who, "if [ ! -e $HOME/.pier_ethereum/ethereum/ethereum.validators ]; then mv $HOME/.pier_ethereum/ethereum/ether.validators $HOME/.pier_ethereum/ethereum/ethereum.validators; fi").Run()
//		if err != nil {
//			return err
//		}
//	}
//
//	color.Blue("====> Copy rule\n")
//	err = sh.
//		Command("ssh", who, fmt.Sprintf("mv $HOME/pier/validating.wasm $HOME/.pier_%s/%s/validating.wasm", chain, chain)).
//		Run()
//
//	color.Green("====> Pier_root: $HOME/.pier_%s, BitXHub_addr: %s\n", chain, bitxhub)
//	return nil
//}
