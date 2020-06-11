package types

const (
	TypeBinary     = "binary"
	TypeDocker     = "docker"
	ClusterMode    = "cluster"
	SoloMode       = "solo"
	PierModeDirect = "direct"
	PierModeRelay  = "relay"

	ChainTypeEther  = "ether"
	ChainTypeFabric = "fabric"

	PlaygroundScript = "playground.sh"
	FabricScript     = "fabric.sh"
	ChaincodeScript  = "chaincode.sh"
	EthereumScript   = "ethereum.sh"

	LinuxWasmLibUrl = "https://raw.githubusercontent.com/meshplus/bitxhub/master/build/libwasmer.so"
	MacOSWasmLibUrl = "https://raw.githubusercontent.com/meshplus/bitxhub/master/build/libwasmer.dylib"

	BitxhubUrlLinux = "https://github.com/meshplus/bitxhub/releases/download/v1.0.0-rc3/bitxhub_linux_amd64.tar.gz"
	BitxhubUrlMacOS = "https://github.com/meshplus/bitxhub/releases/download/v1.0.0-rc3/bitxhub_macos_x86_64.tar.gz"

	Ethereum = "ethereum"
	Fabric   = "fabric"
)
