module github.com/meshplus/goduck

go 1.14

require (
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78 // indirect
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/Nvveen/Gotty v0.0.0-20120604004816-cd527374f1e5 // indirect
	github.com/Rican7/retry v0.1.0
	github.com/cavaliercoder/grab v2.0.0+incompatible
	github.com/cheggaaa/pb v2.0.7+incompatible
	github.com/cheynewallace/tabby v1.1.0
	github.com/codeskyblue/go-sh v0.0.0-20200712050446-30169cf553fe
	github.com/coreos/etcd v3.3.18+incompatible
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v1.4.2-0.20180625184442-8e610b2b55bf
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/go-units v0.4.0 // indirect
	github.com/ethereum/go-ethereum v1.10.6
	github.com/fatih/color v1.7.0
	github.com/gobuffalo/packd v1.0.0
	github.com/gobuffalo/packr v1.30.1
	github.com/gorilla/mux v1.8.0 // indirect
	github.com/hyperledger/fabric v2.0.1+incompatible
	github.com/hyperledger/fabric-sdk-go v1.0.0-beta1
	github.com/libp2p/go-libp2p-core v0.5.7-0.20200520175250-264788628f5a
	github.com/meshplus/bitxhub v1.1.0-rc1.0.20201020024116-dcdc23de5d04
	github.com/meshplus/bitxhub-kit v1.1.2-0.20210112075018-319e668d6359
	github.com/meshplus/go-libp2p-cert v0.0.0-20210125114242-7d9ed2eaaccd
	github.com/meshplus/gosdk v0.1.3
	github.com/mitchellh/go-homedir v1.1.0
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pelletier/go-toml v1.8.0
	github.com/pkg/errors v0.9.1
	github.com/shirou/gopsutil v3.21.4-0.20210419000835-c7a38de76ee5+incompatible
	github.com/sirupsen/logrus v1.6.0
	github.com/spf13/viper v1.6.1
	github.com/stretchr/testify v1.7.0
	github.com/ttacon/chalk v0.0.0-20160626202418-22c06c80ed31
	github.com/urfave/cli/v2 v2.3.0
	gopkg.in/VividCortex/ewma.v1 v1.1.1 // indirect
	gopkg.in/cheggaaa/pb.v2 v2.0.7 // indirect
	gopkg.in/fatih/color.v1 v1.7.0 // indirect
	gopkg.in/mattn/go-colorable.v0 v0.1.0 // indirect
	gopkg.in/mattn/go-isatty.v0 v0.0.4 // indirect
	gopkg.in/mattn/go-runewidth.v0 v0.0.4 // indirect
)

replace github.com/go-kit/kit => github.com/go-kit/kit v0.8.0

replace github.com/meshplus/bitxhub => github.com/meshplus/bitxhub v1.1.0-rc1.0.20201020024116-dcdc23de5d04

// todo(fbz): replace it with the official warehouse
replace github.com/meshplus/crypto-standard => github.com/dawn-to-dusk/crypto-standard v0.1.2-0.20210915031756-9c6750095d70
