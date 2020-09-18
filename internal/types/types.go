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

	BitxhubUrlLinux = "https://github.com/meshplus/bitxhub/releases/download/%s/bitxhub_linux-amd64_%s.tar.gz"
	BitxhubUrlMacOS = "https://github.com/meshplus/bitxhub/releases/download/%s/bitxhub_macos_x86_64_%s.tar.gz"
	PierUrlLinux    = "https://github.com/meshplus/pier/releases/download/%s/pier_linux-amd64_%s.tar.gz"
	PierUrlMacOS    = "https://github.com/meshplus/pier/releases/download/%s/pier_darwin_x86_64_%s.tar.gz"

	FabricRuleUrl   = "https://github.com/meshplus/bitxhub/blob/master/scripts/quick_start/fabric_rule.wasm"
	EthereumRuleUrl = "https://github.com/meshplus/pier-client-ethereum/blob/master/config/validating.wasm"

	Ethereum = "ethereum"
	Fabric   = "fabric"
)
