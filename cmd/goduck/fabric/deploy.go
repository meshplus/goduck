package fabric

import (
	"fmt"
	"os"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/spf13/viper"
)

func Deploy(configPath, ccid, path, version string) error {
	// read config file，create SDK
	configProvider := config.FromFile(configPath)
	sdk, err := fabsdk.New(configProvider)
	if err != nil {
		return err
	}

	pkg, err := gopackager.NewCCPackage(path, os.Getenv("GOPATH"))
	if err != nil {
		return err
	}

	req := resmgmt.InstallCCRequest{
		Name:    ccid,
		Path:    path,
		Version: version,
		Package: pkg,
	}

	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	clientContext := sdk.Context(fabsdk.WithUser("Admin"), fabsdk.WithOrg(viper.GetString("client.organization")))
	cli, err := resmgmt.New(clientContext)
	if err != nil {
		return err
	}

	ret, err := cli.InstallCC(req)
	if err != nil {
		return err
	}

	fmt.Println(ret)

	return nil
}
