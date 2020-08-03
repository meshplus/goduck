package fabric

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

func Invoke(repoRoot, ccID, function, arg string) error {
	s := strings.Split(strings.TrimSpace(arg), ",")
	var args [][]byte
	for _, v := range s {
		args = append(args, []byte(strings.TrimSpace(v)))
	}

	fabCli, err := NewFabric(repoRoot)
	if err != nil {
		return err
	}

	request := channel.Request{
		ChaincodeID: ccID,
		Fcn:         function,
		Args:        args,
	}
	response, err := fabCli.Execute(request)
	if err != nil {
		fmt.Printf("invoke fail: %s\n", err)
	} else {
		fmt.Printf("[fabric] invoke function \"%s\", receipt is %s\n", function, string(response.Payload))
	}

	return nil
}

func NewFabric(repoRoot string) (*channel.Client, error) {
	filePath := filepath.Join(repoRoot, "config.yaml")

	// read config file，create SDK
	configProvider := config.FromFile(filePath)
	sdk, err := fabsdk.New(configProvider)
	if err != nil {
		return nil, fmt.Errorf("create sdk fail: %w\n", err)
	}

	// read config.yaml and find Admin in member1.example.com
	channelProvider := sdk.ChannelContext("mychannel", fabsdk.WithUser("Admin"), fabsdk.WithOrg("org2"))

	channelClient, err := channel.New(channelProvider)
	if err != nil {
		return nil, fmt.Errorf("create channel client fail: %w\n", err)
	}

	return channelClient, nil
}
