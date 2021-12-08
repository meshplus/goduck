package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/meshplus/goduck/internal/utils"
	"github.com/urfave/cli/v2"
)

func getVersionCMD() *cli.Command {
	return &cli.Command{
		Name:  "version",
		Usage: "Components version",
		Subcommands: []*cli.Command{
			{
				Name:   "all",
				Usage:  "All components version",
				Action: allVersion,
			},
			{
				Name:   "goduck",
				Usage:  "Goduck version",
				Action: goduckVersion,
			},
			{
				Name:  "bitxhub",
				Usage: "BitXHub version",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "upType",
						Usage: "Specify the startup type, one of binary or docker",
						Value: types.TypeBinary,
					},
				},
				Action: bxhVersion,
			},
			{
				Name:  "pier",
				Usage: "Pier version",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "appchain",
						Usage: "Specify appchain type, one of ethereum or fabric",
						Value: types.ChainTypeEther,
					},
					&cli.StringFlag{
						Name:  "upType",
						Usage: "Specify the startup type, one of binary or docker",
						Value: types.TypeBinary,
					},
				},
				Action: pierVersion,
			},
		},
	}
}

func allVersion(ctx *cli.Context) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}
	if !fileutil.Exist(repoRoot) {
		return fmt.Errorf("please `goduck init` first")
	}

	data, err := ioutil.ReadFile(filepath.Join(repoRoot, "release.json"))
	if err != nil {
		return err
	}

	fmt.Println(string(data))
	return nil
}

func goduckVersion(ctx *cli.Context) error {
	printVersion()

	return nil
}

func bxhVersion(ctx *cli.Context) error {
	upType := ctx.String("upType")
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}
	if !fileutil.Exist(repoRoot) {
		return fmt.Errorf("please `goduck init` first")
	}

	args := []string{types.VersionScript, "bxh", upType}
	if err := utils.ExecuteShell(args, repoRoot); err != nil {
		return err
	}

	return nil
}

func pierVersion(ctx *cli.Context) error {
	appchain := ctx.String("appchain")
	upType := ctx.String("upType")
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}
	if !fileutil.Exist(repoRoot) {
		return fmt.Errorf("please `goduck init` first")
	}

	args := []string{types.VersionScript, "pier", upType, appchain}
	if err := utils.ExecuteShell(args, repoRoot); err != nil {
		return err
	}

	return nil
}

func printVersion() {
	fmt.Printf("Goduck version: %s-%s-%s\n", goduck.CurrentVersion, goduck.CurrentBranch, goduck.CurrentCommit)
	fmt.Printf("App build date: %s\n", goduck.BuildDate)
	fmt.Printf("System version: %s\n", goduck.Platform)
	fmt.Printf("Golang version: %s\n", goduck.GoVersion)
}
