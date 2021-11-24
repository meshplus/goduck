package main

import (
	"fmt"
	"os"
	"time"

	"github.com/urfave/cli/v2"
)

func main() {
	app := cli.NewApp()
	app.Name = "GoDuck"
	app.Usage = "GoDuck is a command-line management tool that can help to run BitXHub."
	app.Compiled = time.Now()

	cli.VersionPrinter = func(c *cli.Context) {
		printVersion()
	}

	// global flags
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "repo",
			Usage: "GoDuck storage repo path",
		},
	}

	app.Commands = []*cli.Command{
		GetInitCMD(),
		playgroundCMD(),
		bitxhubCMD(),
		pierCMD,
		deployCMD(),
		etherCMD(),
		fabricCMD(),
		hpcCMD,
		mqCMD,
		GetStatusCMD(),
		infoCMD(),
		keyCMD(),
		prometheusCMD(),
		logCMD(),
		getVersionCMD(),
	}

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err)
	}
}
