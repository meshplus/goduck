package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/codeskyblue/go-sh"
	"github.com/fatih/color"
	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/internal/download"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
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
						Name:     "ips",
						Usage:    "servers ip, e.g. 188.0.0.1,188.0.0.2,188.0.0.3,.188.0.0.4",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "username,u",
						Usage:    "server username",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "version,v",
						Usage:    "BitXHub version",
						Required: true,
					},
				},
				Action: deployBitXHub,
			},
			{
				Name:  "pier",
				Usage: "Deploy pier to remote server",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "mode",
						Value: types.PierModeRelay,
						Usage: "configuration mode, one of direct or relay",
					},
					&cli.IntFlag{
						Name:  "pprof-port",
						Value: 44550,
						Usage: "pier pprof port",
					},
					&cli.StringFlag{
						Name:     "chain",
						Usage:    "specify appchain type, ethereum or fabric (default: \"ethereum\")",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "cryptoPath",
						Usage:    "path of crypto-config on server, only useful for fabric chain",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "bitxhub",
						Usage:    "BitXHub's address, only useful in relay mode",
						Value:    "10.1.16.48:60011",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "validators",
						Usage:    "BitXHub's validators, only useful in relay mode",
						Value:    "0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013 0x79a1215469FaB6f9c63c1816b45183AD3624bE34 0x97c8B516D19edBf575D72a172Af7F418BE498C37 0xc0Ff2e0b3189132D815b8eb325bE17285AC898f8",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "peers",
						Usage:    "peers' address, only useful in direct mode",
						Value:    "",
						Required: false,
					},
					&cli.IntFlag{
						Name:     "port",
						Value:    5001,
						Usage:    "pier's port, only useful in direct mode",
						Required: false,
					},
					&cli.IntFlag{
						Name:     "api-port",
						Value:    8080,
						Usage:    "peer's api port",
						Required: false,
					},
					&cli.StringFlag{
						Name:     "ip",
						Usage:    "servers ip, e.g. 188.0.0.1",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "username,u",
						Usage:    "server username",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "version,v",
						Usage:    "pier version",
						Required: true,
					},
				},
				Action: deployPier,
			},
		},
	}
}

