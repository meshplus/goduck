package fabric

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/resmgmt"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	fab2 "github.com/hyperledger/fabric-sdk-go/pkg/fab"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/ccpackager/gopackager"
	"github.com/hyperledger/fabric-sdk-go/pkg/fab/peer"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
	"github.com/hyperledger/fabric/common/cauthdsl"
	"github.com/spf13/viper"
)

func Deploy(configPath, gopath, ccp, ccid, mspid, version string) error {
	// read config file，create SDK
	configProvider := config.FromFile(configPath)
	sdk, err := fabsdk.New(configProvider)
	if err != nil {
		return fmt.Errorf("fab sdk new: %v", err)
	}

	pkg, err := gopackager.NewCCPackage(ccp, gopath)
	if err != nil {
		return fmt.Errorf("new cc package1: %v", err)
	}

	req := resmgmt.InstallCCRequest{
		Name:    ccid,
		Path:    ccp,
		Version: version,
		Package: pkg,
	}

	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("read in config: %v", err)
	}

	clientContext := sdk.Context(fabsdk.WithUser("Admin"), fabsdk.WithOrg(viper.GetString("client.organization")))
	cli, err := resmgmt.New(clientContext)
	if err != nil {
		return fmt.Errorf("resmgmt new: %v", err)
	}

	var peers []fab.Peer
	provider, err := config.FromFile(configPath)()
	if err != nil {
		return fmt.Errorf("fail to get backend: %w", err)
	}
	endpointConfig, err := fab2.ConfigFromBackend(provider...)
	if err != nil {
		return fmt.Errorf("fail to get backend: %w", err)
	}

	nps := endpointConfig.NetworkPeers()
	for i := range nps {
		if nps[i].MSPID != mspid {
			continue
		}
		pr, err := peer.New(endpointConfig,
			peer.FromPeerConfig(&nps[i]))
		if err != nil {
			return fmt.Errorf("fail to new peer: %w", err)
		}
		peers = append(peers, pr)
	}

	installResp, err := cli.InstallCC(req, resmgmt.WithTargets(peers...))
	if err != nil {
		return fmt.Errorf("installcc error: %v", err)
	}

	fmt.Printf("InstallCC ret: %v\n", installResp)

	ccPolicy := cauthdsl.SignedByMspMember(mspid)
	instantiateReq := resmgmt.InstantiateCCRequest{
		Name:    ccid,
		Path:    ccp,
		Version: version,
		Policy:  ccPolicy,
	}

	instantiateResp, err := cli.InstantiateCC("mychannel", instantiateReq, resmgmt.WithTargets(peers...))
	if err != nil {
		return fmt.Errorf("InstantiateCC error: %v", err)
	}

	fmt.Printf("InstantiateCC ret: %v\n", instantiateResp)

	return nil
}
