#!/usr/bin/env bash

set -e
source x.sh
source compare.sh
source retry.sh

MODE=$1
BITXHUB_ADDR=$2
PROMETHEUS=$3
VERSION=$4
ETHVERSION=$5

CURRENT_PATH=$(pwd)
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'
QUICK_PATH="${CURRENT_PATH}/docker/quick_start"
QUICK_PATH_TMP="${QUICK_PATH}/.quick_start"
QUICK_BXH_CONFIG_PATH="${QUICK_PATH}/bxhConfig/${VERSION}"
PROM_PATH="${CURRENT_PATH}/docker/prometheus"
ETH_PATH="${CURRENT_PATH}/pier/ethereum/${ETHVERSION}"
PLUGIN_PATH="${CURRENT_PATH}/bin/pier_linux_${VERSION}/ethereum-client"
BITXHUB_CONFIG_PATH="${CURRENT_PATH}"/bitxhub
PIER_CONFIG_PATH="${CURRENT_PATH}"/pier
PIER_SCRIPTS_PATH="${CURRENT_PATH}"/docker/pier
BITXHUB_SCRIPTS_PATH="${CURRENT_PATH}"/docker/bitxhub
SYSTEM=$(uname -s)

function printHelp() {
  print_blue "Usage:  "
  echo "  quick_start <mode>"
  echo "    <mode> - one of 'up', 'down', 'stop', 'transfer'"
  echo "      - 'up' - bring up the demo interchain system"
  echo "      - 'down' - bring down the demo interchain system"
  echo "      - 'stop' - stop containers in demo interchain system"
  echo "      - 'transfer' - invoke demo transfer event"
  echo "  quick_start.sh -h (print this message)"
}