func deployBitXHub(ctx *cli.Context) error {
	repoRoot, err := repo.PathRoot()
	if err != nil {
		return err
	}

	username := ctx.String("username")
	version := ctx.String("version")

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

	dir, err := ioutil.TempDir("", "bitxhub")
	if err != nil {
		return err
	}

	ips := strings.Split(ctx.String("ips"), ",")

	generator := NewBitXHubConfigGenerator("binary", "cluster", dir, len(ips), ips, false, version)

	if err := generator.InitConfig(); err != nil {
		return err
	}

	binPath := filepath.Join(repoRoot, fmt.Sprintf("bin/bitxhub_linux_%s", version))

	err = os.MkdirAll(binPath, os.ModePerm)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("bitxhub_linux-amd64_%s.tar.gz", version)
	filePath := filepath.Join(binPath, filename)

	if !fileutil.Exist(filePath) {
		url := fmt.Sprintf("https://github.com/meshplus/bitxhub/releases/download/%s/bitxhub_linux-amd64_%s.tar.gz", version, version)
		err = download.Download(binPath, url)
		if err != nil {
			return err
		}
	}

	for idx, ip := range ips {
		color.Blue("====> Operating at node%d\n", idx+1)
		who := fmt.Sprintf("%s@%s", username, ip)
		target := fmt.Sprintf("%s:~/", who)

		err = sh.Command("ssh", who, fmt.Sprintf("mkdir -p ~/.bitxhub/node%d", idx+1)).
			Command("scp", filePath, target).Run()
		if err != nil {
			return err
		}

		err = sh.Command("scp", "-r",
			fmt.Sprintf("%s/node%d", dir, idx+1),
			fmt.Sprintf("%s%s", target, ".bitxhub")).
			Run()
		if err != nil {
			return err
		}

		err = sh.
			Command("ssh", who, fmt.Sprintf("tar xzf %s -C ~/.bitxhub --strip-components 1 && mkdir -p .bitxhub/node%d/plugins && cp .bitxhub/raft.so .bitxhub/node%d/plugins", filename, idx+1, idx+1)).Run()
		if err != nil {
			return err
		}
	}

	color.Blue("====> Run\n")
	for idx, ip := range ips {
		who := fmt.Sprintf("%s@%s", username, ip)

		err = sh.Command("ssh", who,
			fmt.Sprintf("export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:$HOME/.bitxhub && cd ~/.bitxhub && nohup ./bitxhub --repo=node%d start >/dev/null 2>&1 &", idx+1)).Start()
		if err != nil {
			return err
		}

		err = sh.Command("ssh", who,
			fmt.Sprintf("sleep 1 && NODE=`lsof -i:6001%d | grep LISTEN` && NODE=${NODE#* } && echo ${NODE%%%% *} > ~/.bitxhub/bitxhub%d.PID", idx+1, idx+1)).Start()
		if err != nil {
			color.Red("Start BitXHub node%d fail\n", idx+1)
			return err
		} else {
			color.Blue("Start BitXHub node%d end\n", idx+1)
		}
	}

	color.Blue("====> Check\n")
	fmt.Println("You need to wait more than 5 seconds for each node")
	for idx, ip := range ips {
		who := fmt.Sprintf("%s@%s", username, ip)

		out, err := sh.Command("ssh", who, fmt.Sprintf("sleep 5 && cat ~/.bitxhub/bitxhub%d.PID", idx+1)).Output()
		if err != nil {
			return err
		}

		pid := strings.Replace(string(out), "\n", "", -1)

		if pid != "" {
			out, err = sh.Command("ssh", who, fmt.Sprintf("if [[ -n `ps aux | grep %s | grep -v grep | grep node%d` ]]; then echo successful; else echo fail; fi", pid, idx+1)).Output()
			if err != nil {
				return err
			}

			res := strings.Replace(string(out), "\n", "", -1)
			if res == "successful" {
				color.Green("Start BitXHub node%d successful\n", idx+1)
			} else {
				color.Red("Start BitXHub node%d fail\n", idx+1)
			}
		} else {
			color.Red("Start BitXHub node%d fail\n", idx+1)
		}
	}
	return nil
}

func deployPier(ctx *cli.Context) error {
	repoRoot, err := repo.PathRoot()
	if err != nil {
		return err
	}

	mode := ctx.String("mode")
	pprof := ctx.Int("pprof-port")
	chain := ctx.String("chain")
	cryptoPath := ctx.String("cryptoPath")
	if chain == "fabric" {
		if cryptoPath == "" {
			return fmt.Errorf("start fabric pier need crypto-config path")
		}
	}

	bitxhub := ctx.String("bitxhub")
	validators := strings.Fields(ctx.String("validators"))
	peers := strings.Fields(ctx.String("peers"))
	port := ctx.Int("port")
	apiPort := ctx.Int("api-port")
	ip := ctx.String("ip")
	username := ctx.String("username")
	version := ctx.String("version")

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

	who := fmt.Sprintf("%s@%s", username, ip)
	target := fmt.Sprintf("%s:~/", who)

	err = pierPrepare(repoRoot, version, target, who, mode, bitxhub, chain, ip, validators, peers, port, apiPort, pprof, cryptoPath)
	if err != nil {
		return err
	}

	err = appchainRegister(who, chain)
	if err != nil {
		return err
	}

	err = ruleDeploy(who, chain)
	if err != nil {
		return err
	}

	err = pierStartRemote(who, chain)
	if err != nil {
		return err
	}

	err = pierCheck(who, chain, pprof)
	if err != nil {
		return err
	}

	return nil
}

