#!/usr/bin/env bash

set -e

WORKDIR=ethereum

function print_blue() {
  printf "${BLUE}%s${NC}\n" "$1"
}

function printHelp() {
  print_blue "Usage:  "
  echo "  private_chain.sh <mode>"
  echo "    <mode> - one of 'binary', 'docker'"
  echo "      - 'binary' - bring up the ethereum with local binary geth"
  echo "      - 'docker' - clear the ethereum with geth in docker"
  echo "      - 'down' - shut down ethereum binary and docker container"
  echo "  private_chain.sh -h (print this message)"
}

function binaryUp() {
  # clean up datadir
  cd ${WORKDIR}
  rm -rf datadir
  tar xvf datadir.tar.gz

  nohup geth --datadir $HOME/.goduck/ethereum/datadir --dev --ws --rpc \
      --rpccorsdomain https://remix.ethereum.org \
      --wsaddr "0.0.0.0" --rpcaddr "0.0.0.0" --rpcport 8545 \
      --rpcapi "eth,web3,personal,net,miner,admin,debug" >/dev/null 2>&1 &
  echo $! >ethereum.pid
}

function dockerUp() {
  docker run -d --name ethereum-node \
  -p 8545:8545 -p 8546:8546 -p 30303:30303 \
  meshplus/ethereum \
      --datadir /root/datadir --dev --ws --rpc \
      --rpccorsdomain https://remix.ethereum.org \
      --rpcaddr "0.0.0.0" --rpcport 8545 --wsaddr "0.0.0.0" \
      --rpcapi "eth,web3,personal,net,miner,admin,debug"
}

function etherDown() {
  set +e
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

  if [ "$(docker container ls | grep -c ethereum-node)" -ge 1 ]; then
    docker stop ethereum-node
    echo "ethereum docker stop"
  fi
}

MODE=$1

if [ "$MODE" == "binary" ]; then
  binaryUp
elif [ "$MODE" == "docker" ]; then
  dockerUp
elif [ "$MODE" == "down" ]; then
  etherDown
else
  printHelp
  exit 1
fi
