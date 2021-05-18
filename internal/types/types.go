package types

const (
	TypeBinary     = "binary"
	TypeDocker     = "docker"
	ClusterMode    = "cluster"
	SoloMode       = "solo"
	PierModeDirect = "direct"
	PierModeRelay  = "relay"
	PierModeUnion  = "union"

	ChainTypeEther  = "ethereum"
	ChainTypeFabric = "fabric"

	PlaygroundScript = "playground.sh"
	FabricScript     = "fabric.sh"
	ChaincodeScript  = "chaincode.sh"
	EthereumScript   = "ethereum.sh"
	PierScript       = "run_pier.sh"
	QuickStartScript = "quick_start.sh"
	InfoScript       = "info.sh"
	Prometheus       = "prometheus.sh"
	TlsCerts         = "certs"
	TmpPath          = "tmp"

	Pier         = "pier"
	EthClient    = "eth-client"
	FabricClient = "fabric-client-1.4"
	RuleName     = "validating.wasm"

	LinuxWasmLibUrl = "https://raw.githubusercontent.com/meshplus/bitxhub/master/build/libwasmer.so"
	MacOSWasmLibUrl = "https://raw.githubusercontent.com/meshplus/bitxhub/master/build/libwasmer.dylib"

	BitxhubUrlLinux            = "https://github.com/meshplus/bitxhub/releases/download/%s/bitxhub_linux-amd64_%s.tar.gz"
	BitxhubUrlMacOS            = "https://github.com/meshplus/bitxhub/releases/download/%s/bitxhub_darwin_x86_64_%s.tar.gz"
	PierUrlLinux               = "https://github.com/meshplus/pier/releases/download/%s/pier_linux-amd64_%s.tar.gz"
	PierUrlMacOS               = "https://github.com/meshplus/pier/releases/download/%s/pier_darwin_x86_64_%s.tar.gz"
	PierFabricClientUrlLinux   = "https://github.com/meshplus/pier-client-fabric/releases/download/%s/fabric-client-%s-Linux"
	PierFabricClientUrlMacOS   = "https://github.com/meshplus/pier-client-fabric/releases/download/%s/fabric-client-%s-Darwin"
	PierEthereumClientUrlLinux = "https://github.com/meshplus/pier-client-ethereum/releases/download/%s/eth-client-%s-Linux"
	PierEthereumClientUrlMacOS = "https://github.com/meshplus/pier-client-ethereum/releases/download/%s/eth-client-%s-Darwin"

	FabricRuleUrl   = "https://raw.githubusercontent.com/meshplus/pier-client-fabric/master/config/validating.wasm"
	EthereumRuleUrl = "https://raw.githubusercontent.com/meshplus/pier-client-ethereum/master/config/validating.wasm"

	FabricConfig = "config.yaml"
)
