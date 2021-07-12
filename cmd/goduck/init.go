package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	"github.com/codeskyblue/go-sh"
	"github.com/coreos/etcd/pkg/fileutil"
	"github.com/fatih/color"
	"github.com/gobuffalo/packd"
	"github.com/gobuffalo/packr"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/urfave/cli/v2"
)

const (
	ScriptsPath = "../../scripts"
	ConfigPath  = "../../config"
)

func GetInitCMD() *cli.Command {
	return &cli.Command{
		Name:   "init",
		Usage:  "Init config home for GoDuck",
		Action: Initialize,
	}
}

func Initialize(ctx *cli.Context) error {
	repoRoot := ctx.String("repo")
	if repoRoot == "" {
		root, err := repo.PathRoot()
		if err != nil {
			return err
		}
		repoRoot = root
	}
	if fileutil.Exist(repoRoot) {
		fmt.Println("GoDuck configuration file already exists")
		fmt.Println("reinitializing would overwrite your configuration, Y/N?")
		input := bufio.NewScanner(os.Stdin)
		input.Scan()
		if input.Text() != "Y" && input.Text() != "y" {
			return nil
		}
	}
	scriptBox := packr.NewBox(ScriptsPath)
	configBox := packr.NewBox(ConfigPath)

	var walkFn = func(s string, file packd.File) error {
		p := filepath.Join(repoRoot, s)
		dir := filepath.Dir(p)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				return err
			}
		}
		return ioutil.WriteFile(p, []byte(file.String()), 0644)
	}

	if err := scriptBox.Walk(walkFn); err != nil {
		return err
	}

	if err := configBox.Walk(walkFn); err != nil {
		return err
	}

	if err := ModifyConfigInit(repoRoot); err != nil {
		return err
	}

	color.Green("Init goduck successfully in %s!\n", repoRoot)
	return nil
}

func ModifyConfigInit(repo string) error {
	var repoTmp string
	for _, c := range repo {
		if c == '/' {
			cm := "\\/"
			repoTmp += cm
		} else {
			repoTmp += string(c)
		}
	}

	if runtime.GOOS == types.DarwinSystem {
		for _, v := range bxhConfigMap {
			bxhConfigPath := filepath.Join(repo, types.BxhConfigRepo, v, types.BxhModifyConfig)
			err := sh.Command("/bin/bash", "-c", fmt.Sprintf("sed -i '' \"s/REPO/%s/g\" %s", repoTmp, bxhConfigPath)).Run()
			if err != nil {
				return err
			}
		}

		for _, v := range pierConfigMap {
			pierConfigPath := filepath.Join(repo, types.PierConfigRepo, v, types.PierModifyConfig)
			err := sh.Command("/bin/bash", "-c", fmt.Sprintf("sed -i '' \"s/REPO/%s/g\" %s", repoTmp, pierConfigPath)).Run()
			if err != nil {
				return err
			}
		}
	} else if runtime.GOOS == types.LinuxSystem {
		for _, v := range bxhConfigMap {
			bxhConfigPath := filepath.Join(repo, types.BxhConfigRepo, v, types.BxhModifyConfig)
			err := sh.Command("/bin/bash", "-c", fmt.Sprintf("sed -i \"s/REPO/%s/g\" %s", repoTmp, bxhConfigPath)).Run()
			if err != nil {
				return err
			}
		}

		for _, v := range pierConfigMap {
			pierConfigPath := filepath.Join(repo, types.PierConfigRepo, v, types.PierModifyConfig)
			err := sh.Command("/bin/bash", "-c", fmt.Sprintf("sed -i \"s/REPO/%s/g\" %s", repoTmp, pierConfigPath)).Run()
			if err != nil {
				return err
			}
		}
	} else {
		color.Red("Unsupported system!")
	}

	return nil
}