function docker-compose-up() {
  chmod +x $PIER_SCRIPTS_PATH/registerAppchain.sh
  chmod +x $PIER_SCRIPTS_PATH/deployRule.sh
  chmod +x $PIER_SCRIPTS_PATH/getPierId.sh
  chmod +x $BITXHUB_SCRIPTS_PATH/vote.sh
  chmod +x ${PLUGIN_PATH}

  if [ -z $VERSION ]; then
    print_red "Please specify version！"
    exit 0
  fi

  if [ ! -d "$QUICK_PATH_TMP" ]; then
    mkdir "$QUICK_PATH_TMP"
  fi

  # rewrite bitxhub.toml
  cp $QUICK_BXH_CONFIG_PATH/bitxhub.toml $QUICK_PATH_TMP/bitxhub.toml
  x_replace "s/solo = false/solo = true/g" "${QUICK_PATH_TMP}"/bitxhub.toml
  x_replace "s/raft/solo/g" "${QUICK_PATH_TMP}"/bitxhub.toml
  x_replace "s/address = .*/address = \"$BITXHUB_ADDR\"/g" "${QUICK_PATH_TMP}"/bitxhub.toml
  x_replace "s/dider = .*/dider = \"$BITXHUB_ADDR\"/g" "${QUICK_PATH_TMP}"/bitxhub.toml
  x_replace "s/bvm_gas_price = .*/bvm_gas_price = 0/g" "${QUICK_PATH_TMP}"/bitxhub.toml
  delete_line_start=$(sed -n "/genesis.admins/=" "${QUICK_PATH_TMP}"/bitxhub.toml | head -n 2 | tail -n 1)
  delete_line_end=$(sed -n '/weight/=' "${QUICK_PATH_TMP}"/bitxhub.toml | head -n 4 | tail -n 1)
  x_replace "${delete_line_start},${delete_line_end}d" "${QUICK_PATH_TMP}"/bitxhub.toml

  # rewrite ethereum.toml
  if [ -d $QUICK_PATH_TMP/ethereum1 ]; then
    rm -r $QUICK_PATH_TMP/ethereum1
  fi
  cp -r $ETH_PATH $QUICK_PATH_TMP/ethereum1
  x_replace "s/{{.AppchainAddr}}/ws:\/\/host.docker.internal:8546/" $QUICK_PATH_TMP/ethereum1/ethereum.toml
  x_replace "s/{{.AppchainContractAddr}}/0xD3880ea40670eD51C3e3C0ea089fDbDc9e3FBBb4/" $QUICK_PATH_TMP/ethereum1/ethereum.toml
  if [ -d $QUICK_PATH_TMP/ethereum2 ]; then
    rm -r $QUICK_PATH_TMP/ethereum2
  fi
  cp -r $ETH_PATH $QUICK_PATH_TMP/ethereum2
  x_replace "s/{{.AppchainAddr}}/ws:\/\/host.docker.internal:8548/" $QUICK_PATH_TMP/ethereum2/ethereum.toml
  x_replace "s/{{.AppchainContractAddr}}/0xD3880ea40670eD51C3e3C0ea089fDbDc9e3FBBb4/" $QUICK_PATH_TMP/ethereum2/ethereum.toml

  # rewrite quick_start.yml
  cp $QUICK_PATH/quick_start.yml $QUICK_PATH_TMP/quick_start.yml
  x_replace "s/quickStartVersion/${VERSION}/g" "${QUICK_PATH_TMP}"/quick_start.yml

  x_replace "s/image: meshplus\/ethereum:.*/image: meshplus\/ethereum:${ETHVERSION}/g" "${QUICK_PATH_TMP}"/quick_start.yml

  if [ $SYSTEM == "Darwin" ]; then
    IP=$(ifconfig -a | grep -e "inet[^6]" | sed -e "s/.*inet[^6][^0-9]*\([0-9.]*\)[^0-9]*.*/\1/" | grep -v "^127\.")
  elif [ $SYSTEM == "Linux" ]; then
    IP=$(ip -4 route list | grep docker0 | awk '{print $9}')
  else
    print_red "Bitxhub does not support the current operating system"
    exit 0
  fi
  x_replace "s/host.docker.internal:0.0.0.0/host.docker.internal:$IP/g" "${QUICK_PATH_TMP}"/quick_start.yml

  ETH_PATH_TMP1=$(echo $QUICK_PATH_TMP/ethereum1 | sed 's/\//\\\//g')
  x_replace "s/ethereum-path1/$ETH_PATH_TMP1/g" "${QUICK_PATH_TMP}"/quick_start.yml
  ETH_PATH_TMP2=$(echo $QUICK_PATH_TMP/ethereum2 | sed 's/\//\\\//g')
  x_replace "s/ethereum-path2/$ETH_PATH_TMP2/g" "${QUICK_PATH_TMP}"/quick_start.yml

  PLUGIN_PATH_TMP=$(echo "${PLUGIN_PATH}" | sed 's/\//\\\//g')
  x_replace "s/plugin-path/$PLUGIN_PATH_TMP/g" "${QUICK_PATH_TMP}"/quick_start.yml

  # start
  echo ${VERSION} >"${BITXHUB_CONFIG_PATH}"/bitxhub.version
  if [ ! "$(docker network ls -q -f name=quick_start_default)" ]; then
    print_blue "======> Start the demo service...."
    if [ "${PROMETHEUS}" == "true" ]; then
      docker-compose -f "${QUICK_PATH_TMP}"/quick_start.yml up -d
      sleep 5
      curl -X POST http://127.0.0.1:3000/api/datasources -H "Content-Type:application/json" -d '{"name":"Prometheus","type":"prometheus","url":"http://prom:9090","access":"proxy","isDefault":true}' 2>$PROM_PATH/datasources2.log 1>$PROM_PATH/datasources1.log
      curl -X POST http://127.0.0.1:3000/api/dashboards/db -H 'Accept: application/json' -H 'Content-Type: application/json' -H 'cache-control: no-cache' -d @$PROM_PATH/Go_Processes.json 2>$PROM_PATH/dashboards2.log 1>$PROM_PATH/dashboards1.log
    else
      docker-compose -f "${QUICK_PATH_TMP}"/quick_start.yml up -d bitxhub_solo ethereum-1 ethereum-2 pier-ethereum-1 pier-ethereum-2
      sleep 5
    fi
  else
    print_blue "======> Restart the demo service...."
    if [ "${PROMETHEUS}" == "true" ]; then
      docker-compose -f "${QUICK_PATH_TMP}"/quick_start.yml restart -d
      sleep 5
      curl -X POST http://127.0.0.1:3000/api/datasources -H "Content-Type:application/json" -d '{"name":"Prometheus","type":"prometheus","url":"http://prom:9090","access":"proxy","isDefault":true}' 2>$PROM_PATH/datasources2.log 1>$PROM_PATH/datasources1.log
      curl -X POST http://127.0.0.1:3000/api/dashboards/db -H 'Accept: application/json' -H 'Content-Type: application/json' -H 'cache-control: no-cache' -d @$PROM_PATH/Go_Processes.json 2>$PROM_PATH/dashboards2.log 1>$PROM_PATH/dashboards1.log
    else
      docker-compose -f "${QUICK_PATH_TMP}"/quick_start.yml restart bitxhub_solo ethereum-1 ethereum-2 pier-ethereum-1 pier-ethereum-2
      sleep 5
    fi
  fi

  # get cid
  bitxhubCID=$(docker ps -qf name="bitxhub_solo")
  pier1CID=$(docker ps -qf name="pier-ethereum-1")
  pier2CID=$(docker ps -qf name="pier-ethereum-2")
  docker exec $bitxhubCID bitxhub version >"${BITXHUB_CONFIG_PATH}"/bitxhub-docker.version
  docker exec $pier1CID pier version >"${PIER_CONFIG_PATH}"/pier-ethereum-docker.version

  error=false

  # register appchain
  print_blue "======> Register appchain...."
  if [ "${VERSION}" == "v1.8.0" ]; then
    command_retry "docker exec $bitxhubCID bitxhub client did init"
  fi

  if [ "${VERSION}" == "v1.11.0" ]; then
    command_retry "docker exec $bitxhubCID bitxhub client tx send --key /root/.bitxhub/key.json --to 0xD389be2C1e6cCC9fB33aDc2235af8b449e3d14B4 --amount 100000000000000000"
    command_retry "docker exec $bitxhubCID bitxhub client tx send --key /root/.bitxhub/key.json --to 0x4768E44fB5e85E1D86D403D767cC5898703B2e78 --amount 100000000000000000"
  fi

  command_retry "docker exec $pier1CID /root/.pier/scripts/registerAppchain.sh appchain1 chainA ethereum chainA-description 1.9.13 /root/.pier/ethereum/ether.validators consensusType "${VERSION}""
  command_retry "docker exec $pier2CID /root/.pier/scripts/registerAppchain.sh appchain2 chainB ethereum chainB-description 1.9.13 /root/.pier/ethereum/ether.validators consensusType "${VERSION}""

  pier1ID=$(echo $(command_retry "docker exec $pier1CID /root/.pier/scripts/getPierId.sh") | awk -F " " '{print $NF}')
  proposal11ID="${pier1ID}-0"
  command_retry "docker exec $bitxhubCID /root/.bitxhub/scripts/vote.sh $proposal11ID approve reason"

  pier2ID=$(echo $(command_retry "docker exec $pier2CID /root/.pier/scripts/getPierId.sh") | awk -F " " '{print $NF}')
  proposal21ID="${pier2ID}-0"
  command_retry "docker exec $bitxhubCID /root/.bitxhub/scripts/vote.sh $proposal21ID approve reason"

  # deploy rule
  print_blue "======> Deploy rule...."
  command_retry "docker exec $pier1CID /root/.pier/scripts/deployRule.sh /root/.pier/ethereum/validating.wasm appchain1 "${VERSION}""
  command_retry "docker exec $pier2CID /root/.pier/scripts/deployRule.sh /root/.pier/ethereum/validating.wasm appchain2 "${VERSION}""

  version1=${VERSION}
  version2="v1.7.0"
  version_compare
  if [[ $versionComPareRes -gt 0 ]]; then
    #  if [ "${VERSION}" \> "v1.7.0" ]; then
    proposal12ID="${pier1ID}-1"
    command_retry "docker exec $bitxhubCID /root/.bitxhub/scripts/vote.sh $proposal12ID approve reason"

    proposal22ID="${pier2ID}-1"
    command_retry "docker exec $bitxhubCID /root/.bitxhub/scripts/vote.sh $proposal22ID approve reason"
  fi

  if [ "${PROMETHEUS}" == "true" ]; then
    docker-compose -f "$QUICK_PATH_TMP"/quick_start.yml logs --follow bitxhub_solo ethereum-1 ethereum-2 pier-ethereum-1 pier-ethereum-2 prom grafana
  else
    docker-compose -f "$QUICK_PATH_TMP"/quick_start.yml logs --follow bitxhub_solo ethereum-1 ethereum-2 pier-ethereum-1 pier-ethereum-2
  fi

}

