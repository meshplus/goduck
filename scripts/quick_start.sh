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

function queryAccount() {
  print_blue "Query Alice account in ethereum-1 appchain"
  goduck ether contract invoke \
   --key_path ./docker/quick_start/account.key --ether_addr http://localhost:8545 \
   --abi_path=./transfer.abi 0x668a209Dc6562707469374B8235e37b8eC25db08 getBalance Alice
  print_blue "Query Alice account in ethereum-2 appchain"
  goduck ether contract invoke \
   --key_path ./docker/quick_start/account.key --ether_addr http://localhost:8547 \
   --abi_path=./transfer.abi 0x668a209Dc6562707469374B8235e37b8eC25db08 getBalance Alice
}

function interchainTransfer() {
  print_blue "1. Query original accounts in appchains"
  queryAccount

  print_blue "2. Send 1 coin from Alice in ethereum-1 to Alice in ethereum-2"
  goduck ether contract invoke \
  --key_path ./docker/quick_start/account.key --abi_path ./transfer.abi \
  --ether_addr http://localhost:8545 \
  0x668a209Dc6562707469374B8235e37b8eC25db08 transfer 0x9f5cf4b97965ababe19fcf3f1f12bb794a7dc279,0x668a209Dc6562707469374B8235e37b8eC25db08,Alice,Alice,1

  sleep 1
  print_blue "3. Query accounts after the first-round invocation"
  queryAccount

  print_blue "4. Send 1 coin from Alice in ethereum-2 to Alice in ethereum-1"
  goduck ether contract invoke \
  --key_path ./docker/quick_start/account.key --abi_path ./transfer.abi \
  --ether_addr http://localhost:8547 \
  0x668a209Dc6562707469374B8235e37b8eC25db08 transfer 0x9c13a0ee57f2b6f5c08c98a7395bcfc167dcde91,0x668a209Dc6562707469374B8235e37b8eC25db08,Alice,Alice,1

  sleep 1
  print_blue "5. Query accounts after the second-round invocation"
  queryAccount
}

MODE=$1

if [ "$MODE" == "up" ]; then
  docker-compose-up
elif [ "$MODE" == "transfer" ]; then
  interchainTransfer
else
  printHelp
  exit 1
fi