#!/usr/bin/env bash

set -e

CURRENT_PATH=$(pwd)
RED='\033[0;31m'
GREEN='\033[0;32m'
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

function installInterchainChaincode() {
  prepareContract

  print_blue "===> Copy chaincode"

  docker cp ./contracts cli:/opt/gopath/src/github.com/hyperledger/fabric/peer

  print_blue "===> Install chaincode"

  cd "${CURRENT_PATH}"
  export CONFIG_PATH=${CURRENT_PATH}

  print_blue "===> 1. Deploying broker, transfer and data_swapper chaincode"

  print_blue "===> Step1 package"
  docker exec cli peer lifecycle chaincode package broker.tar.gz \
    --path /opt/gopath/src/github.com/hyperledger/fabric/peer/contracts/src/broker/ \
    --label broker

  docker exec cli peer lifecycle chaincode package transfer.tar.gz \
    --path /opt/gopath/src/github.com/hyperledger/fabric/peer/contracts/src/transfer/ \
    --label transfer

  docker exec cli peer lifecycle chaincode package data_swapper.tar.gz \
    --path /opt/gopath/src/github.com/hyperledger/fabric/peer/contracts/src/data_swapper/ \
    --label data_swapper

  if [ "${TYPE}" == "direct" ]; then
    docker exec cli peer lifecycle chaincode package transaction.tar.gz \
      --path /opt/gopath/src/github.com/hyperledger/fabric/peer/contracts/src/transaction/ \
      --label transaction
  fi

  print_blue "===> Step2 install"

  docker exec cli peer lifecycle chaincode install broker.tar.gz
  sleep 1
  docker exec \
    -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
    -e CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
    -e CORE_PEER_LOCALMSPID="Org2MSP" \
    -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    cli peer lifecycle chaincode install broker.tar.gz

  sleep 1
  docker exec cli peer lifecycle chaincode install transfer.tar.gz
  sleep 1
  docker exec \
    -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
    -e CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
    -e CORE_PEER_LOCALMSPID="Org2MSP" \
    -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    cli peer lifecycle chaincode install transfer.tar.gz

  sleep 1
  docker exec cli peer lifecycle chaincode install data_swapper.tar.gz
  sleep 1
  docker exec \
    -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
    -e CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
    -e CORE_PEER_LOCALMSPID="Org2MSP" \
    -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    cli peer lifecycle chaincode install data_swapper.tar.gz

  if [ "${TYPE}" == "direct" ]; then
    sleep 1
    docker exec cli peer lifecycle chaincode install transaction.tar.gz

    sleep 1
    docker exec \
      -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
      -e CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
      -e CORE_PEER_LOCALMSPID="Org2MSP" \
      -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
      cli peer lifecycle chaincode install transaction.tar.gz
  fi

  print_blue "===> Step3 query"
  sleep 1
  docker exec cli peer lifecycle chaincode queryinstalled >"$CURRENT_PATH"/identifier
  brokerId=$(grep -o 'broker.\{65\}' <./identifier)
  echo $brokerId
  transferId=$(grep -o 'transfer.\{65\}' <./identifier)
  echo $transferId
  dataSwapperId=$(grep -o 'data_swapper.\{65\}' <./identifier)
  echo $dataSwapperId
  if [ "${TYPE}" == "direct" ]; then
    transactionId=$(grep -o 'transaction.\{65\}' <./identifier)
    echo $transactionId
  fi

  print_blue "===> Step4 approveformyorg"
  sleep 1
  docker exec cli peer lifecycle chaincode approveformyorg \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
    --channelID mychannel --name broker --version 1 \
    --init-required --sequence 1 --waitForEvent --package-id $brokerId
  sleep 1
  docker exec \
    -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
    -e CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
    -e CORE_PEER_LOCALMSPID="Org2MSP" \
    -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    cli peer lifecycle chaincode approveformyorg \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
    --channelID mychannel --name broker --version 1 --init-required \
    --sequence 1 --waitForEvent --package-id $brokerId

  sleep 1
  docker exec cli peer lifecycle chaincode approveformyorg \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
    --channelID mychannel --name transfer --version 1 \
    --init-required --sequence 1 --waitForEvent --package-id $transferId
  sleep 1
  docker exec \
    -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
    -e CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
    -e CORE_PEER_LOCALMSPID="Org2MSP" \
    -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    cli peer lifecycle chaincode approveformyorg \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
    --channelID mychannel --name transfer --version 1 --init-required \
    --sequence 1 --waitForEvent --package-id $transferId

  sleep 1
  docker exec cli peer lifecycle chaincode approveformyorg \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
    --channelID mychannel --name data_swapper --version 1 \
    --init-required --sequence 1 --waitForEvent --package-id $dataSwapperId
  sleep 1
  docker exec \
    -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
    -e CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
    -e CORE_PEER_LOCALMSPID="Org2MSP" \
    -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    cli peer lifecycle chaincode approveformyorg \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
    --channelID mychannel --name data_swapper --version 1 --init-required \
    --sequence 1 --waitForEvent --package-id $dataSwapperId
  if [ "${TYPE}" == "direct" ]; then
    sleep 1
    docker exec cli peer lifecycle chaincode approveformyorg \
      --tls \
      --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
      --channelID mychannel --name transaction --version 1 \
      --init-required --sequence 1 --waitForEvent --package-id $transactionId
    sleep 1
    docker exec \
      -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
      -e CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
      -e CORE_PEER_LOCALMSPID="Org2MSP" \
      -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
      cli peer lifecycle chaincode approveformyorg \
      --tls \
      --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
      --channelID mychannel --name transaction --version 1 --init-required \
      --sequence 1 --waitForEvent --package-id $transactionId

  fi

  print_blue "===> Step5 commit"
  sleep 1
  docker exec \
    -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
    -e CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
    -e CORE_PEER_LOCALMSPID="Org2MSP" \
    -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    cli peer lifecycle chaincode commit -o orderer.example.com:7050 \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
    --peerAddresses peer0.org1.example.com:7051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
    --peerAddresses peer0.org2.example.com:9051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    --channelID mychannel --name broker --version 1 --sequence 1 --init-required

  sleep 1
  docker exec \
    -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
    -e CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
    -e CORE_PEER_LOCALMSPID="Org2MSP" \
    -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    cli peer lifecycle chaincode commit -o orderer.example.com:7050 \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
    --peerAddresses peer0.org1.example.com:7051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
    --peerAddresses peer0.org2.example.com:9051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    --channelID mychannel --name transfer --version 1 --sequence 1 --init-required

  sleep 1
  docker exec \
    -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
    -e CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
    -e CORE_PEER_LOCALMSPID="Org2MSP" \
    -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    cli peer lifecycle chaincode commit -o orderer.example.com:7050 \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
    --peerAddresses peer0.org1.example.com:7051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
    --peerAddresses peer0.org2.example.com:9051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    --channelID mychannel --name data_swapper --version 1 --sequence 1 --init-required
  if [ "${TYPE}" == "direct" ]; then
    sleep 1
    docker exec \
      -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
      -e CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
      -e CORE_PEER_LOCALMSPID="Org2MSP" \
      -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
      cli peer lifecycle chaincode commit -o orderer.example.com:7050 \
      --tls \
      --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
      --peerAddresses peer0.org1.example.com:7051 \
      --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
      --peerAddresses peer0.org2.example.com:9051 \
      --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
      --channelID mychannel --name transaction --version 1 --sequence 1 --init-required
  fi

  print_blue "===> Step6 Init"
  sleep 1
  docker exec \
    -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
    -e CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
    -e CORE_PEER_LOCALMSPID="Org2MSP" \
    -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    cli peer chaincode invoke -o orderer.example.com:7050 \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
    --peerAddresses peer0.org1.example.com:7051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
    --peerAddresses peer0.org2.example.com:9051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    -C mychannel -n broker --isInit -c '{"Args":[]}'

  sleep 1
  docker exec \
    -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
    -e CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
    -e CORE_PEER_LOCALMSPID="Org2MSP" \
    -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    cli peer chaincode invoke -o orderer.example.com:7050 \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
    --peerAddresses peer0.org1.example.com:7051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
    --peerAddresses peer0.org2.example.com:9051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    -C mychannel -n transfer --isInit -c '{"Args":[]}'

  sleep 1
  docker exec \
    -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
    -e CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
    -e CORE_PEER_LOCALMSPID="Org2MSP" \
    -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    cli peer chaincode invoke -o orderer.example.com:7050 \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
    --peerAddresses peer0.org1.example.com:7051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
    --peerAddresses peer0.org2.example.com:9051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    -C mychannel -n data_swapper --isInit -c '{"Args":[]}'

  if [ "${TYPE}" == "direct" ]; then
    sleep 1
    docker exec \
      -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
      -e CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
      -e CORE_PEER_LOCALMSPID="Org2MSP" \
      -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
      cli peer chaincode invoke -o orderer.example.com:7050 \
      --tls \
      --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
      --peerAddresses peer0.org1.example.com:7051 \
      --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
      --peerAddresses peer0.org2.example.com:9051 \
      --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
      -C mychannel -n transaction --isInit -c '{"Args":[]}'

  fi

  sleep 1

  print_blue "===> 2. Set Alice 10000 amout in transfer chaincode"
  sleep 1
  goduck fabric contract invoke transfer setBalance Alice,10000

  sleep 1
  print_blue "===> 3. Set (key: path, value: ${CURRENT_PATH}) in data_swapper chaincode"
  goduck fabric contract invoke data_swapper path,"'"${CURRENT_PATH}"'"

  print_blue "===> 4. Register transfer and data_swapper chaincode to broker chaincode"
  sleep 1
  goduck fabric contract invoke transfer register

  sleep 1
  goduck fabric contract invoke broker audit mychannel,transfer,1

  sleep 1
  goduck fabric contract invoke data_swapper register

  sleep 1
  goduck fabric contract invoke broker audit mychannel,data_swapper,1

}

