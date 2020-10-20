#!/usr/bin/env bash

set -e

CURRENT_PATH=$(pwd)
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'
BITXHUB_PATH=${CURRENT_PATH}/bitxhub
PIER_PATH=${CURRENT_PATH}/pier

function print_blue() {
  printf "${BLUE}%s${NC}\n" "$1"
}

function print_red() {
  printf "${RED}%s${NC}\n" "$1"
}

function printHelp() {
  print_blue "Usage:  "
  echo "  info <mode>"
  echo "    <mode> - one of 'bitxhub', 'pier'"
  echo "      - 'bitxhub' - show basic info about bitxhub"
  echo "      - 'pier' - basic info about piers"
  echo "  info.sh (print this message)"
}

function showBxhInfo() {
  if [ -e ${BITXHUB_PATH}/bitxhub.pid ]; then
    if [ -d ${BITXHUB_PATH}/nodeSolo ]; then
      if [ "$(ps aux | grep bitxhub | grep -v grep | grep -v info)" ]; then
        print_blue "======> Show address of solo bitxhub node started in binary"
        bitxhub key address --path ${BITXHUB_PATH}/nodeSolo/certs/node.priv
      fi
    fi

    nodes=$(ls ${BITXHUB_PATH} | grep node | grep -v nodeSolo || true)
    if [ -n "$nodes" ]; then
      if [ "$(ps aux | grep bitxhub | grep -v grep | grep -v info)" ]; then
        print_blue "======> Show address of each bitxhub node started in binary"
        for n in $nodes ; do
          bitxhub key address --path ${BITXHUB_PATH}/$n/certs/node.priv
        done
      fi
    fi
  fi

  if [ "$(docker ps -q -f name=bitxhub_solo)" ]; then
    print_blue "======> address of solo bitxhub node started in docker"
    docker exec bitxhub_solo bitxhub key address --path /root/.bitxhub/certs/node.priv
  fi

  if [ "$(docker ps -q -f name=bitxhub_node)" ]; then
    print_blue "======> address of each bitxhub node started in docker"
    cids=$(docker ps -q -f name=bitxhub_node)
    i=0
    for container_id in $cids; do
      echo "node ${i} address:"
      docker exec $container_id bitxhub key address --path /root/.bitxhub/certs/node.priv
      i=$((i+1))
    done
  fi
}

function showPierInfo() {
  if [ "$(docker ps -q -f name=pier-ethereum)" ]; then
    print_blue "======> info about piers of ethereum in docker"
    cat ${CURRENT_PATH}/pier/pier-ethereum-docker.addr
  fi

  if [ "$(docker ps -q -f name=pier-fabric)" ]; then
    print_blue "======> info about piers of fabric in docker"
    cat ${CURRENT_PATH}/pier/pier-fabric-docker.addr
  fi

  if [ -d ${PIER_PATH}/.pier_ethereum ]; then
    if [ "$(ps aux | grep pier | grep -v grep | grep -v info)" ]; then
      print_blue "======> info about piers of ethereum in binary"
      cat ${CURRENT_PATH}/pier/pier-ethereum-binary.addr
    fi
  fi

  if [ -d ${PIER_PATH}/.pier_fabric ]; then
    if [ "$(ps aux | grep pier | grep -v grep | grep -v info)" ]; then
      print_blue "======> info about piers of fabric in binary"
      cat ${CURRENT_PATH}/pier/pier-fabric-binary.addr
    fi
  fi
}

MODE=$1

if [ "$MODE" == "bitxhub" ]; then
  showBxhInfo
elif [ "$MODE" == "pier" ]; then
  showPierInfo
else
  printHelp
  exit 1
fi
