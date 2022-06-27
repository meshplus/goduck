#!/usr/bin/env bash

set -e

CURRENT_PATH=$(pwd)
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

function print_red() {
  printf "${RED}%s${NC}\n" "$1"
}

function printHelp() {
  print_blue "Usage:  "
  echo "  chaincode.sh <mode> [-c <config_path>] [-v <chaincode_version>] [-t <target_appchain_id>]"
  echo "    <mode> - one of 'install', 'upgrade', 'init'"
  echo "      - 'install' - install broker, transfer and data_swapper chaincode"
  echo "      - 'upgrade <chaincode_version(default: v1)>' - upgrade broker, transfer and data_swapper chaincode"
  echo "    -c <config_path> - specify which config.yaml file use (default \"./config.yaml\")"
  echo "    -g <chaincode_path> - specify which chaincode project to use (default \"./contracts/broker)"
  echo "    -v <chaincode_version> - upgrade fabric chaincode version (default \"v1\")"
  echo "  chaincode.sh -h (print this message)"
}

function prepare() {
  if ! type fabric-cli >/dev/null 2>&1; then
    print_blue "===> Install fabric-cli"
    go get github.com/securekey/fabric-examples/fabric-cli/cmd/fabric-cli
  fi

  if [ ! -f config-template.yaml ]; then
    print_blue "===> Download config-template.yaml"
    wget https://raw.githubusercontent.com/meshplus/bitxhub/$BXH_VERSION/scripts/quick_start/config-template.yaml
  fi

  if [ ! -f config.yaml ]; then
    cp "${CURRENT_PATH}"/config-template.yaml "${CURRENT_PATH}"/config.yaml
    x_replace "s|\${CONFIG_PATH}|${CURRENT_PATH}|g" config.yaml
  fi

  if [ ! -d fabric/crypto-config ]; then
    print_red "===> Please provide the 'crypto-config' of your fabric network"
    exit 1
  fi
}

function prepareContract() {
  if [ ! -d contracts ]; then
    print_blue "===> Download chaincode"
    wget https://github.com/meshplus/pier-client-fabric/raw/$BXH_VERSION/example/contracts.zip
    unzip -q contracts.zip
    rm contracts.zip
  fi
}

function installChaincode() {
  if [ -z ${CONFIG_YAML} ]; then
    CONFIG_YAML=./config.yaml
  fi
  prepare

  if [ -z "${CHAINCODE_PATH}" ]; then
    print_blue "install default interchain chaincode"
    installInterchainChaincode
  else
    print_blue "install chaincode in path "${CHAINCODE_PATH}""
    installNormalChaincode
  fi
}

function installNormalChaincode() {
  cd "${CURRENT_PATH}"
  export CONFIG_PATH=${CURRENT_PATH}

  mkdir -p contracts/src
  if [ ! -d "${CHAINCODE_PATH}" ]; then
    print_red "chaincode path is not valid project directory"
    exit 1
  fi
  cp -r "${CHAINCODE_PATH}" contracts/src/
  CCName=$(basename "${CHAINCODE_PATH}")

  fabric-cli chaincode install --gopath ./contracts --ccp "${CCName}" --ccid "${CCName}" --config "${CONFIG_YAML}" --orgid org2 --user Admin --cid mychannel
  fabric-cli chaincode instantiate --ccp "${CCName}" --ccid "${CCName}" --config "${CONFIG_YAML}" --orgid org2 --user Admin --cid mychannel
}

function installInterchainChaincode() {
  prepareContract

  print_blue "===> Install chaincode"

  cd "${CURRENT_PATH}"
  export CONFIG_PATH=${CURRENT_PATH}

  print_blue "===> 1. Deploying broker, transfer and data_swapper chaincode"
  fabric-cli chaincode install --gopath ./contracts --ccp broker --ccid broker --config "${CONFIG_YAML}" --orgid org2 --user Admin --cid mychannel
  fabric-cli chaincode instantiate --ccp broker --ccid broker --config "${CONFIG_YAML}" --orgid org2 --user Admin --cid mychannel
  fabric-cli chaincode install --gopath ./contracts --ccp transfer --ccid transfer --config "${CONFIG_YAML}" --orgid org2 --user Admin --cid mychannel
  fabric-cli chaincode instantiate --ccp transfer --ccid transfer --config "${CONFIG_YAML}" --orgid org2 --user Admin --cid mychannel
  fabric-cli chaincode install --gopath ./contracts --ccp data_swapper --ccid data_swapper --config "${CONFIG_YAML}" --orgid org2 --user Admin --cid mychannel
  fabric-cli chaincode instantiate --ccp data_swapper --ccid data_swapper --config "${CONFIG_YAML}" --orgid org2 --user Admin --cid mychannel
  if [ "${TYPE}" == "direct" ]; then
      print_blue "===> Deploying transaction"
      fabric-cli chaincode install --gopath ./contracts --ccp transaction --ccid transaction --config "${CONFIG_YAML}" --orgid org2 --user Admin --cid mychannel
      fabric-cli chaincode instantiate --ccp transaction --ccid transaction --config "${CONFIG_YAML}" --orgid org2 --user Admin --cid mychannel
  fi

  print_blue "===> 2. Set Alice 10000 amout in transfer chaincode"
  fabric-cli chaincode invoke --cid mychannel --ccid=transfer \
    --args='{"Func":"setBalance","Args":["Alice", "10000"]}' \
    --user Admin --orgid org2 --payload --config "${CONFIG_YAML}"

  print_blue "===> 3. Set (key: path, value: ${CURRENT_PATH}) in data_swapper chaincode"
  fabric-cli chaincode invoke --cid mychannel --ccid=data_swapper \
    --args='{"Func":"set","Args":["path", "'"${CURRENT_PATH}"'"]}' \
    --user Admin --orgid org2 --payload --config "${CONFIG_YAML}"

  print_blue "===> 4. Register transfer and data_swapper chaincode to broker chaincode"
  fabric-cli chaincode invoke --cid mychannel --ccid=transfer \
    --args='{"Func":"register"}' --user Admin --orgid org2 --payload --config "${CONFIG_YAML}"
  fabric-cli chaincode invoke --cid mychannel --ccid=data_swapper \
    --args='{"Func":"register"}' --user Admin --orgid org2 --payload --config "${CONFIG_YAML}"

  print_blue "===> 6. Audit transfer and data_swapper chaincode"
  fabric-cli chaincode invoke --cid mychannel --ccid=broker \
    --args='{"Func":"audit", "Args":["mychannel", "transfer", "1"]}' \
    --user Admin --orgid org2 --payload --config "${CONFIG_YAML}"
  fabric-cli chaincode invoke --cid mychannel --ccid=broker \
    --args='{"Func":"audit", "Args":["mychannel", "data_swapper", "1"]}' \
    --user Admin --orgid org2 --payload --config "${CONFIG_YAML}"

}

