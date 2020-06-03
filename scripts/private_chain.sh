#!/usr/bin/env bash

set -e

function print_blue() {
  printf "${BLUE}%s${NC}\n" "$1"
}

function printHelp() {
  print_blue "Usage:  "
  echo "  private_chain.sh <mode>"
  echo "    <mode> - one of 'binary', 'docker'"
  echo "      - 'binary' - bring up the ethereum with local binary geth"
  echo "      - 'docker' - clear the ethereum with geth in docker"
  echo "  private_chain.sh -h (print this message)"
}

function binaryUp() {
  # clean up datadir
  rm -rf datadir
  mkdir datadir

  # init genesis block
  geth init --datadir=datadir genesis.json
  cp account.key ./datadir/keystore/account.key

  nohup geth --datadir $HOME/.goduck/datadir --rpc \
      --rpccorsdomain https://remix.ethereum.org --rpcaddr "0.0.0.0" --rpcport 8545 \
      --rpcapi "eth,web3,personal,net,miner,admin,debug" \
      --allow-insecure-unlock --nodiscover \
      --unlock 0c7cd0feddf37a350530446bf3ebdddd447d2790 --password password \
      --mine --miner.threads=1 --etherbase=0c7cd0feddf37a350530446bf3ebdddd447d2790 > /dev/null 2>&1 &
}

function dockerUp() {
  rm -rf datadir
  tar xvf datadir.tar.gz

  docker run -d --name ethereum-node -v $HOME/.goduck:/root \
  -p 8545:8545 -p 30303:30303 \
  ethereum/client-go \
      --datadir /root/datadir --rpc --rpcaddr "0.0.0.0" --rpcport 8545 \
      --rpccorsdomain https://remix.ethereum.org \
      --rpcapi "eth,web3,personal,net,miner,admin,debug" \
      --allow-insecure-unlock --nodiscover \
      --unlock 0c7cd0feddf37a350530446bf3ebdddd447d2790 --password /root/password \
      --mine --miner.threads=1 --etherbase=0c7cd0feddf37a350530446bf3ebdddd447d2790
}

MODE=$1

if [ "$MODE" == "binary" ]; then
  binaryUp
elif [ "$MODE" == "docker" ]; then
  dockerUp
else
  printHelp
  exit 1
fi