function docker-compose-down() {
  print_blue "======> Clean up the demo service...."
  cleanPierInfoFile
  cleanBxhInfoFile

  if [ -e "${QUICK_PATH_TMP}"/quick_start.yml ]; then
    docker-compose -f "${QUICK_PATH_TMP}"/quick_start.yml down
  fi
}

function cleanPierInfoFile() {
  PIER_CONFIG_PATH="${CURRENT_PATH}"/pier

  if [ -e "${PIER_CONFIG_PATH}"/pier-ethereum.pid ]; then
    rm "${PIER_CONFIG_PATH}"/pier-ethereum.pid
  fi
  if [ -e "${PIER_CONFIG_PATH}"/pier-ethereum-binary.addr ]; then
    rm "${PIER_CONFIG_PATH}"/pier-ethereum-binary.addr
  fi
  if [ -e "${PIER_CONFIG_PATH}"/pier-fabric.pid ]; then
    rm "${PIER_CONFIG_PATH}"/pier-fabric.pid
  fi
  if [ -e "${PIER_CONFIG_PATH}"/pier-fabric-binary.addr ]; then
    rm "${PIER_CONFIG_PATH}"/pier-fabric-binary.addr
  fi

  if [ -e "${PIER_CONFIG_PATH}"/pier-ethereum.cid ]; then
    rm "${PIER_CONFIG_PATH}"/pier-ethereum.cid
  fi
  if [ -e "${PIER_CONFIG_PATH}"/pier-ethereum-docker.addr ]; then
    rm "${PIER_CONFIG_PATH}"/pier-ethereum-docker.addr
  fi
  if [ -e "${PIER_CONFIG_PATH}"/pier-fabric.cid ]; then
    rm "${PIER_CONFIG_PATH}"/pier-fabric.cid
  fi
  if [ -e "${PIER_CONFIG_PATH}"/pier-fabric-docker.addr ]; then
    rm "${PIER_CONFIG_PATH}"/pier-fabric-docker.addr
  fi
}

