package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/coreos/etcd/pkg/fileutil"
	"github.com/gobuffalo/packd"
	"github.com/gobuffalo/packr"
	"github.com/meshplus/goduck/repo"
	"github.com/urfave/cli/v2"
)

const (
	ScriptsPath = "../../scripts"
	ConfigPath  = "../../config"
)

func GetInitCMD() *cli.Command {
	return &cli.Command{
		Name:   "init",
		Usage:  "init config home for goduck",
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
		fmt.Println("goduck configuration file already exists")
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

	return nil
}
