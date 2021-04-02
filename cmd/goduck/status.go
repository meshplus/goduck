package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cheynewallace/tabby"
	"github.com/codeskyblue/go-sh"
	"github.com/docker/docker/client"
	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/internal/repo"
	gops "github.com/shirou/gopsutil/process"
	"github.com/urfave/cli/v2"
)

var processes = []string{
	"bitxhub/bitxhub.pid",
	"ethereum/ethereum.pid",
	"pier/pier-ethereum.pid",
	"pier/pier-fabric.pid",
}

var containers = []string{
	"bitxhub/bitxhub.cid",
	"ethereum/ethereum.cid",
	"pier/pier-ethereum.cid",
	"pier/pier-fabric.cid",
}

var modes = []string{
	"binary",
	"docker",
}

func GetStatusCMD() *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "check status of interchain system",
		Subcommands: []*cli.Command{
			{
				Name:   "list",
				Usage:  "Check the BitxHub, PIER and other components which in default configuration directories or start in docker",
				Action: showStatus,
			},
			{
				Name:  "component",
				Usage: "Check status of specified BitXHub or Pier",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:  "port",
						Value: "60011",
						Usage: "BitXHub's port(e.g. 60011), or peer's port(e.g. 44550)",
					},
				},
				Action: showComponentStatus,
			},
		},
	}
}

func showStatus(ctx *cli.Context) error {
	repoRoot, err := repo.PathRoot()
	if err != nil {
		return err
	}

	if !fileutil.Exist(repoRoot) {
		return fmt.Errorf("please `goduck init` first")
	}

	var table [][]string
	table = append(table, []string{"Name", "Component", "Mode", "PID/ContanierID", "Status", "Created Time", "Args"})

	for _, pro := range processes {
		table, err = existProcess(filepath.Join(repoRoot, pro), table)
		if err != nil {
			return err
		}
	}

	for _, con := range containers {
		table, err = existContainer(filepath.Join(repoRoot, con), table)
		if err != nil {
			return err
		}
	}
	PrintTable(table, true)
	return nil
}

func showComponentStatus(ctx *cli.Context) error {
	port := ctx.String("port")

	repoRoot, err := repo.PathRoot()
	if err != nil {
		return err
	}

	if !fileutil.Exist(repoRoot) {
		return fmt.Errorf("please `goduck init` first")
	}

	var table [][]string
	table = append(table, []string{"Name", "Component", "Mode", "PID/ContanierID", "Status", "Created Time", "Args"})

	paramsBin, err := binaryStatus(port)
	if err != nil {
		return fmt.Errorf("binary status error: %w", err)
	}
	if paramsBin != nil {
		if paramsBin[1] == "pier" || paramsBin[1] == "bitxhub" {
			table = append(table, paramsBin)
		}
	}

	paramsDocker, err := dockerStatus(port)
	if err != nil {
		return fmt.Errorf("docker status error: %w", err)
	}
	if paramsDocker != nil {
		table = append(table, paramsDocker)
	}

	PrintTable(table, true)
	return nil
}

func binaryStatus(port string) ([]string, error) {
	// lsof -i:60011 | grep LISTEN | awk '{print $2}'
	pidOut, err := sh.Command("/bin/bash", "-c", fmt.Sprintf("lsof -i:%s | grep LISTEN | awk '{print $2}'", port)).Output()
	if err != nil {
		return nil, fmt.Errorf("get pid error: %w", err)
	}

	if string(pidOut) != "" {
		params, err := getProcessParam(string(pidOut[:len(pidOut)-1]))
		if err != nil {
			return nil, fmt.Errorf("get process param error: %w", err)
		}
		return params, nil
	}

	return nil, nil
}

func dockerStatus(port string) ([]string, error) {
	tmpOut, _ := sh.Command("/bin/bash", "-c", "docker ps -a -q").Output()
	if string(tmpOut) == "" {
		return nil, nil
	}
	// docker inspect -f "{{.Id}} {{.HostConfig.PortBindings}} {{.Name}} " $(docker ps -q) | grep HostPort:6001
	cidOut, err := sh.Command("/bin/bash", "-c", fmt.Sprintf("docker inspect -f \"{{.Id}} {{.HostConfig.PortBindings}} {{.Name}} \" $(docker ps -a -q) | grep HostPort:%s | awk '{print $1}'", port)).Output()
	if err != nil {
		return nil, fmt.Errorf("get cid error: %w", err)
	}

	ctx := context.Background()
	mycli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, err
	}

	if string(cidOut) != "" {
		params, err := getContainerParam(string(cidOut[:len(cidOut)-1]), mycli, ctx)
		if err != nil {
			return nil, fmt.Errorf("get container param error: %w", err)
		}
		return params, nil
	}

	return nil, nil
}