function cleanBxhInfoFile() {
  BITXHUB_CONFIG_PATH="${CURRENT_PATH}"/bitxhub
  if [ -e "${BITXHUB_CONFIG_PATH}"/bitxhub.pid ]; then
    rm "${BITXHUB_CONFIG_PATH}"/bitxhub.pid
  fi
  if [ -e "${BITXHUB_CONFIG_PATH}"/bitxhub.cid ]; then
    rm "${BITXHUB_CONFIG_PATH}"/bitxhub.cid
  fi
  if [ -e "${BITXHUB_CONFIG_PATH}"/bitxhub.version ]; then
    rm "${BITXHUB_CONFIG_PATH}"/bitxhub.version
  fi
}

function docker-compose-stop() {
  print_blue "======> Stop the demo cluster...."
  if [ -e "${QUICK_PATH_TMP}"/quick_start.yml ]; then
    docker-compose -f "${QUICK_PATH_TMP}"/quick_start.yml stop
  fi
}

function queryAccount() {
  print_blue "Query Alice account in ethereum-1 appchain"
  goduck ether contract invoke \
    --key-path ./docker/quick_start/account.key --address http://localhost:8545 \
    --abi-path=./pier/ethereum/$1/transfer.abi  0x668a209Dc6562707469374B8235e37b8eC25db08 getBalance Alice
  print_blue "Query Alice account in ethereum-2 appchain"
  goduck ether contract invoke \
    --key-path ./docker/quick_start/account.key --address http://localhost:8547 \
    --abi-path=./pier/ethereum/$1/transfer.abi  0x668a209Dc6562707469374B8235e37b8eC25db08 getBalance Alice
}

