package repo

import (
	"os"
	"path/filepath"

	"github.com/mitchellh/go-homedir"
)

const (
	// defaultPathName is the default config dir name
	defaultPathName = ".goduck"
	// defaultPathRoot is the path to the default config dir location.
	defaultPathRoot = "~/" + defaultPathName
	// envDir is the environment variable used to change the path root.
	envDir = "GODUCK_PATH"
	// Network config name
	NetworkConfigName = "network.toml"
	// BitXHub config name
	BitXHubConfigName = "bitxhub.toml"
	// Genesis config name
	GenesisConfigName = "genesis.json"
	// CA cert name
	caCertName = "ca.cert"
	// CA private key name
	caPrivKeyName = "ca.priv"
	// Agency name
	AgencyName = "agency"
	// key name
	KeyName = "key.json"
	// NodeKeyName
	NodeKeyName = "node.priv"
	// PierAdminKeyName
	AdminKeyName = "admin.json"
	// Pier config name
	PierConfigName = "pier.toml"
	// KeyPassword
	KeyPassword = "bitxhub"
	// private key name
	KeyPriv = "key"
	// proposal strategy: simple majority
	SimpleMajority = "SimpleMajority"
)

func PathRoot() (string, error) {
	dir := os.Getenv(envDir)
	var err error
	if len(dir) == 0 {
		dir, err = homedir.Expand(defaultPathRoot)
	}
	return dir, err
}

func PathRootWithDefault(path string) (string, error) {
	if len(path) == 0 {
		return PathRoot()
	}

	return path, nil
}

func GetCAPrivKeyPath(dir string) string {
	return filepath.Join(dir, caPrivKeyName)
}

func GetCACertPath(dir string) string {
	return filepath.Join(dir, caCertName)
}

func GetPrivKeyPath(name, dir string) string {
	return filepath.Join(dir, name+".priv")
}

func GetCSRPath(name, dir string) string {
	return filepath.Join(dir, name+".csr")
}

func GetCertPath(name, dir string) string {
	return filepath.Join(dir, name+".cert")
}
