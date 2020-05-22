package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/meshplus/goduck/internal/repo"
	"github.com/pelletier/go-toml"
	"github.com/urfave/cli/v2"
)

type Genesis struct {
	Addresses []string `toml:"addresses" json:"addresses" `
}

type NetworkConfig struct {
	ID    uint64 `toml:"id" json:"id"`
	N     uint64
	Nodes []*NetworkNodes `toml:"nodes" json:"nodes"`
}

type NetworkNodes struct {
	ID   uint64 `toml:"id" json:"id"`
	Addr string `toml:"addr" json:"addr"`
}

func configCMD() *cli.Command {
	return &cli.Command{
		Name:  "config",
		Usage: "Generate configuration for BitXHub nodes",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "target",
				Value: ".",
				Usage: "where to put the generated configuration files",
			},
			&cli.Uint64Flag{
				Name:  "count",
				Value: 4,
				Usage: "Node count",
			},
			&cli.StringSliceFlag{
				Name:  "ips",
				Usage: "node IPs, use 127.0.0.1 for all nodes by default",
			},
		},
		Action: generateConfig,
	}
}

func generateConfig(ctx *cli.Context) error {
	target := ctx.String("target")
	count := ctx.Uint64("count")
	ips := ctx.StringSlice("ips")

	if err := checkParams(count, ips); err != nil {
		return err
	}

	fmt.Printf("initializing %d BitXHub nodes at %s\n", count, target)

	if len(ips) == 0 {
		for i := 0; i < int(count); i++ {
			ips = append(ips, "127.0.0.1")
		}
	}

	if repo.Initialized(target) {
		fmt.Println("BitXHub configuration file already exists")
		fmt.Println("reinitializing would overwrite your configuration, Y/N?")
		input := bufio.NewScanner(os.Stdin)
		input.Scan()

		if input.Text() != "Y" && input.Text() != "y" {
			return nil
		}

		if err := cleanOldConfig(target); err != nil {
			return fmt.Errorf("clean old configuration: %w", err)
		}
	}

	if _, err := os.Stat(target); os.IsNotExist(err) {
		if err := os.MkdirAll(target, 0755); err != nil {
			return err
		}
	}

	caPrivKey, caCertPath, err := generateCA(target)
	if err != nil {
		return fmt.Errorf("generate CA: %w", err)
	}

	agencyPrivKey, agencyCertPath, err := generateCert(repo.AgencyName, strings.ToUpper(repo.AgencyName), target,
		caPrivKey, caCertPath, true)
	if err != nil {
		return fmt.Errorf("generate agency cert: %w", err)
	}

	addrs, nodes, err := generateNodesConfig(target, agencyPrivKey, agencyCertPath, ips)
	if err != nil {
		return fmt.Errorf("generate nodes config: %w", err)
	}

	if err := writeNetworkAndGenesis(target, addrs, nodes); err != nil {
		return fmt.Errorf("write network and genesis config: %w", err)
	}

	fmt.Printf("%d BitXHub nodes at %s are initialized successfully\n", count, target)

	return nil
}

func generateNodesConfig(repoRoot, agencyPrivKey, agencyCertPath string, ips []string) ([]string, []*NetworkNodes, error) {
	count := len(ips)
	ipToId := make(map[string]int)
	addrs := make([]string, 0, count)
	nodes := make([]*NetworkNodes, 0, count)

	for i := 1; i <= count; i++ {
		ip := ips[i-1]
		ipToId[ip] = ipToId[ip] + 1

		addr, node, err := generateNodeConfig(repoRoot, agencyPrivKey, agencyCertPath, ip, i, ipToId)
		if err != nil {
			return nil, nil, err
		}

		addrs = append(addrs, addr)
		nodes = append(nodes, node)
	}

	return addrs, nodes, nil
}