function installNormalChaincode() {
  cd "${CURRENT_PATH}"
  export CONFIG_PATH=${CURRENT_PATH}

  mkdir -p contracts/src
  if [ ! -d "${CHAINCODE_PATH}" ]; then
    print_red "chaincode path is not valid project directory"
    exit 1
  fi
  docker cp -r "${CHAINCODE_PATH}" cli:/opt/gopath/src/github.com/hyperledger/fabric/peer
  CCName=$(basename "${CHAINCODE_PATH}")

  docker exec cli peer lifecycle chaincode package "${CCName}".tar.gz \
    --path /opt/gopath/src/github.com/hyperledger/fabric/peer/contracts/src/"${CCName}"/ \
    --label "${CCName}"

  docker exec cli peer lifecycle chaincode install "${CCName}".tar.gz

  docker exec \
    -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
    -e CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
    -e CORE_PEER_LOCALMSPID="Org2MSP" \
    -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    cli peer lifecycle chaincode install "${CCName}".tar.gz

  docker exec cli peer lifecycle chaincode queryinstalled >"$CURRENT_PATH"/identifier
  ADDRESS=$(grep -o '"${CCName}".\{65\}' <./identifier)

  docker exec cli peer lifecycle chaincode approveformyorg \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
    --channelID mychannel --name "${CCName}" --version 1 \
    --init-required --sequence 1 --waitForEvent --package-id $ADDRESS

  docker exec \
    -e CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/users/Admin@org2.example.com/msp \
    -e CORE_PEER_ADDRESS=peer0.org2.example.com:9051 \
    -e CORE_PEER_LOCALMSPID="Org2MSP" \
    -e CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    cli peer lifecycle chaincode approveformyorg \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
    --channelID mychannel --name "${CCName}" --version 1 --init-required \
    --sequence 1 --waitForEvent --package-id $ADDRESS

  docker exec cli peer lifecycle chaincode commit -o orderer.example.com:7050 \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
    --peerAddresses peer0.org1.example.com:7051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
    --peerAddresses peer0.org2.example.com:9051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    --channelID mychannel --name "${CCName}" --version 1 --sequence 1 --init-required

  docker exec cli peer chaincode invoke -o orderer.example.com:7050 \
    --tls \
    --cafile /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/example.com/orderers/orderer.example.com/msp/tlscacerts/tlsca.example.com-cert.pem \
    --peerAddresses peer0.org1.example.com:7051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org1.example.com/peers/peer0.org1.example.com/tls/ca.crt \
    --peerAddresses peer0.org2.example.com:9051 \
    --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/org2.example.com/peers/peer0.org2.example.com/tls/ca.crt \
    -C mychannel -n "${CCName}" --isInit -c '{"Args":[]}'

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
  b)
    BXH_VERSION=$OPTARG
    ;;
  t)
    TYPE=$OPTARG
    ;;
  esac
done

if [ "$MODE" == "install" ]; then
  installChaincode
else
  printHelp
  exit 1
fi
