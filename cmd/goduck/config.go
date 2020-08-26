package main

import (
	"bufio"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/hex"
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
	"text/template"
	"time"

	"github.com/gobuffalo/packd"
	"github.com/gobuffalo/packr"
	"github.com/libp2p/go-libp2p-core/peer"
	crypto2 "github.com/meshplus/bitxhub-kit/crypto"
	ecdsa2 "github.com/meshplus/bitxhub-kit/crypto/asym/ecdsa"
	"github.com/meshplus/bitxhub-kit/fileutil"
	key2 "github.com/meshplus/bitxhub-kit/key"
	"github.com/meshplus/bitxhub/pkg/cert"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/meshplus/goduck/internal/types"
	"github.com/pelletier/go-toml"
	"github.com/urfave/cli/v2"
)

var (
	PackPath                 = "../../config"
	_        ConfigGenerator = (*BitXHubConfigGenerator)(nil)
	_        ConfigGenerator = (*PierConfigGenerator)(nil)
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

type ReadinNetworkConfig struct {
	Addrs [][]string
}

type ConfigGenerator interface {
	// Initialized
	Initialized() (bool, error)

	// CleanOldConfig
	CleanOldConfig() error

	// InitConfig
	InitConfig() error

	// ProcessParams
	ProcessParams() error
}

type BitXHubConfigGenerator struct {
	typ     string
	mode    string
	target  string
	num     int
	ips     []string
	version string
}

type PierConfigGenerator struct {
	mode         string
	appchainType string
	appchainIP   string
	bitxhub      string
	target       string
	validators   []string
	peers        []string
	port         int
	id           int
}

func NewBitXHubConfigGenerator(typ string, mode string, target string, num int, ips []string, version string) *BitXHubConfigGenerator {
	return &BitXHubConfigGenerator{typ: typ, mode: mode, target: target, num: num, ips: ips, version: version}
}

func NewPierConfigGenerator(mode, appchainType, appchainIP, bitxhub, target string, validators, peers []string, port, id int) *PierConfigGenerator {
	return &PierConfigGenerator{
		mode:         mode,
		appchainType: appchainType,
		appchainIP:   appchainIP,
		bitxhub:      bitxhub,
		target:       target,
		validators:   validators,
		peers:        peers,
		port:         port,
		id:           id,
	}
}

func (b *BitXHubConfigGenerator) Initialized() (bool, error) {
	if fileutil.Exist(repo.GetCAPrivKeyPath(b.target)) ||
		fileutil.Exist(repo.GetCACertPath(b.target)) ||
		fileutil.Exist(repo.GetPrivKeyPath(b.target, repo.AgencyName)) ||
		fileutil.Exist(repo.GetCertPath(b.target, repo.AgencyName)) {
		return true, nil
	}

	if ok, err := existDir(filepath.Join(b.target, "node1")); err != nil {
		return ok, err
	} else if ok {
		return true, nil
	}

	if ok, err := existDir(filepath.Join(b.target, "nodeSolo")); err != nil {
		return ok, err
	} else if ok {
		return true, nil
	}

	return false, nil
}

func (b *BitXHubConfigGenerator) CleanOldConfig() error {
	if err := os.Remove(repo.GetCAPrivKeyPath(b.target)); err != nil {
		return fmt.Errorf("remove ca private key: %w", err)
	}
	if err := os.Remove(repo.GetCACertPath(b.target)); err != nil {
		return fmt.Errorf("remove ca certificate: %w", err)
	}

	if err := os.Remove(repo.GetPrivKeyPath(repo.AgencyName, b.target)); err != nil {
		return fmt.Errorf("remove agency private key: %w", err)
	}

	if err := os.Remove(repo.GetCertPath(repo.AgencyName, b.target)); err != nil {
		return fmt.Errorf("remove agency certificate: %w", err)
	}

	for i := 1; ; i++ {
		nodeDir := filepath.Join(b.target, "node"+strconv.Itoa(i))
		exist, err := removeDir(nodeDir)
		if err != nil {
			return err
		}

		if !exist {
			break
		}
	}

	if _, err := removeDir(filepath.Join(b.target, "nodeSolo")); err != nil {
		return err
	}

	return nil
}

func (b *BitXHubConfigGenerator) InitConfig() error {
	if err := b.ProcessParams(); err != nil {
		return err
	}

	fmt.Printf("initializing %d BitXHub nodes at %s\n", b.num, b.target)

	if ok, err := b.Initialized(); err != nil {
		return fmt.Errorf("check if BitXHub is initialized: %w", err)
	} else if ok {
		fmt.Println("BitXHub configuration file already exists")
		fmt.Println("reinitializing would overwrite your configuration, Y/N (default: N)?")
		input := bufio.NewScanner(os.Stdin)
		input.Scan()

		if input.Text() != "Y" && input.Text() != "y" {
			return nil
		}

		if err := b.CleanOldConfig(); err != nil {
			return fmt.Errorf("clean old configuration: %w", err)
		}
	}

	if _, err := os.Stat(b.target); os.IsNotExist(err) {
		if err := os.MkdirAll(b.target, 0755); err != nil {
			return err
		}
	}

	caPrivKey, caCertPath, err := generateCA(b.target)
	if err != nil {
		return fmt.Errorf("generate CA: %w", err)
	}

	agencyPrivKey, agencyCertPath, err := generateCert(repo.AgencyName, strings.ToUpper(repo.AgencyName), b.target,
		caPrivKey, caCertPath, true)
	if err != nil {
		return fmt.Errorf("generate agency cert: %w", err)
	}

	addrs, nodes, err := b.generateNodesConfig(b.target, b.mode, agencyPrivKey, agencyCertPath, b.ips)

	if err != nil {
		return fmt.Errorf("generate nodes config: %w", err)
	}

	if err := writeNetworkAndGenesis(b.target, b.mode, addrs, nodes, b.version); err != nil {
		return fmt.Errorf("write network and genesis config: %w", err)
	}

	fmt.Printf("%d BitXHub nodes at %s are initialized successfully\n", b.num, b.target)

	return nil
}

func (b *BitXHubConfigGenerator) ProcessParams() error {
	if b.mode == types.SoloMode {
		b.num = 1
	}

	if b.num == 0 {
		return fmt.Errorf("invalid node number")
	}

	if b.typ != types.TypeDocker && b.typ != types.TypeBinary {
		return fmt.Errorf("invalid type, choose one of docker or binary")
	}

	if b.mode != types.SoloMode && b.mode != types.ClusterMode {
		return fmt.Errorf("invalid mode, choose one of solo or cluster")
	}

	if b.typ == types.TypeDocker && b.mode == types.ClusterMode && b.num != 4 {
		return fmt.Errorf("docker type supports 4 nodes only")
	}

	if len(b.ips) != 0 && len(b.ips) != b.num {
		return fmt.Errorf("IPs' number is not equal to nodes' number")
	}

	if len(b.ips) == 0 && b.num >= 10 {
		return fmt.Errorf("can not create more than 10 nodes with one IP address")
	}

	if err := checkIPs(b.ips); err != nil {
		return err
	}

	if len(b.ips) == 0 {
		if b.typ == types.TypeBinary || b.mode == types.SoloMode {
			for i := 0; i < b.num; i++ {
				b.ips = append(b.ips, "127.0.0.1")
			}
		} else {
			for i := 2; i < b.num+2; i++ {
				ip := fmt.Sprintf("172.19.0.%d", i)
				b.ips = append(b.ips, ip)
			}
		}
	}

	return nil
}

func (p *PierConfigGenerator) Initialized() (bool, error) {
	if fileutil.Exist(filepath.Join(p.target, repo.PierConfigName)) {
		return true, nil
	}

	return existDir(filepath.Join(p.target, p.appchainType))
}

func (p *PierConfigGenerator) copyConfigFiles() error {
	privKey, err := generatePierKey(p.target)
	if err != nil {
		return fmt.Errorf("generate Pier's private key: %w", err)
	}

	validators := ""
	bitxhub := ""
	peers := ""
	if p.mode == types.PierModeRelay {
		bitxhub = p.bitxhub
		for _, v := range p.validators {
			validators += "\"" + v + "\",\n"
		}
	} else {
		for _, v := range p.peers {
			peers += "\"" + v + "\",\n"
		}

		libp2pPrivKey, err := convertToLibp2pPrivKey(privKey)
		if err != nil {
			return fmt.Errorf("convert to libp2p private key: %w", err)
		}

		pid, err := peer.IDFromPrivateKey(libp2pPrivKey)
		if err != nil {
			return fmt.Errorf("get ID from libp2p private key: %w", err)
		}

		peers += "\"" + fmt.Sprintf("/ip4/0.0.0.0/tcp/%d/p2p/%s", p.port, pid) + "\",\n"
	}

	pluginFile := "fabric-client-1.4.so"
	pluginConfig := p.appchainType
	if p.appchainType == types.ChainTypeEther {
		pluginFile = "eth-client.so"
	}

	data := struct {
		Mode         string
		Bitxhub      string
		Validators   string
		Peers        string
		PluginFile   string
		PluginConfig string
		Id           int
	}{p.mode, bitxhub, validators, peers, pluginFile, pluginConfig, p.id}

	files := []string{
		filepath.Join("pier", "api"),
		filepath.Join("pier", "pier.toml"),
	}
	if err := renderConfigFile(p.target, files, data); err != nil {
		return fmt.Errorf("initialize Pier configuration file: %w", err)
	}

	dstDir := filepath.Join(p.target, p.appchainType)
	srcDir := filepath.Join("pier", p.appchainType)
	if err := renderConfigFiles(dstDir, srcDir, []string{p.appchainType + ".toml"}, p.appchainIP); err != nil {
		return fmt.Errorf("initialize Pier plugin configuration files: %w", err)
	}

	return nil
}

func (p *PierConfigGenerator) CleanOldConfig() error {
	if err := os.Remove(filepath.Join(p.target, repo.PierConfigName)); err != nil {
		return fmt.Errorf("remove Pier's old configuration file: %w", err)
	}

	if _, err := removeDir(filepath.Join(p.target, p.appchainType)); err != nil {
		return fmt.Errorf("remove Pier's old plugin configuration: %w", err)
	}

	return nil
}

func (p *PierConfigGenerator) InitConfig() error {
	if err := p.ProcessParams(); err != nil {
		return err
	}

	fmt.Printf("initializing Pier configuration at %s\n", p.target)

	if ok, err := p.Initialized(); err != nil {
		return fmt.Errorf("check if Pier is initialized: %w", err)
	} else if ok {
		fmt.Println("Pier configuration file already exists")
		fmt.Println("reinitializing would overwrite your configuration, Y/N?")
		input := bufio.NewScanner(os.Stdin)
		input.Scan()

		if input.Text() != "Y" && input.Text() != "y" {
			return nil
		}

		if err := p.CleanOldConfig(); err != nil {
			return fmt.Errorf("clean old configuration: %w", err)
		}
	}

	if _, err := os.Stat(p.target); os.IsNotExist(err) {
		if err := os.MkdirAll(p.target, 0755); err != nil {
			return err
		}
	}

	if err := p.copyConfigFiles(); err != nil {
		return err
	}

	fmt.Printf("Pier configuration at %s are initialized successfully\n", p.target)

	return nil
}

func (p *PierConfigGenerator) ProcessParams() error {
	if p.mode != types.PierModeDirect && p.mode != types.PierModeRelay {
		return fmt.Errorf("invalid mode, choose one of direct or relay")
	}

	if p.mode == types.PierModeRelay && p.bitxhub == "" {
		return fmt.Errorf("BitXhub's address is needed in relay mode")
	}

	if p.mode == types.PierModeRelay && len(p.validators) == 0 {
		return fmt.Errorf("BitXhub validators' information is needed in relay mode")
	}

	if p.mode == types.PierModeDirect && len(p.peers) == 0 {
		fmt.Println("You have to add peers' information manually after the configuration files are generated")
	}

	if p.appchainType != types.ChainTypeEther && p.appchainType != types.ChainTypeFabric {
		return fmt.Errorf("invalid appchain type, choose one of ethereum or fabric")
	}

	if p.id < 0 || p.id > 9 {
		return fmt.Errorf("invalid ID, should be in [0, 9]")
	}

	if err := checkIPs([]string{p.appchainIP}); err != nil {
		return err
	}

	return nil
}

func generateBitXHubConfig(ctx *cli.Context) error {
	num := ctx.Int("num")
	typ := ctx.String("type")
	mode := ctx.String("mode")
	ips := ctx.StringSlice("ips")
	target := ctx.String("target")

	return InitBitXHubConfig(typ, mode, target, num, ips, "")
}

func generatePierConfig(ctx *cli.Context) error {
	mode := ctx.String("mode")
	bitxhub := ctx.String("bitxhub")
	validators := ctx.StringSlice("validators")
	peers := ctx.StringSlice("peers")
	appchainType := ctx.String("appchain-type")
	appchainIP := ctx.String("appchain-IP")
	target := ctx.String("target")
	port := ctx.Int("port")
	id := ctx.Int("ID")

	return InitPierConfig(mode, bitxhub, appchainType, appchainIP, target, validators, peers, port, id)
}

func InitBitXHubConfig(typ, mode, target string, num int, ips []string, version string) error {
	bcg := NewBitXHubConfigGenerator(typ, mode, target, num, ips, version)
	return bcg.InitConfig()
}

func InitPierConfig(mode, bitxhub, appchainType, appchainIP, target string, validators, peers []string, port, id int) error {
	pcg := NewPierConfigGenerator(mode, appchainType, appchainIP, bitxhub, target, validators, peers, port, id)
	return pcg.InitConfig()
}

func (b *BitXHubConfigGenerator) generateNodesConfig(repoRoot, mode, agencyPrivKey, agencyCertPath string, ips []string) ([]string, []*NetworkNodes, error) {
	count := len(ips)
	ipToId := make(map[string]int)
	addrs := make([]string, 0, count)
	nodes := make([]*NetworkNodes, 0, count)

	for i := 1; i <= count; i++ {
		ip := ips[i-1]
		ipToId[ip]++

		addr, node, err := b.generateNodeConfig(repoRoot, mode, agencyPrivKey, agencyCertPath, ip, i, ipToId)
		if err != nil {
			return nil, nil, err
		}

		addrs = append(addrs, addr)
		nodes = append(nodes, node)
	}

	return addrs, nodes, nil
}

func (b *BitXHubConfigGenerator) generateNodeConfig(repoRoot, mode, agencyPrivKey, agencyCertPath, ip string, id int, ipToId map[string]int) (string, *NetworkNodes, error) {
	name := "node"
	org := "Node" + strconv.Itoa(id)
	nodeRoot := filepath.Join(repoRoot, name+strconv.Itoa(id))
	if mode == types.SoloMode {
		org = "NodeSolo"
		nodeRoot = filepath.Join(repoRoot, name+"Solo")
	}
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

	if err := b.copyConfigFiles(nodeRoot, ipToId[ip]); err != nil {
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

func (b *BitXHubConfigGenerator) copyConfigFiles(nodeRoot string, id int) error {
	consensus := types.SoloMode
	if b.mode == "cluster" {
		consensus = "raft"
	}

	data := struct {
		Id        int
		Solo      bool
		Consensus string
	}{id, b.mode == "solo", consensus}

	files := []string{"bitxhub.toml", "api"}

	return renderConfigFiles(nodeRoot, "bitxhub", files, data)
}

func writeNetworkAndGenesis(repoRoot, mode string, addrs []string, nodes []*NetworkNodes, version string) error {
	genesis := Genesis{Addresses: addrs}
	content, err := json.MarshalIndent(genesis, "", " ")
	if err != nil {
		return fmt.Errorf("marshal genesis: %w", err)
	}

	count := len(addrs)

	for i := 1; i <= count; i++ {
		nodeRoot := filepath.Join(repoRoot, "node"+strconv.Itoa(i))
		if mode == "solo" {
			nodeRoot = filepath.Join(repoRoot, "nodeSolo")
		}

		if err := ioutil.WriteFile(filepath.Join(nodeRoot, repo.GenesisConfigName), content, 0644); err != nil {
			return err
		}

		if version >= "v1.1.0-rc1" {
			var addrs [][]string
			for _, node := range nodes {
				addrs = append(addrs, []string{node.Addr})
			}
			netConfig := ReadinNetworkConfig{Addrs: addrs}
			netContent, err := toml.Marshal(netConfig)
			if err != nil {
				return err
			}

			if err := ioutil.WriteFile(filepath.Join(nodeRoot, repo.NetworkConfigName), netContent, 0644); err != nil {
				return err
			}
			continue
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

func removeDir(dir string) (bool, error) {
	ok, err := existDir(dir)
	if err != nil {
		return false, err
	}

	if !ok {
		return false, nil
	}

	if err := os.RemoveAll(dir); err != nil {
		return true, fmt.Errorf("remove node configuration: %w", err)
	}

	return true, nil
}

func existDir(path string) (bool, error) {
	if !fileutil.Exist(path) {
		return false, nil
	}

	s, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return s.IsDir(), nil
}

func generatePierKey(target string) (crypto2.PrivateKey, error) {
	privKey, err := ecdsa2.GenerateKey(ecdsa2.Secp256r1)
	if err != nil {
		return nil, fmt.Errorf("generate key: %w", err)
	}

	keyBytes, err := privKey.Bytes()
	if err != nil {
		return nil, fmt.Errorf("get private key bytes: %w", err)
	}

	cipher := hex.EncodeToString(keyBytes)
	address, err := privKey.PublicKey().Address()
	if err != nil {
		return nil, fmt.Errorf("get public key address: %w", err)
	}

	key := &key2.Key{
		Address:    address,
		PrivateKey: cipher,
		Encrypted:  false,
	}

	ret, err := json.MarshalIndent(key, "", "	")
	if err != nil {
		return nil, fmt.Errorf("marshal key: %w", err)
	}

	keyPath := filepath.Join(target, repo.KeyName)
	if err = ioutil.WriteFile(keyPath, ret, os.ModePerm); err != nil {
		return nil, fmt.Errorf("persist key: %s", err)
	}

	return privKey, nil
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

func renderConfigFiles(dstDir, srcDir string, filesToRender []string, data interface{}) error {
	filesM := make(map[string]struct{})
	for _, file := range filesToRender {
		filesM[file] = struct{}{}
	}

	box := packr.NewBox(filepath.Join(PackPath, srcDir))
	if err := box.Walk(func(s string, file packd.File) error {
		p := filepath.Join(dstDir, s)
		dir := filepath.Dir(p)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				return err
			}
		}

		fileName := file.Name()
		if _, ok := filesM[fileName]; ok {
			t := template.New(fileName)
			t, err := t.Parse(file.String())
			if err != nil {
				return err
			}

			f, err := os.Create(p)
			if err != nil {
				return err
			}

			return t.Execute(f, data)
		} else {
			return ioutil.WriteFile(p, []byte(file.String()), 0644)
		}
	}); err != nil {
		return err
	}

	return nil
}

func renderConfigFile(dstDir string, srcFiles []string, data interface{}) error {
	box := packr.NewBox(PackPath)

	for _, srcFile := range srcFiles {
		fileStr, err := box.FindString(srcFile)
		if err != nil {
			return fmt.Errorf("find file in box: %w", err)
		}

		t := template.New(filepath.Base(srcFile))
		t, err = t.Parse(fileStr)
		if err != nil {
			return err
		}

		f, err := os.Create(filepath.Join(dstDir, filepath.Base(srcFile)))
		if err != nil {
			return err
		}

		if err := t.Execute(f, data); err != nil {
			return err
		}
	}

	return nil
}
