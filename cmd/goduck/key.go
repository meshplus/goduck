package main

import (
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/peer"
	crypto2 "github.com/meshplus/bitxhub-kit/crypto"
	"github.com/meshplus/bitxhub-kit/crypto/asym"
	"github.com/meshplus/bitxhub-kit/crypto/asym/ecdsa"
	"github.com/meshplus/bitxhub-kit/fileutil"
	"github.com/meshplus/bitxhub/pkg/cert"
	libp2pcert "github.com/meshplus/go-libp2p-cert"
	"github.com/meshplus/goduck/internal/repo"
	"github.com/urfave/cli/v2"
)

func keyCMD() *cli.Command {
	return &cli.Command{
		Name:  "key",
		Usage: "Create and show key information",
		Subcommands: []*cli.Command{
			Secp256k1(),
			ECDSA_P256(),
		},
	}
}

func Secp256k1() *cli.Command {
	return &cli.Command{
		Name:  "Secp256k1",
		Usage: "Create and show Secp256k1 key information",
		Subcommands: []*cli.Command{
			{
				Name:  "gen",
				Usage: "Create new Secp256k1 private key",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Specific private key name",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "target",
						Usage:    "Specific target directory (default: $repo/key/$name)",
						Required: false,
					},
				},
				Action: func(ctx *cli.Context) error {
					return generateKey(ctx, crypto2.Secp256k1)
				},
			},
			{
				Name:  "convert",
				Usage: "Convert new key file from private key",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Aliases: []string{"save", "s"},
						Usage:   "Save key into repo",
					},
					&cli.StringFlag{
						Name:     "priv",
						Usage:    "Private key path",
						Required: true,
					},
				},
				Action: convertKey,
			},
			{
				Name:   "address",
				Usage:  "Show address from private key",
				Action: getAddress,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "path",
						Usage:    "Specific private key path",
						Required: true,
					},
				},
			},
		},
	}
}

func ECDSA_P256() *cli.Command {
	return &cli.Command{
		Name:  "ECDSA_P256",
		Usage: "Create and show ECDSA_P256 key information",
		Subcommands: []*cli.Command{
			{
				Name:  "gen",
				Usage: "Create new ECDSA_P256 private key",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "name",
						Usage:    "Specific private key name",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "target",
						Usage:    "Specific target directory (default: $repo/key/$name)",
						Required: false,
					},
				},
				Action: func(ctx *cli.Context) error {
					return generateKey(ctx, crypto2.ECDSA_P256)
				},
			},
			{
				Name:   "pid",
				Usage:  "Show pid from private key",
				Action: getPid,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "path",
						Usage:    "Private Key Path",
						Required: true,
					},
				},
			},
		},
	}
}

func generateKey(ctx *cli.Context, opt crypto2.KeyType) error {
	repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
	if err != nil {
		return err
	}

	name := ctx.String("name")
	target := ctx.String("target")
	if target == "" {
		target = filepath.Join(repoRoot, "key")
	}
	keyPath := filepath.Join(target, fmt.Sprintf("%s.priv", name))

	privKey, err := asym.GenerateKeyPair(opt)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}

	priKeyEncode, err := privKey.Bytes()
	if err != nil {
		return fmt.Errorf("marshal key: %w", err)
	}

	if !fileutil.Exist(target) {
		err := os.MkdirAll(target, 0755)
		if err != nil {
			return fmt.Errorf("create folder: %w", err)
		}
	}
	f, err := os.Create(keyPath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	err = pem.Encode(f, &pem.Block{Type: "EC PRIVATE KEY", Bytes: priKeyEncode})
	if err != nil {
		return fmt.Errorf("pem encode: %w", err)
	}

	if fileutil.Exist(keyPath) {
		color.Green("Generate key in %s successful", keyPath)
	}
	return nil
}

func convertKey(ctx *cli.Context) error {
	privPath := ctx.String("priv")

	data, err := ioutil.ReadFile(privPath)
	if err != nil {
		return fmt.Errorf("read private key: %w", err)
	}
	privKey, err := cert.ParsePrivateKey(data, crypto2.Secp256k1)
	if err != nil {
		return fmt.Errorf("parse private key: %w", err)
	}

	if ctx.Bool("save") {
		repoRoot, err := repo.PathRootWithDefault(ctx.String("repo"))
		if err != nil {
			return fmt.Errorf("get reporoot: %w", err)
		}

		keyPath := filepath.Join(repoRoot, "key")
		if !fileutil.Exist(keyPath) {
			err := os.MkdirAll(keyPath, 0755)
			if err != nil {
				return fmt.Errorf("create folder: %w", err)
			}
		}

		if err := asym.StorePrivateKey(privKey, filepath.Join(keyPath, repo.KeyName), "bitxhub"); err != nil {
			return fmt.Errorf("store private key: %w", err)
		} else {
			color.Green("Store converted key in %s successful", filepath.Join(keyPath, repo.KeyName))
		}
	} else {
		keyStore, err := asym.GenKeyStore(privKey, "bitxhub")
		if err != nil {
			return err
		}

		pretty, err := keyStore.Pretty()
		if err != nil {
			return err
		}

		fmt.Println(pretty)
	}

	return nil
}

func getPid(ctx *cli.Context) error {
	privPath := ctx.String("path")

	pid, err := getPidFromPrivateKey(privPath)
	if err != nil {
		return err
	}

	fmt.Println(pid)

	return nil
}

func getPidFromPrivateKey(privPath string) (string, error) {
	data, err := ioutil.ReadFile(privPath)
	if err != nil {
		return "", fmt.Errorf("read private key: %w", err)
	}
	privKey, err := cert.ParsePrivateKey(data, crypto2.ECDSA_P256)
	if err != nil {
		return "", err
	}

	_, pk, err := crypto.KeyPairFromStdKey(privKey.K)
	if err != nil {
		return "", err
	}

	pid, err := peer.IDFromPublicKey(pk)
	if err != nil {
		return "", err
	}

	return pid.String(), nil
}

func getAddress(ctx *cli.Context) error {
	privPath := ctx.String("path")

	addr, err := getAddressFromPrivateKey(privPath, crypto.Secp256k1)
	if err != nil {
		return fmt.Errorf("get address from private key: %s", err)
	}

	fmt.Println(addr)

	return nil
}

func getAddressFromPrivateKey(privPath string, opt crypto2.KeyType) (string, error) {
	data, err := ioutil.ReadFile(privPath)
	if err != nil {
		return "", fmt.Errorf("read private key: %w", err)
	}

	privKey, err := libp2pcert.ParsePrivateKey(data, opt) //crypto2.Secp256k1
	if err != nil {
		return "", err
	}

	addr, err := privKey.PublicKey().Address()
	if err != nil {
		return "", fmt.Errorf("priv to address: %w", err)
	}

	return addr.String(), nil
}

func convertToLibp2pPrivKey(privateKey crypto2.PrivateKey) (crypto.PrivKey, error) {
	ecdsaPrivKey, ok := privateKey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, fmt.Errorf("convert to libp2p private key: not ecdsa private key")
	}

	libp2pPrivKey, _, err := crypto.ECDSAKeyPairFromKey(ecdsaPrivKey.K)
	if err != nil {
		return nil, err
	}

	return libp2pPrivKey, nil
}