func generateNodeConfig(repoRoot, agencyPrivKey, agencyCertPath, ip string, id int, ipToId map[string]int) (string, *NetworkNodes, error) {
	name := "node"
	org := "Node" + strconv.Itoa(id)
	nodeRoot := filepath.Join(repoRoot, name+strconv.Itoa(id))
	certRoot := filepath.Join(nodeRoot, "certs")

	if err := os.MkdirAll(certRoot, 0755); err != nil {
		return "", nil, err
	}

	if _, _, err := generateCert(name, org, certRoot, agencyPrivKey, agencyCertPath, false); err != nil {
		return "", nil, fmt.Errorf("generate node cert: %w", err)
	}

	if err := copyFile(repo.GetCACertPath(certRoot), repo.GetCACertPath(repoRoot)); err != nil {
		return "", nil, fmt.Errorf("copy ca cert: %w", err)
	}

	if err := copyFile(repo.GetCertPath("agency", certRoot), agencyCertPath); err != nil {
		return "", nil, fmt.Errorf("copy agency cert: %w", err)
	}

	if err := repo.Initialize(nodeRoot, ipToId[ip]); err != nil {
		return "", nil, fmt.Errorf("initialize configuration for node %d: %w", id, err)
	}

	addr, err := getAddressFromPrivateKey(repo.GetPrivKeyPath(name, certRoot))
	if err != nil {
		return "", nil, fmt.Errorf("get address from private key: %w", err)
	}

	pid, err := getPidFromPrivateKey(repo.GetPrivKeyPath(name, certRoot))
	if err != nil {
		return "", nil, fmt.Errorf("get pid from private key: %w", err)
	}

	node := &NetworkNodes{
		ID:   uint64(id),
		Addr: fmt.Sprintf("/ip4/%s/tcp/%d/p2p/%s", ip, 4000+ipToId[ip], pid),
	}

	return addr, node, nil
}

func writeNetworkAndGenesis(repoRoot string, addrs []string, nodes []*NetworkNodes) error {
	genesis := Genesis{Addresses: addrs}
	content, err := json.MarshalIndent(genesis, "", " ")
	if err != nil {
		return fmt.Errorf("marshal genesis: %w", err)
	}

	count := len(addrs)
	for i := 1; i <= count; i++ {
		nodeRoot := filepath.Join(repoRoot, "node"+strconv.Itoa(i))

		if err := ioutil.WriteFile(filepath.Join(nodeRoot, repo.GenesisConfigName), content, 0644); err != nil {
			return err
		}

		netConfig := NetworkConfig{
			ID:    uint64(i),
			N:     uint64(count),
			Nodes: nodes,
		}

		netContent, err := toml.Marshal(netConfig)
		if err != nil {
			return err
		}

		if err := ioutil.WriteFile(filepath.Join(nodeRoot, repo.NetworkConfigName), netContent, 0644); err != nil {
			return err
		}
	}

	return nil
}

func generateCert(name string, org string, target string, privKey string, caCertPath string, isCA bool) (string, string, error) {
	if err := generatePrivKey(name, target); err != nil {
		return "", "", fmt.Errorf("generate private key: %w", err)
	}

	if err := generateCSR(org, repo.GetPrivKeyPath(name, target), target); err != nil {
		return "", "", fmt.Errorf("generate csr: %w", err)
	}

	if err := issueCert(repo.GetCSRPath(name, target), privKey, caCertPath, target, isCA); err != nil {
		return "", "", fmt.Errorf("issue cert: %w", err)
	}

	if err := os.Remove(repo.GetCSRPath(name, target)); err != nil {
		return "", "", fmt.Errorf("remove csr: %w", err)
	}

	return repo.GetPrivKeyPath(name, target), repo.GetCertPath(name, target), nil
}

func copyFile(dstFile, srcFile string) error {
	src, err := os.Open(srcFile)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(dstFile, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}

func checkIPs(ips []string) error {
	for _, ip := range ips {
		parsedIP := net.ParseIP(ip)
		if parsedIP.To4() == nil {
			return fmt.Errorf("%v is not an IPv4 address", parsedIP)
		}
	}
	return nil
}

func checkParams(count uint64, ips []string) error {
	if count == 0 {
		return fmt.Errorf("invalid count")
	}

	if len(ips) != 0 && uint64(len(ips)) != count {
		return fmt.Errorf("IPs' count is not equal to nodes' count")
	}

	if err := checkIPs(ips); err != nil {
		return err
	}

	return nil
}

func cleanOldConfig(target string) error {
	if err := os.Remove(repo.GetCAPrivKeyPath(target)); err != nil {
		return fmt.Errorf("remove ca private key: %w", err)
	}
	if err := os.Remove(repo.GetCACertPath(target)); err != nil {
		return fmt.Errorf("remove ca certificate: %w", err)
	}

	if err := os.Remove(repo.GetPrivKeyPath(repo.AgencyName, target)); err != nil {
		return fmt.Errorf("remove agency private key: %w", err)
	}

	if err := os.Remove(repo.GetCertPath(repo.AgencyName, target)); err != nil {
		return fmt.Errorf("remove agency certificate: %w", err)
	}

	for i := 1; ; i++ {
		nodeDir := filepath.Join(target, "node"+strconv.Itoa(i))

		s, err := os.Stat(nodeDir)
		if err != nil {
			break
		}

		if s.IsDir() {
			if err := os.RemoveAll(nodeDir); err != nil {
				return fmt.Errorf("remove node configuration: %w", err)
			}
		}
	}

	return nil
}
