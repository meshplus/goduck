#!/usr/bin/env bash

set -e

VERSION=1.0
CURRENT_PATH=$(pwd)
FABRIC_SAMPLE_PATH=${CURRENT_PATH}/fabric-samples
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

# The sed commend with system judging
# Examples:
# sed -i 's/a/b/g' bob.txt => x_replace 's/a/b/g' bob.txt
function x_replace() {
  system=$(uname)

  if [ "${system}" = "Linux" ]; then
    sed -i "$@"
  else
    sed -i '' "$@"
  fi
}

function print_blue() {
  printf "${BLUE}%s${NC}\n" "$1"
}

function printHelp() {
  print_blue "Usage:  "
  echo "  fabric.sh <mode>"
  echo "    <mode> - one of 'up', 'down', 'restart'"
  echo "      - 'up' - bring up the fabric first network"
  echo "      - 'down' - clear the fabric first network"
  echo "      - 'restart' - restart the fabric first network"
  echo "  fabric.sh -h (print this message)"
}

function prepare() {
  if [ ! -d "${FABRIC_SAMPLE_PATH}"/bin ]; then
    print_blue "===> Download the necessary dependencies"
    curl -sSL https://raw.githubusercontent.com/hyperledger/fabric/master/scripts/bootstrap.sh | bash -s -- 1.4.3 1.4.3 0.4.18
  fi
  docker volume prune -f
}


function networkUp() {
  if [ "$(docker ps | grep hyperledger/fabric)" ]; then
    print_blue "fabric network already running, use old container..."
    exit 0
  fi

  prepare

  cd "${FABRIC_SAMPLE_PATH}"/first-network
  cp "${CURRENT_PATH}"/byfn.sh ./byfn.sh
  # choose to regenerate crypto-config or not
  if [ -z "${CRYPTO_CONFIG_PATH}" ]; then
    print_blue "fabric crypto-config not specified, use new generated crypto-config..."
    ./byfn.sh generate
  else
    print_blue "use existing crypto-config in "${CRYPTO_CONFIG_PATH}"..."
    if [  -d "crypto-config" ]; then
      rm -r ./crypto-config
    fi
    git clean -f -d
    cp -r "${CRYPTO_CONFIG_PATH}" ./crypto-config
  fi

  ./byfn.sh up -n

  rm -rf "${CURRENT_PATH}"/fabric/crypto-config
  mv "${FABRIC_SAMPLE_PATH}"/first-network/crypto-config "${CURRENT_PATH}"/fabric/crypto-config
}

function networkDown() {
  prepare
  # stop all fabric nodes
  cd "${FABRIC_SAMPLE_PATH}"/first-network
  ./byfn.sh down
}

function networkClean() {
  print_blue "===> stop fabric ..."
  networkDown

  print_blue "===> clean contract images ..."
  if [[ -n `docker ps -a |grep example.com-broker | awk '{print $1}'` ]]; then
    docker rm  `docker ps -a |grep example.com-broker | awk '{print $1}'`
  fi

  if [[ -n `docker ps -a |grep example.com-transfer | awk '{print $1}'` ]]; then
    docker rm  `docker ps -a |grep example.com-transfer | awk '{print $1}'`
  fi

  if [[ `docker ps -a |grep example.com-data_swapper | awk '{print $1}'` ]]; then
    docker rm  `docker ps -a |grep example.com-data_swapper | awk '{print $1}'`
  fi

  if [[ `docker images |grep example.com-broker | awk '{print $1}'` ]]; then
    docker rmi `docker images |grep example.com-broker | awk '{print $1}'`
  fi

  if [[ `docker images |grep example.com-transfer | awk '{print $1}'` ]]; then
    docker rmi `docker images |grep example.com-transfer | awk '{print $1}'`
  fi

  if [[ `docker images |grep example.com-data_swapper | awk '{print $1}'` ]]; then
    docker rmi `docker images |grep example.com-data_swapper | awk '{print $1}'`
  fi
}

function networkRestart() {
  prepare

  cd "${FABRIC_SAMPLE_PATH}"/first-network
  ./byfn.sh restart -n

}

print_blue "===> Script version: $VERSION"

MODE=$1
CRYPTO_CONFIG_PATH=$2

if [ "$MODE" == "up" ]; then
  networkUp
elif [ "$MODE" == "down" ]; then
  networkDown
elif [ "$MODE" == "clean" ]; then
  networkClean
elif [ "$MODE" == "restart" ]; then
  networkRestart
else
  printHelp
  exit 1
fi
