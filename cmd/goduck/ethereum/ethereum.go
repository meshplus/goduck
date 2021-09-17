package ethereum

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io/ioutil"
	"math/big"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/codeskyblue/go-sh"
	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/goduck/internal/download"
	"github.com/meshplus/goduck/internal/types"
)

type Ethereum struct {
	etherCli   *ethclient.Client
	privateKey *ecdsa.PrivateKey
	cid        *big.Int
}

type Config struct {
	EtherAddr    string
	KeyPath      string
	PasswordPath string
}

func New(config Config, repoRoot string) (*Ethereum, error) {
	configPath := filepath.Join(repoRoot, "ethereum")
	var keyPath string
	if len(config.KeyPath) == 0 {
		keyPath = filepath.Join(configPath, "account.key")
	} else {
		keyPath = config.KeyPath
	}

	etherCli, err := ethclient.Dial(config.EtherAddr)
	if err != nil {
		return nil, err
	}

	keyByte, err := ioutil.ReadFile(keyPath)
	if err != nil {
		return nil, err
	}

	var password string
	if len(config.PasswordPath) == 0 {
		psdPath := filepath.Join(configPath, "password")
		psd, err := ioutil.ReadFile(psdPath)
		if err != nil {
			return nil, err
		}
		password = strings.TrimSpace(string(psd))
	} else {
		psd, err := ioutil.ReadFile(config.PasswordPath)
		if err != nil {
			return nil, err
		}
		password = strings.TrimSpace(string(psd))
	}

	unlockedKey, err := keystore.DecryptKey(keyByte, password)
	if err != nil {
		return nil, err
	}

	Cid, err := etherCli.ChainID(context.Background())
	if err != nil {
		return nil, err
	}

	return &Ethereum{
		etherCli:   etherCli,
		privateKey: unlockedKey.PrivateKey,
		cid:        Cid,
	}, nil
}

func DownloadGethBinary(repoPath string) error {
	path := filepath.Join(repoPath, "bin", fmt.Sprintf("geth_%s_1.9.6", runtime.GOOS))
	if !fileutil.Exist(path) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}

	if !fileutil.Exist(filepath.Join(path, "geth")) {
		err := download.Download(path, fmt.Sprintf(types.GethUrl, runtime.GOOS))
		if err != nil {
			return err
		}

		err = sh.Command("/bin/bash", "-c", fmt.Sprintf("cd %s && tar xf %s --strip-components 1 ", path, fmt.Sprintf(types.GethTarName, runtime.GOOS))).Run()
		if err != nil {
			return err
		}
	}

	return nil
}