function interchainTransfer() {
  print_blue "1. Query original accounts in appchains"
  queryAccount $1

  print_blue "2. Send 1 coin from Alice in ethereum-1 to Alice in ethereum-2"
  version1=$1
  version2="1.3.0"
  version_compare
  if [[ $versionComPareRes -lt 0 ]]; then
    goduck ether contract invoke \
      --key-path ./docker/quick_start/account.key --abi-path ./pier/ethereum/$1/transfer.abi \
      --address http://localhost:8545 \
     0x668a209Dc6562707469374B8235e37b8eC25db08 transfer 0x4768E44fB5e85E1D86D403D767cC5898703B2e78,0x668a209Dc6562707469374B8235e37b8eC25db08,Alice,Alice,1
  else
    goduck ether contract invoke \
      --key-path ./docker/quick_start/account.key --abi-path ./pier/ethereum/$1/transfer.abi \
      --address http://localhost:8545 \
      0x668a209Dc6562707469374B8235e37b8eC25db08 transfer did:bitxhub:appchain2:0x668a209Dc6562707469374B8235e37b8eC25db08,Alice,Alice,1
  fi


  sleep 4
  print_blue "3. Query accounts after the first-round invocation"
  queryAccount $1

  print_blue "4. Send 1 coin from Alice in ethereum-2 to Alice in ethereum-1"
  version1=$1
  version2="1.3.0"
  version_compare
  if [[ $versionComPareRes -lt 0 ]]; then
    #  if [ "${VERSION}" \< "v1.3.0" ]; then
    goduck ether contract invoke \
    --key-path ./docker/quick_start/account.key --abi-path ./pier/ethereum/$1/transfer.abi \
    --address http://localhost:8547 \
    0x668a209Dc6562707469374B8235e37b8eC25db08 transfer 0xD389be2C1e6cCC9fB33aDc2235af8b449e3d14B4,0x668a209Dc6562707469374B8235e37b8eC25db08,Alice,Alice,1
  else
    goduck ether contract invoke \
      --key-path ./docker/quick_start/account.key --abi-path ./pier/ethereum/$1/transfer.abi \
      --address http://localhost:8547 \
      0x668a209Dc6562707469374B8235e37b8eC25db08 transfer did:bitxhub:appchain1:0x668a209Dc6562707469374B8235e37b8eC25db08,Alice,Alice,1
  fi

  sleep 4
  print_blue "5. Query accounts after the second-round invocation"
  queryAccount $1
}

if [ "$MODE" == "up" ]; then
  docker-compose-up
elif [ "$MODE" == "down" ]; then
  docker-compose-down
elif [ "$MODE" == "stop" ]; then
  docker-compose-stop
elif [ "$MODE" == "transfer" ]; then
  interchainTransfer $2
else
  printHelp
  exit 1
fi