function upgradeChaincode() {
  prepare

  print_blue "Upgrade to version: $CHAINCODE_VERSION"

  cd "${CURRENT_PATH}"
  export CONFIG_PATH=${CURRENT_PATH}


  print_blue "===> 1. Deploying broker, transfer and data_swapper chaincode"
  fabric-cli chaincode install --gopath ./contracts --ccp broker --ccid broker \
    --v $CHAINCODE_VERSION \
    --config "${CONFIG_YAML}" --orgid org2 --user Admin --cid mychannel
  fabric-cli chaincode upgrade --ccp broker --ccid broker \
    --v $CHAINCODE_VERSION \
    --config "${CONFIG_YAML}" --orgid org2 --user Admin --cid mychannel

  fabric-cli chaincode install --gopath ./contracts --ccp transfer --ccid transfer \
    --v $CHAINCODE_VERSION \
    --config "${CONFIG_YAML}" --orgid org2 --user Admin --cid mychannel
  fabric-cli chaincode upgrade --ccp transfer --ccid transfer \
    --v $CHAINCODE_VERSION \
    --config "${CONFIG_YAML}" --orgid org2 --user Admin --cid mychannel

  fabric-cli chaincode install --gopath ./contracts --ccp data_swapper --ccid data_swapper \
    --v $CHAINCODE_VERSION \
    --config "${CONFIG_YAML}" --orgid org2 --user Admin --cid mychannel
  fabric-cli chaincode upgrade --ccp data_swapper --ccid data_swapper \
    --v $CHAINCODE_VERSION \
    --config "${CONFIG_YAML}" --orgid org2 --user Admin --cid mychannel

  print_blue "===> 2. Set Alice 10000 amout in transfer chaincode"
  fabric-cli chaincode invoke --cid mychannel --ccid=transfer \
    --args='{"Func":"setBalance","Args":["Alice", "10000"]}' \
    --user Admin --orgid org2 --payload --config "${CONFIG_YAML}"

  print_blue "===> 3. Set (key: path, value: ${CURRENT_PATH}) in data_swapper chaincode"
  fabric-cli chaincode invoke --cid mychannel --ccid=data_swapper \
    --args='{"Func":"set","Args":["path", "'"${CURRENT_PATH}"'"]}' \
    --user Admin --orgid org2 --payload --config "${CONFIG_YAML}"

  print_blue "===> 4. Register transfer and data_swapper chaincode to broker chaincode"
  fabric-cli chaincode invoke --cid mychannel --ccid=transfer \
    --args='{"Func":"register"}' --user Admin --orgid org2 --payload --config "${CONFIG_YAML}"
  fabric-cli chaincode invoke --cid mychannel --ccid=data_swapper \
    --args='{"Func":"register"}' --user Admin --orgid org2 --payload --config "${CONFIG_YAML}"

  print_blue "===> 6. Audit transfer and data_swapper chaincode"
  fabric-cli chaincode invoke --cid mychannel --ccid=broker \
    --args='{"Func":"audit", "Args":["mychannel", "transfer", "1"]}' \
    --user Admin --orgid org2 --payload --config "${CONFIG_YAML}"
  fabric-cli chaincode invoke --cid mychannel --ccid=broker \
    --args='{"Func":"audit", "Args":["mychannel", "data_swapper", "1"]}' \
    --user Admin --orgid org2 --payload --config "${CONFIG_YAML}"
}

CONFIG_YAML=./config.yaml
CHAINCODE_PATH=""
CHAINCODE_VERSION=v1
BXH_VERSION=v1.6.5
TYPE="relay"

MODE=$1
shift

while getopts "h?c:g:v:b:t:" opt; do
  case "$opt" in
  h | \?)
    printHelp
    exit 0
    ;;
  c)
    CONFIG_YAML=$OPTARG
    ;;
  g)
    CHAINCODE_PATH=$OPTARG
    ;;
  v)
    CHAINCODE_VERSION=$OPTARG
    ;;
  b)BXH_VERSION=$OPTARG
    ;;
  t)TYPE=$OPTARG
    ;;
  esac
done

if [ "$MODE" == "install" ]; then
  installChaincode
elif [ "$MODE" == "upgrade" ]; then
  upgradeChaincode
else
  printHelp
  exit 1
fi

