package types

const (
	TypeBinary     = "binary"
	TypeDocker     = "docker"
	ClusterMode    = "cluster"
	SoloMode       = "solo"
	PierModeDirect = "direct"
	PierModeRelay  = "relay"

	ChainTypeEther  = "ethereum"
	ChainTypeFabric = "fabric"

	PlaygroundScript = "playground.sh"
	FabricScript     = "fabric.sh"
	ChaincodeScript  = "chaincode.sh"
	EthereumScript   = "ethereum.sh"
	PierScript       = "run_pier.sh"
	QuickStartScript = "quick_start.sh"
	InfoScript       = "info.sh"

	LinuxWasmLibUrl = "https://raw.githubusercontent.com/meshplus/bitxhub/master/build/libwasmer.so"
	MacOSWasmLibUrl = "https://raw.githubusercontent.com/meshplus/bitxhub/master/build/libwasmer.dylib"

	BitxhubUrlLinux = "https://github.com/meshplus/bitxhub/releases/download/v1.0.0-rc1/bitxhub_linux-amd64_v1.0.0-rc1.tar.gz"
	BitxhubUrlMacOS = "https://github.com/meshplus/bitxhub/releases/download/v1.0.0-rc1/bitxhub_macos_x86_64_v1.0.0-rc1.tar.gz"
	PierUrlLinux    = "https://github.com/meshplus/pier/releases/download/v1.0.0-rc1/pier-linux-amd64.tar.gz"
	PierUrlMacOS    = "https://github.com/meshplus/pier/releases/download/v1.0.0-rc1/pier-macos-x86-64.tar.gz"

	Ethereum = "ethereum"
	Fabric   = "fabric"
)
