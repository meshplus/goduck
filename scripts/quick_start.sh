#!/usr/bin/env bash

set -e

CURRENT_PATH=$(pwd)
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

function print_blue() {
  printf "${BLUE}%s${NC}\n" "$1"
}

function print_red() {
  printf "${RED}%s${NC}\n" "$1"
}

function printHelp() {
  print_blue "Usage:  "
  echo "  quick_start <mode>"
  echo "    <mode> - one of 'up', 'transfer'"
  echo "      - 'up' - bring up the demo interchain system"
  echo "      - 'transfer' - invoke demo transfer event"
  echo "  quick_start.sh -h (print this message)"
}

function docker-compose-up() {
  if [ ! "$(docker network ls -f name=quick_start_default)" ]; then
    print_blue "======> Start the demo service...."
    docker-compose -f ./docker/quick_start/quick_start.yml up
  else
    print_blue "======> Restart the demo service...."
    docker-compose -f ./docker/quick_start/quick_start.yml restart
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
