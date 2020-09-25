#!/usr/bin/env bash

set -e
source x.sh

CURRENT_PATH=$(pwd)
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'
QUICK_PATH="${CURRENT_PATH}/docker/quick_start"
PROM_PATH="${CURRENT_PATH}/docker/prometheus"
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
  if [ -z $VERSION ]; then
      print_red "Please specify version！"
      exit 0
  fi
  version=${VERSION:1}
  quickConfig=$QUICK_PATH/quick_start.yml
  x_replace "s/bitxhub-solo:1.0.0-rc1/bitxhub-solo:${version}/g" "${quickConfig}"
  x_replace "s/pier-ethereum:1.0.0-rc1/pier-ethereum:${version}/g" "${quickConfig}"
  x_replace "s/ethereum:1.0.0/ethereum:${version%-*}/g" "${quickConfig}"

  if [ $SYSTEM == "Darwin" ]; then
    localIP=`ifconfig -a | grep -e "inet[^6]" | sed -e "s/.*inet[^6][^0-9]*\([0-9.]*\)[^0-9]*.*/\1/" | grep -v "^127\."`
  elif [ $SYSTEM == "Linux" ]; then
    localIP=`hostname -I | awk '{print $1}'`
  else
    print_red "Bitxhub does not support the current operating system"
    exit 0
  fi
  x_replace "s/host.docker.internal:0.0.0.0/host.docker.internal:$localIP/g" $QUICK_PATH/quick_start.yml

  if [ ! "$(docker network ls -q -f name=quick_start_default)" ]; then
    print_blue "======> Start the demo service...."
    docker-compose -f ./docker/quick_start/quick_start.yml up -d
  else
    print_blue "======> Restart the demo service...."
    docker-compose -f ./docker/quick_start/quick_start.yml restart -d
  fi

  sleep 3
  docker-compose-check
  curl -X POST http://127.0.0.1:3000/api/datasources -H "Content-Type:application/json" -d '{"name":"Prometheus","type":"prometheus","url":"http://prom:9090","access":"proxy","isDefault":true}'
  curl -X POST http://127.0.0.1:3000/api/dashboards/db -H 'Accept: application/json' -H 'Content-Type: application/json' -H 'cache-control: no-cache' -d @$PROM_PATH/Go_Processes.json
  echo ""
}

function docker-compose-check() {
  if [[ -z `docker ps -qf "name=bitxhub_solo"` ]]; then
    print_red "===> Fail to start bitxhub_solo!!!"
  else
    print_green "===> Start bitxhub_solo successfully!!!"
  fi

  if [[ -z `docker ps -qf "name=ethereum-1"` ]]; then
    print_red "===> Fail to start ethereum-1!!!"
  else
    print_green "===> Start ethereum-1 successfully!!!"
  fi

  if [[ -z `docker ps -qf "name=ethereum-2"` ]]; then
    print_red "===> Fail to start ethereum-2!!!"
  else
    print_green "===> Start ethereum-2 successfully!!!"
  fi

  if [[ -z `docker ps -qf "name=pier-ethereum-1"` ]]; then
    print_red "===> Fail to start pier-ethereum-1!!!"
  else
    print_green "===> Start pier-ethereum-1 successfully!!!"
  fi

  if [[ -z `docker ps -qf "name=pier-ethereum-2"` ]]; then
    print_red "===> Fail to start pier-ethereum-2!!!"
  else
    print_green "===> Start pier-ethereum-2 successfully!!!"
  fi

  if [[ -z `docker ps -qf "name=prometheus"` ]]; then
    print_red "===> Fail to start prometheus!!!"
  else
    print_green "===> Start prometheus successfully!!!"
  fi

  if [[ -z `docker ps -qf "name=grafana"` ]]; then
    print_red "===> Fail to start grafana!!!"
  else
    print_green "===> Start grafana successfully!!!"
  fi
}

function docker-compose-down() {
  print_blue "======> Clean up the demo service...."
  docker-compose -f ./docker/quick_start/quick_start.yml down
}

function docker-compose-stop() {
  print_blue "======> Stop the demo cluster...."
  docker-compose -f ./docker/quick_start/quick_start.yml stop
}

function queryAccount() {
  print_blue "Query Alice account in ethereum-1 appchain"
  goduck ether contract invoke \
    --key_path ./docker/quick_start/account.key --ether_addr http://localhost:8545 \
    --abi_path=./docker/quick_start/transfer.abi 0x668a209Dc6562707469374B8235e37b8eC25db08 getBalance Alice
  print_blue "Query Alice account in ethereum-2 appchain"
  goduck ether contract invoke \
    --key_path ./docker/quick_start/account.key --ether_addr http://localhost:8547 \
    --abi_path=./docker/quick_start/transfer.abi 0x668a209Dc6562707469374B8235e37b8eC25db08 getBalance Alice
}

function interchainTransfer() {
  print_blue "1. Query original accounts in appchains"
  queryAccount

  print_blue "2. Send 1 coin from Alice in ethereum-1 to Alice in ethereum-2"
  goduck ether contract invoke \
    --key_path ./docker/quick_start/account.key --abi_path ./docker/quick_start/transfer.abi \
    --ether_addr http://localhost:8545 \
    0x668a209Dc6562707469374B8235e37b8eC25db08 transfer 0x9f5cf4b97965ababe19fcf3f1f12bb794a7dc279,0x668a209Dc6562707469374B8235e37b8eC25db08,Alice,Alice,1

  sleep 2
  print_blue "3. Query accounts after the first-round invocation"
  queryAccount

  print_blue "4. Send 1 coin from Alice in ethereum-2 to Alice in ethereum-1"
  goduck ether contract invoke \
    --key_path ./docker/quick_start/account.key --abi_path ./docker/quick_start/transfer.abi \
    --ether_addr http://localhost:8547 \
    0x668a209Dc6562707469374B8235e37b8eC25db08 transfer 0xb132702a7500507411f3bd61ab33d9d350d41a37,0x668a209Dc6562707469374B8235e37b8eC25db08,Alice,Alice,1

  sleep 2
  print_blue "5. Query accounts after the second-round invocation"
  queryAccount
}

MODE=$1
VERSION=$2

if [ "$MODE" == "up" ]; then
  docker-compose-up
elif [ "$MODE" == "down" ]; then
  docker-compose-down
elif [ "$MODE" == "stop" ]; then
  docker-compose-stop
elif [ "$MODE" == "transfer" ]; then
  interchainTransfer
else
  printHelp
  exit 1
fi
