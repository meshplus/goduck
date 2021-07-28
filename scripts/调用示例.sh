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
  print_blue "Start the demo system...."
  docker-compose -f ./docker/quick_start/quick_start.yml up
}

function docker-compose-down() {
  print_blue "Stop the demo system...."
  docker-compose -f ./docker/quick_start/quick_start.yml down
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
  0x668a209Dc6562707469374B8235e37b8eC25db08 transfer 0xD389be2C1e6cCC9fB33aDc2235af8b449e3d14B4,0x668a209Dc6562707469374B8235e37b8eC25db08,Alice,Alice,1

  sleep 4
  print_blue "3. Query accounts after the first-round invocation"
  queryAccount

  print_blue "4. Send 1 coin from Alice in ethereum-2 to Alice in ethereum-1"
  goduck ether contract invoke \
  --key_path ./docker/quick_start/account.key --abi_path ./docker/quick_start/transfer.abi \
  --ether_addr http://localhost:8547 \
  0x668a209Dc6562707469374B8235e37b8eC25db08 transfer 0x570C2E736B28F04d621eF108C1D2f3DE06c71208,0x668a209Dc6562707469374B8235e37b8eC25db08,Alice,Alice,1

  sleep 4
  print_blue "5. Query accounts after the second-round invocation"
  queryAccount
}

MODE=$1

if [ "$MODE" == "up" ]; then
  docker-compose-up
elif [ "$MODE" == "down" ]; then
  docker-compose-down
elif [ "$MODE" == "transfer" ]; then
  interchainTransfer
else
  printHelp
  exit 1
fi