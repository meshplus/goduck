package main

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/cavaliercoder/grab"
	"github.com/cheggaaa/pb"
	"github.com/meshplus/goduck/repo"
	"github.com/urfave/cli/v2"
)

const (
	wasmlibUrl = "https://raw.githubusercontent.com/meshplus/bitxhub/master/build/libwasmer.so"

	bitxhubUrlLinux = "https://github.com/meshplus/bitxhub/releases/download/v1.0.0-rc2/bitxhub_linux-amd64"
	bitxhubUrlMacOS = "https://github.com/meshplus/bitxhub/releases/download/v1.0.0-rc2/bitxhub_macos_x86_64"
)

func installCMD() *cli.Command {
	return &cli.Command{
		Name:  "install",
		Usage: "install BitXHub release binary app",
		Subcommands: []*cli.Command{
			{
				Name:  "bitxhub",
				Usage: "install bitxhub",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "path",
						Usage:    "Path for binary",
						Required: true,
					},
				},
				Action: installBitXHub,
			},
		},
	}
}

func installBitXHub(ctx *cli.Context) error {
	dir := ctx.String("path")
	if dir == "" {
		return fmt.Errorf("missing arg path")
	}

	_, err := os.Stat(dir)
	if err != nil {
		return err
	}

	root, err := repo.PathRoot()
	if err != nil {
		return err
	}

	err = download(root, wasmlibUrl)
	if err != nil {
		return err
	}

	if runtime.GOOS == "linux" {
		err := download(dir, bitxhubUrlLinux)
		if err != nil {
			return err
		}
	}
	if runtime.GOOS == "darwin" {
		err := download(dir, bitxhubUrlMacOS)
		if err != nil {
			return err
		}
	}

	fmt.Println("Download finished. Please follow instruction bellow:")
	fmt.Println("copy the file libwasmer.so under ~/.goduck directory to /usr/lib or $LD_LIBRARY_PATH")
	fmt.Println("You can set the $LD_LIBRARY_PATH and export it in user environment")

	return nil
}

func download(dst string, url string) error {
	client := grab.NewClient()
	req, err := grab.NewRequest(dst, url)
	if err != nil {
		return err
	}
	resp := client.Do(req)

	t := time.NewTicker(time.Millisecond)
	defer t.Stop()

	progress := &ProgressBar{}
	progress.Start(url, resp.Size)

L:
	for {
		select {
		case <-t.C:
			progress.SetCurrent(resp.BytesComplete())
		case <-resp.Done:
			progress.Finish()
			break L
		}
	}

	if err := resp.Err(); err != nil {
		if grab.IsStatusCodeError(err) {
			code := err.(grab.StatusCodeError)
			if int(code) == http.StatusNotFound {
				return fmt.Errorf("resource not found")
			}
		}
		return err
	}

	return nil
}

type ProgressBar struct {
	bar *pb.ProgressBar
}

// Start implement the DownloadProgress interface
func (p *ProgressBar) Start(url string, size int64) {
	p.bar = pb.Start64(size)
	p.bar.Set(pb.Bytes, true)
	p.bar.SetTemplateString(fmt.Sprintf(`download %s {{counters . }} {{percent . }} {{speed . }}`, url))
}

// SetCurrent implement the DownloadProgress interface
func (p *ProgressBar) SetCurrent(size int64) {
	p.bar.SetCurrent(size)
}

// Finish implement the DownloadProgress interface
func (p *ProgressBar) Finish() {
	p.bar.Finish()
}
