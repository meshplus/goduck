package main

import (
	"context"
	"fmt"
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

var processesNames = []string{
	"bitxhub/node",
	"pier_ethereum",
	"pier_fabric",
}

var processesAppchainNames = []string{
	".goduck/ethereum/datadir",
}

var containerNames = []string{
	"bitxhub_solo",
	"bitxhub_node",
	"pier-ethereum",
	"pier-fabric",
	"ethereum-node",
	"ethereum-1",
	"ethereum-2",
	"cli",
	"peer0.org1.example.com",
	"peer1.org1.example.com",
	"orderer.example.com",
	"peer1.org2.example.com",
	"peer0.org2.example.com",
}

var modes = []string{
	"binary",
	"docker",
}

func GetStatusCMD() *cli.Command {
	return &cli.Command{
		Name:  "status",
		Usage: "Check status of interchain system",
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
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	if !fileutil.Exist(repoRoot) {
		return fmt.Errorf("please `goduck init` first")
	}

	var table [][]string
	table = append(table, []string{"Name", "Component", "Mode", "PID/ContanierID", "Status", "Created Time", "Args"})

	table, err = existProcess(table)
	if err != nil {
		return err
	}

	table, err = existContainer(table)
	if err != nil {
		return err
	}

	PrintTable(table, true)
	return nil
}

func showComponentStatus(ctx *cli.Context) error {
	port := ctx.String("port")

	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
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
		params, err := getProcessParam(string(pidOut[:len(pidOut)-1]), "")
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

func existProcess(table [][]string) ([][]string, error) {
	for _, pname := range processesNames {
		pidOut, err := sh.Command("/bin/bash", "-c", fmt.Sprintf("ps aux | grep %s | grep start | grep -v grep | awk '{print $2}'", pname)).Output()
		if err != nil {
			return nil, fmt.Errorf("get pid error: %w", err)
		}
		pidOutStr := string(pidOut)

		if pidOutStr != "" {
			processesIDs := strings.Split(pidOutStr, "\n")
			for i, pid := range processesIDs {
				if pid != "" {
					params, err := getProcessParam(pid, fmt.Sprintf("%s_%d", pname, i+1))
					if err != nil {
						return table, err
					}

					table = append(table, params)
				}
			}
		}
	}

	for _, pname := range processesAppchainNames {
		pidOut, err := sh.Command("/bin/bash", "-c", fmt.Sprintf("ps aux | grep %s | grep -v grep | awk '{print $2}'", pname)).Output()
		if err != nil {
			return nil, fmt.Errorf("get pid error: %w", err)
		}
		pidOutStr := string(pidOut)

		if pidOutStr != "" {
			processesIDs := strings.Split(pidOutStr, "\n")
			for i, pid := range processesIDs {
				if pid != "" {
					params, err := getProcessParam(pid, fmt.Sprintf("%s_%d", pname, i+1))
					if err != nil {
						return table, err
					}

					table = append(table, params)
				}
			}
		}
	}

	return table, nil
}

func getProcessParam(pidStr, pName string) ([]string, error) {
	pid, err := strconv.Atoi(string(pidStr))
	if err != nil {
		return nil, err
	}

	var status, name, timeFormat, args string
	exist, err := gops.PidExists(int32(pid))
	if err != nil {
		// the process is killed by the outside
		return nil, fmt.Errorf("pid exist: %w", err)
	} else if exist {
		status = "running"
		process, err := gops.NewProcess(int32(pid))
		if err != nil {
			return nil, err
		}

		createTime, err := process.CreateTime()
		if err != nil {
			return nil, err
		}

		tm := time.Unix(0, createTime*int64(time.Millisecond))
		timeFormat = tm.Format(time.RFC3339)

		name, _ = process.Name()

		if pName == "" {
			pName = name
		}

		slice, _ := process.CmdlineSlice()
		args = strings.Join(slice, " ")
		if len(strings.Join(slice, " ")) > 500 {
			args = args[:70] + "..."
		}
	} else {
		status = "exited"
		if strings.Contains(pName, "bitxhub") {
			name = "bitxhub"
		} else if strings.Contains(pName, "pier") {
			name = "pier"
		}
	}

	return []string{
		pName,
		name,
		modes[0],
		pidStr,
		status,
		timeFormat,
		args,
	}, nil
}

func existContainer(table [][]string) ([][]string, error) {
	ctx := context.Background()
	mycli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return table, err
	}

	for _, cname := range containerNames {
		cidOut, err := sh.Command("/bin/bash", "-c", fmt.Sprintf("docker ps -qf \"name=%s\"", cname)).Output()
		if err != nil {
			return nil, fmt.Errorf("get cid error: %w", err)
		}
		cidOutStr := string(cidOut)
		if cidOutStr != "" {
			containerIDs := strings.Split(cidOutStr, "\n")
			for _, cid := range containerIDs {
				if cid != "" {
					params, err := getContainerParam(cid, mycli, ctx)
					if err != nil {
						return table, err
					}

					table = append(table, params)
				}
			}
		}
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
	pos2 := len(info.RepoTags[0])
	component := info.RepoTags[0][pos1+1 : pos2]

	args := strings.Join(containerInfo.Args, " ")
	if len(strings.Join(containerInfo.Args, " ")) > 70 {
		args = args[:70] + "..."
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
