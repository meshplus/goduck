package main

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/meshplus/bitxhub/pkg/cert"
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
			&cli.Uint64Flag{
				Name:  "num",
				Value: 4,
				Usage: "Node number, only useful in cluster mode, ignored in solo mode",
			},
			&cli.StringFlag{
				Name:  "type",
				Value: "binary",
				Usage: "configuration type, one of binary or docker",
			},
			&cli.StringFlag{
				Name:  "mode",
				Value: "cluster",
				Usage: "configuration mode, one of solo or cluster",
			},
			&cli.StringSliceFlag{
				Name:  "ips",
				Usage: "node IPs, use 127.0.0.1 for all nodes by default",
			},
			&cli.StringFlag{
				Name:  "target",
				Value: ".",
				Usage: "where to put the generated configuration files",
			},
		},
		Action: generateConfig,
	}
}

func generateConfig(ctx *cli.Context) error {
	num := ctx.Int("num")
	typ := ctx.String("type")
	mode := ctx.String("mode")
	ips := ctx.StringSlice("ips")
	target := ctx.String("target")

	return InitConfig(typ, mode, target, num, ips)
}

func InitConfig(typ, mode, target string, num int, ips []string) error {
	num, ips, err := processParams(num, typ, mode, ips)
	if err != nil {
		return err
	}

	fmt.Printf("initializing %d BitXHub nodes at %s\n", num, target)

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

	addrs, nodes, err := generateNodesConfig(target, mode, agencyPrivKey, agencyCertPath, ips)
	if err != nil {
		return fmt.Errorf("generate nodes config: %w", err)
	}

	if err := writeNetworkAndGenesis(target, addrs, nodes); err != nil {
		return fmt.Errorf("write network and genesis config: %w", err)
	}

	fmt.Printf("%d BitXHub nodes at %s are initialized successfully\n", num, target)

	return nil
}

func generateNodesConfig(repoRoot, mode, agencyPrivKey, agencyCertPath string, ips []string) ([]string, []*NetworkNodes, error) {
	count := len(ips)
	ipToId := make(map[string]int)
	addrs := make([]string, 0, count)
	nodes := make([]*NetworkNodes, 0, count)

	for i := 1; i <= count; i++ {
		ip := ips[i-1]
		ipToId[ip]++

		addr, node, err := generateNodeConfig(repoRoot, mode, agencyPrivKey, agencyCertPath, ip, i, ipToId)
		if err != nil {
			return nil, nil, err
		}

		addrs = append(addrs, addr)
		nodes = append(nodes, node)
	}

	return addrs, nodes, nil
}

func generateNodeConfig(repoRoot, mode, agencyPrivKey, agencyCertPath, ip string, id int, ipToId map[string]int) (string, *NetworkNodes, error) {
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

	if err := repo.Initialize(nodeRoot, mode, ipToId[ip]); err != nil {
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

func processParams(num int, typ string, mode string, ips []string) (int, []string, error) {
	if mode == "solo" {
		num = 1
	}

	if num == 0 {
		return 0, nil, fmt.Errorf("invalid node number")
	}

	if typ != "docker" && typ != "binary" {
		return 0, nil, fmt.Errorf("invalid type, choose one of docker or binary")
	}

	if mode != "solo" && mode != "cluster" {
		return 0, nil, fmt.Errorf("invalid mode, choose one of solo or cluster")
	}

	if typ == "docker" && mode == "cluster" && num != 4 {
		return 0, nil, fmt.Errorf("docker type supports 4 nodes only")
	}

	if len(ips) != 0 && len(ips) != num {
		return 0, nil, fmt.Errorf("IPs' number is not equal to nodes' number")
	}

	if err := checkIPs(ips); err != nil {
		return 0, nil, err
	}

	if len(ips) == 0 {
		if typ == "binary" {
			for i := 0; i < int(num); i++ {
				ips = append(ips, "127.0.0.1")
			}
		} else {
			for i := 2; i < int(num)+2; i++ {
				ip := fmt.Sprintf("172.19.0.%d", i)
				ips = append(ips, ip)
			}
		}
	}

	return num, ips, nil
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

func generateCA(dir string) (string, string, error) {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return "", "", err
	}

	priKeyEncode, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		return "", "", err
	}

	f, err := os.Create(repo.GetCAPrivKeyPath(dir))
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	err = pem.Encode(f, &pem.Block{Type: "EC PRIVATE KEY", Bytes: priKeyEncode})
	if err != nil {
		return "", "", err
	}

	c, err := cert.GenerateCert(privKey, true, "Hyperchain")
	if err != nil {
		return "", "", err
	}

	x509certEncode, err := x509.CreateCertificate(rand.Reader, c, c, privKey.Public(), privKey)
	if err != nil {
		return "", "", err
	}

	f, err = os.Create(repo.GetCACertPath(dir))
	if err != nil {
		return "", "", err
	}
	defer f.Close()

	if err := pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: x509certEncode}); err != nil {
		return "", "", err
	}

	return repo.GetCAPrivKeyPath(dir), repo.GetCACertPath(dir), nil
}