func existProcess(pidPath string, table [][]string) ([][]string, error) {
	if !fileutil.Exist(pidPath) {
		return table, nil
	}
	fi, err := os.Open(pidPath)
	if err != nil {
		return table, err
	}
	defer fi.Close()

	br := bufio.NewReader(fi)
	i := 1
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		params, err := getProcessParam(string(a))
		if err != nil {
			return table, err
		}
		table = append(table, params)
		i++
	}
	return table, nil

}

func getProcessParam(pidStr string) ([]string, error) {
	pid, err := strconv.Atoi(string(pidStr))
	if err != nil {
		return nil, err
	}

	exist, err := gops.PidExists(int32(pid))

	status := "term"
	if err == nil && exist {
		status = "running"
	}

	process, err := gops.NewProcess(int32(pid))
	if err != nil {
		return nil, err
	}

	createTime, err := process.CreateTime()
	if err != nil {
		return nil, err
	}

	tm := time.Unix(0, createTime*int64(time.Millisecond))
	timeFormat := tm.Format(time.RFC3339)

	name, _ := process.Name()

	slice, _ := process.CmdlineSlice()
	args := strings.Join(slice, " ")
	if len(strings.Join(slice, " ")) > 1000 {
		args = args[:1000] + "..."
	}

	return []string{
		name,
		name,
		modes[0],
		pidStr,
		status,
		timeFormat,
		args,
	}, nil
}

func existContainer(cidPath string, table [][]string) ([][]string, error) {
	if !fileutil.Exist(cidPath) {
		return table, nil
	}
	fi, err := os.Open(cidPath)
	if err != nil {
		return table, err
	}
	defer fi.Close()

	ctx := context.Background()
	mycli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return table, err
	}

	br := bufio.NewReader(fi)
	for {
		a, _, c := br.ReadLine()
		if c == io.EOF {
			break
		}
		cid := string(a)

		params, err := getContainerParam(cid, mycli, ctx)
		if err != nil {
			return table, err
		}

		table = append(table, params)
	}
	return table, nil
}

func getContainerParam(cid string, mycli *client.Client, ctx context.Context) ([]string, error) {
	containerInfo, err := mycli.ContainerInspect(ctx, cid)
	if err != nil {
		return nil, err
	}

	info, _, err := mycli.ImageInspectWithRaw(ctx, containerInfo.Image[7:])
	if err != nil {
		return nil, err
	}

	pos1 := strings.LastIndex(info.RepoTags[0], "/")
	pos2 := strings.Index(info.RepoTags[0], "-")
	if pos2 == -1 {
		pos2 = len(info.RepoTags[0])
	}
	component := info.RepoTags[0][pos1+1 : pos2]

	args := strings.Join(containerInfo.Args, " ")
	if len(strings.Join(containerInfo.Args, " ")) > 1000 {
		args = args[:1000] + "..."
	}
	return []string{
		containerInfo.Name[1:],
		component,
		modes[1],
		cid,
		containerInfo.State.Status,
		containerInfo.Created,
		args,
	}, nil
}

// PrintTable accepts a matrix of strings and print them as ASCII table to terminal
func PrintTable(rows [][]string, header bool) {
	// Print the table
	t := tabby.New()
	if header {
		addRow(t, rows[0], header)
		rows = rows[1:]
	}
	for _, row := range rows {
		addRow(t, row, false)
	}
	t.Print()
}

func addRow(t *tabby.Tabby, rawLine []string, header bool) {
	// Convert []string to []interface{}
	row := make([]interface{}, len(rawLine))
	for i, v := range rawLine {
		row[i] = v
	}

	// Add line to the table
	if header {
		t.AddHeader(row...)
	} else {
		t.AddLine(row...)
	}
}