func pierCheck(who string, chain string, pprof int) error {
	color.Blue("===> Check pier of %s\n", chain)
	fmt.Println("You need to wait more than 15 seconds")

	err := sh.Command("ssh", who, fmt.Sprintf("sleep 15 && echo `lsof -i:%d | grep LISTEN` | awk '{print $2}' > ~/.pier_%s/pier.PID", pprof, chain)).Run()
	if err != nil {
		return err
	}

	out, err := sh.Command("ssh", who, fmt.Sprintf("cat ~/.pier_%s/pier.PID", chain)).Output()
	if err != nil {
		return err
	}

	pid := strings.Replace(string(out), "\n", "", -1)

	if pid != "" {
		out, err = sh.Command("ssh", who, fmt.Sprintf("if [[ -n `ps aux | grep %s | grep -v grep | grep .pier_%s` ]]; then echo successful; else echo fail; fi", pid, chain)).Output()
		if err != nil {
			return err
		}
		res := strings.Replace(string(out), "\n", "", -1)
		if res == "successful" {
			color.Green("Start pier successful\n")
		} else {
			color.Red("Start pier fail\n")
		}
	} else {
		color.Red("Start pier fail\n")
	}

	return nil
}

func pierStartRemote(who string, chain string) error {
	color.Blue("===> Start pier of %s\n", chain)
	err := sh.
		Command("ssh", who, fmt.Sprintf("export LD_LIBRARY_PATH=$HOME/pier && export CONFIG_PATH=$HOME/.pier_fabric/fabric && nohup $HOME/pier/pier --repo $HOME/.pier_%s start >/dev/null 2>&1 &", chain)).Start()
	if err != nil {
		color.Red("Start pier fail\n")
		return err
	}

	color.Blue("Start pier end\n")
	return nil
}

func ruleDeploy(who string, chain string) error {
	color.Blue("====> Deploy rule in bitxhub\n")
	err := sh.Command("ssh", who,
		fmt.Sprintf("export LD_LIBRARY_PATH=$HOME/pier && $HOME/pier/pier --repo $HOME/.pier_%s rule deploy --path $HOME/.pier_%s/%s/%s_rule.wasm", chain, chain, chain, chain)).Run()
	if err != nil {
		return err
	}
	return nil
}

func appchainRegister(who, chain string) error {
	color.Blue("====> Register pier(%s) to BitXHub \n", chain)

	chainVersion := ""
	if chain == "fabric" {
		chainVersion = "1.4.3"
	} else if chain == "ethereum" {
		chainVersion = "1.9.13"
	} else {
		return fmt.Errorf("not support chain type")
	}

	err := sh.Command("ssh", who,
		fmt.Sprintf("export LD_LIBRARY_PATH=$HOME/pier && $HOME/pier/pier --repo $HOME/.pier_%s appchain register "+
			"--name chain-%s "+
			"--type %s "+
			"--desc chain-%s-description "+
			"--version %s "+
			"--validators $HOME/.pier_%s/%s/%s.validators",
			chain, chain, chain, chain, chainVersion, chain, chain, chain)).Run()
	if err != nil {
		return err
	}

	return nil
}