func generateCSR(org string, privPath string, target string) error {
	privData, err := ioutil.ReadFile(privPath)
	if err != nil {
		return err
	}
	block, _ := pem.Decode(privData)
	privKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("parse private key: %w", err)
	}

	template := &x509.CertificateRequest{
		Subject: pkix.Name{
			Country:            []string{"CN"},
			Locality:           []string{"HangZhou"},
			Province:           []string{"ZheJiang"},
			OrganizationalUnit: []string{"BitXHub"},
			Organization:       []string{org},
			StreetAddress:      []string{"street", "address"},
			PostalCode:         []string{"324000"},
			CommonName:         "BitXHub",
		},
	}
	data, err := x509.CreateCertificateRequest(rand.Reader, template, privKey)
	if err != nil {
		return err
	}

	name := getFileName(privPath)

	f, err := os.Create(repo.GetCSRPath(name, target))
	if err != nil {
		return err
	}
	defer f.Close()

	return pem.Encode(f, &pem.Block{Type: "CSR", Bytes: data})
}

func issueCert(csrPath, privPath, certPath, target string, isCA bool) error {
	privData, err := ioutil.ReadFile(privPath)
	if err != nil {
		return fmt.Errorf("read ca private key: %w", err)
	}
	block, _ := pem.Decode(privData)
	privKey, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("parse ca private key: %w", err)
	}

	caCertData, err := ioutil.ReadFile(certPath)
	if err != nil {
		return err
	}
	block, _ = pem.Decode(caCertData)
	caCert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return fmt.Errorf("parse ca cert: %w", err)
	}

	crsData, err := ioutil.ReadFile(csrPath)
	if err != nil {
		return fmt.Errorf("read crs: %w", err)
	}

	block, _ = pem.Decode(crsData)

	crs, err := x509.ParseCertificateRequest(block.Bytes)
	if err != nil {
		return fmt.Errorf("parse csr: %w", err)
	}

	if err := crs.CheckSignature(); err != nil {
		return fmt.Errorf("wrong csr sign: %w", err)
	}

	sn, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return err
	}

	notBefore := time.Now().Add(-5 * time.Minute).UTC()
	template := &x509.Certificate{
		Signature:             crs.Signature,
		SignatureAlgorithm:    crs.SignatureAlgorithm,
		PublicKey:             crs.PublicKey,
		PublicKeyAlgorithm:    crs.PublicKeyAlgorithm,
		SerialNumber:          sn,
		NotBefore:             notBefore,
		NotAfter:              notBefore.Add(50 * 365 * 24 * time.Hour).UTC(),
		BasicConstraintsValid: true,
		IsCA:                  isCA,
		Issuer:                caCert.Subject,
		KeyUsage: x509.KeyUsageDigitalSignature |
			x509.KeyUsageKeyEncipherment | x509.KeyUsageCertSign |
			x509.KeyUsageCRLSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageAny},
		Subject:     crs.Subject,
	}

	x509certEncode, err := x509.CreateCertificate(rand.Reader, template, caCert, crs.PublicKey, privKey)
	if err != nil {
		return fmt.Errorf("create cert: %w", err)
	}

	name := getFileName(csrPath)

	f, err := os.Create(repo.GetCertPath(name, target))
	if err != nil {
		return err
	}
	defer f.Close()

	return pem.Encode(f, &pem.Block{Type: "CERTIFICATE", Bytes: x509certEncode})
}

func generatePrivKey(name, target string) error {
	privKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}

	priKeyEncode, err := x509.MarshalECPrivateKey(privKey)
	if err != nil {
		return fmt.Errorf("marshal key: %w", err)
	}

	f, err := os.Create(repo.GetPrivKeyPath(name, target))
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	err = pem.Encode(f, &pem.Block{Type: "EC PRIVATE KEY", Bytes: priKeyEncode})
	if err != nil {
		return fmt.Errorf("pem encode: %w", err)
	}

	return nil
}

func getFileName(path string) string {
	def := "default"
	name := filepath.Base(path)
	bs := strings.Split(name, ".")
	if len(bs) != 2 {
		return def
	}

	return bs[0]
}
