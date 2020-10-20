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

func GetStatusCMD() *cli.Command {
	return &cli.Command{
		Name:   "status",
		Usage:  "List the status of instantiated components",
		Action: showStatus,
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
	table = append(table, []string{"Name", "Component", "PID", "Status", "Created Time", "Args"})

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
		status := "TERM"
		pid, err := strconv.Atoi(string(a))
		if err != nil {
			return table, err
		}
		exist, err := gops.PidExists(int32(pid))
		if err == nil && exist {
			status = "RUNNING"
		}
		process, err := gops.NewProcess(int32(pid))
		if err != nil {
			continue
		}
		createTime, err := process.CreateTime()
		if err != nil {
			continue
		}
		tm := time.Unix(0, createTime*int64(time.Millisecond))
		timeFormat := tm.Format(time.RFC3339)
		name, _ := process.Name()
		nodeName := fmt.Sprintf(name+"-%d", i)

		slice, _ := process.CmdlineSlice()
		args := strings.Join(slice, " ")
		if len(strings.Join(slice, " ")) > 50 {
			args = args[:50] + "..."
		}
		table = append(table, []string{
			nodeName,
			name,
			strconv.Itoa(pid),
			status,
			timeFormat,
			args,
		})
		i++
	}
	return table, nil

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
	cli, err := client.NewClientWithOpts(client.FromEnv)
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

		containerInfo, err := cli.ContainerInspect(ctx, cid)
		if err != nil {
			return table, err
		}

		info, _, err := cli.ImageInspectWithRaw(ctx, containerInfo.Image[7:])
		if err != nil {
			return table, err
		}

		pos1 := strings.LastIndex(info.RepoTags[0], "/")
		pos2 := strings.Index(info.RepoTags[0], "-")
		if pos2 == -1 {
			pos2 = len(info.RepoTags[0])
		}
		component := info.RepoTags[0][pos1+1 : pos2]

		table = append(table, []string{
			containerInfo.Name[1:],
			component,
			cid,
			containerInfo.State.Status,
			containerInfo.Created,
			containerInfo.Args[0],
		})
	}
	return table, nil
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