func pierPrepare(repoRoot, version, target, who, mode, bitxhub, chain, ip string, validators, peers []string, port, apiPort, pprof int, cryptoPath string) error {
	configPath := filepath.Join(repoRoot, "pier_deploy/.")
	err := os.MkdirAll(configPath, os.ModePerm)
	if err != nil {
		return err
	}

	binPath := filepath.Join(repoRoot, fmt.Sprintf("bin/pier_linux_%s", version))
	err = os.MkdirAll(binPath, os.ModePerm)
	if err != nil {
		return err
	}

	filename := fmt.Sprintf("pier_linux-amd64_%s.tar.gz", version)
	filePath := filepath.Join(binPath, filename)
	if !fileutil.Exist(filePath) {
		url := fmt.Sprintf(types.PierUrlLinux, version, version)
		err = download.Download(binPath, url)
		if err != nil {
			return err
		}
	}

	libPath := filepath.Join(binPath, "libwasmer.so")
	if !fileutil.Exist(libPath) {
		err = download.Download(binPath, types.LinuxWasmLibUrl)
		if err != nil {
			return err
		}
	}

	ruleName := fmt.Sprintf("%s_rule.wasm", chain)
	rulePath := filepath.Join(binPath, ruleName)
	if !fileutil.Exist(rulePath) {
		if chain == "fabric" {
			err = download.Download(rulePath, types.FabricRuleUrl)
		} else {
			err = download.Download(rulePath, types.EthereumRuleUrl)
		}
		if err != nil {
			return err
		}
	}

	color.Blue("====> Generate pier configure locally\n")
	err = InitPierConfig(mode, bitxhub, chain, ip, configPath, validators, peers, port, pprof, apiPort, version)
	if err != nil {
		return err
	}

	color.Blue("====> Uploading files to server %s\n", ip)
	err = sh.Command("ssh", who, "if [ ! -d pier ]; then mkdir pier; fi && cd pier").Run()
	if err != nil {
		return err
	}
	// The files needed for deployment are placed in the ~/pier folder, and the configuration folder for actual deployment is ~/.pier_${chaintype}.
	err = sh.
		Command("scp", libPath, fmt.Sprintf("%spier/", target)).
		Command("scp", rulePath, fmt.Sprintf("%spier/", target)).
		Command("scp", "-r", configPath, fmt.Sprintf("%s.pier_%s/", target, chain)).
		Command("scp", filePath, fmt.Sprintf("%spier/", target)).
		Run()
	if err != nil {
		return err
	}
	err = sh.
		Command("ssh", who, fmt.Sprintf("cd $HOME/pier && tar xf %s -C $HOME/pier --strip-components 1", filename)).
		Run()
	if err != nil {
		return err
	}

	color.Blue("====> Update key\n")
	err = sh.Command("ssh", who, fmt.Sprintf("if [ ! -d $HOME/.pier_%s/tmp ]; then mkdir $HOME/.pier_%s/tmp; fi && export LD_LIBRARY_PATH=$HOME/pier && $HOME/pier/pier --repo $HOME/.pier_%s/tmp init && mv $HOME/.pier_%s/tmp/key.json $HOME/.pier_%s/key.json && rm -r $HOME/.pier_%s/tmp", chain, chain, chain, chain, chain, chain)).Run()
	if err != nil {
		return err
	}

	color.Blue("====> Copy appchain plugin\n")
	chainPlugin := ""
	if chain == "fabric" {
		if version == "v1.0.0" || version == "v1.0.0-rc1" {
			chainPlugin = "fabric-client-1.4.so"
		} else {
			chainPlugin = "fabric-client-1.4"
		}
	} else {
		if version == "v1.0.0" || version == "v1.0.0-rc1" {
			chainPlugin = "eth-client.so"
		} else {
			chainPlugin = "eth-client"
		}
	}

	err = sh.
		Command("ssh", who, fmt.Sprintf("mkdir -p $HOME/.pier_%s/plugins && cp $HOME/pier/%s $HOME/.pier_%s/plugins/", chain, chainPlugin, chain)).
		Run()
	if err != nil {
		return err
	}

	if version != "v1.0.0" && version != "v1.0.0-rc1" {
		err = sh.Command("ssh", who, fmt.Sprintf("mv $HOME/.pier_%s/plugins/%s $HOME/.pier_%s/plugins/appchain_plugin", chain, chainPlugin, chain)).Run()
		if err != nil {
			return err
		}
	}

	if chain == "fabric" {
		err = sh.
			Command("ssh", who, fmt.Sprintf("cp -r %s $HOME/.pier_%s/fabric/", cryptoPath, chain)).
			Run()
		if err != nil {
			return err
		}
		err = sh.Command("ssh", who, fmt.Sprintf("cp $HOME/.pier_%s/fabric/crypto-config/peerOrganizations/org2.example.com/peers/peer1.org2.example.com/msp/signcerts/peer1.org2.example.com-cert.pem $HOME/.pier_%s/fabric/fabric.validators", chain, chain)).Run()
		if err != nil {
			return err
		}
	} else if chain == "ethereum" {
		err = sh.Command("ssh", who, "if [ ! -e $HOME/.pier_ethereum/ethereum/ethereum.validators ]; then mv $HOME/.pier_ethereum/ethereum/ether.validators $HOME/.pier_ethereum/ethereum/ethereum.validators; fi").Run()
		if err != nil {
			return err
		}
	}

	color.Blue("====> Copy rule\n")
	err = sh.
		Command("ssh", who, fmt.Sprintf("mv $HOME/pier/%s_rule.wasm $HOME/.pier_%s/%s/%s_rule.wasm", chain, chain, chain, chain)).
		Run()

	color.Green("====> Pier_root: $HOME/.pier_%s, BitXHub_addr: %s\n", chain, bitxhub)
	return nil
}
