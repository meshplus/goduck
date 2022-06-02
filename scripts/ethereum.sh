#!/usr/bin/env bash

set -e

CURRENT_PATH=$(pwd)
WORKDIR=ethereum
SYSTEM=$(uname -s)
if [ $SYSTEM == "Linux" ]; then
  SYSTEM="linux"
elif [ $SYSTEM == "Darwin" ]; then
  SYSTEM="darwin"
fi

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
  echo "  ethereum.sh <mode>"
  echo "    <mode> - one of 'binary', 'docker'"
  echo "      - 'binary' - bring up the ethereum with local binary geth"
  echo "      - 'docker' - clear the ethereum with geth in docker"
  echo "      - 'down' - shut down ethereum binary and docker container"
  echo "  ethereum.sh -h (print this message)"
}

function binaryUp() {
  # clean up datadir
  cd ${WORKDIR}
  rm -rf datadir
  tar xf datadir.tar.gz

  print_blue "start geth with datadir in ${WORKDIR}/datadir"
  nohup $CURRENT_PATH/bin/geth_${SYSTEM}_1.9.6/geth --datadir $CURRENT_PATH/ethereum/datadir --dev --ws --rpc \
    --rpccorsdomain https://remix.ethereum.org \
    --wsaddr "0.0.0.0" --rpcaddr "0.0.0.0" --rpcport $HTTP_PORT --wsport $WS_PORT --port $PORT \
    --rpcapi "eth,web3,personal,net,miner,admin,debug" >/dev/null 2>&1 &
  echo $! >ethereum.pid
}

function dockerUp() {
  if [ ! "$(docker ps -q -f name=ethereum-node-$HTTP_PORT-$WS_PORT-$PORT)" ]; then
    if [ "$(docker ps -aq -f status=exited -f name=ethereum-node-$HTTP_PORT-$WS_PORT-$PORT)" ]; then
      # restart your container
      print_blue "restart your ethereum-node container"
      docker restart ethereum-node
    else
      print_blue "start a new ethereum-node container"
      docker run -d --name ethereum-node-$HTTP_PORT-$WS_PORT-$PORT \
        -p $HTTP_PORT:8545 -p $WS_PORT:8546 -p $PORT:30303 \
        meshplus/ethereum:$VERSION \
        --datadir /root/datadir --dev --ws --rpc \
        --rpccorsdomain https://remix.ethereum.org \
        --rpcaddr "0.0.0.0" --rpcport 8545 --wsaddr "0.0.0.0" \
        --rpcapi "eth,web3,personal,net,miner,admin,debug"
    fi
  else
    print_red "ethereum-node is already running, use old container..."
  fi

}

function etherDown() {
  set +e
  print_blue "===> stop ethereum in binary..."
  if [ -a "${WORKDIR}"/ethereum.pid ]; then
    list=$(cat "${WORKDIR}"/ethereum.pid)
    for pid in $list; do
      kill "$pid"
      if [ $? -eq 0 ]; then
        echo "node pid:$pid exit"
      else
        print_red "program exit fail, try use kill -9 $pid"
      fi
    done
    rm "${WORKDIR}"/ethereum.pid
  fi

  print_blue "===> stop ethereum in docker..."
  if [ "$(docker container ls | grep -c ethereum-node)" -ge 1 ]; then
    docker kill $(docker container ls | grep ethereum-node | awk '{print $1}')
    echo "ethereum docker container stopped"
  fi
}

function etherClean() {
  set +e
  etherDown

  print_blue "===> clean ethereum in binary..."
  if [[ ! -z $(ps | grep $CURRENT_PATH/ethereum/datadir | grep -v "grep") ]]; then
    ethPid=$(ps | grep $CURRENT_PATH/ethereum/datadir | grep -v "grep")
    kill $ethPid
    if [ $? -eq 0 ]; then
      echo "ethereum pid:$ethPid exit"
    else
      print_red "ethereum exit fail, try use kill -9 $ethPid"
    fi
  fi

  print_blue "===> clean ethereum in docker..."
  if [ "$(docker container ls -a | grep -c ethereum-node)" -ge 1 ]; then
    docker rm $(docker container ls -a | grep ethereum-node | awk '{print $1}')
    echo "ethereum docker container cleaned"
  fi
}

MODE=$1
HTTP_PORT=$2
WS_PORT=$3
PORT=$4
VERSION=$5

if [ "$MODE" == "binary" ]; then
  binaryUp
elif [ "$MODE" == "docker" ]; then
  dockerUp
elif [ "$MODE" == "down" ]; then
  etherDown
elif [ "$MODE" == "clean" ]; then
  etherClean
else
  printHelp
  exit 1
fi
