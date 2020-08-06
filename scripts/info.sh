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
  echo "  info <mode>"
  echo "    <mode> - one of 'bitxhub', 'pier'"
  echo "      - 'bitxhub' - show basic info about bitxhub"
  echo "      - 'pier' - basic info about piers"
  echo "  info.sh (print this message)"
}

function showBxhInfo() {
  if [ -d ${CURRENT_PATH}/nodeSolo ]; then
    print_blue "======> Show address of solo bitxhub node started in binary"
    bitxhub key address --path ${CURRENT_PATH}/nodeSolo/certs/node.priv
  fi

  if [ -d ${CURRENT_PATH}/nodeSolo ]; then
    print_blue "======> Show address of each bitxhub node started in binary"
    nodes=$(ls ${CURRENT_PATH}/)
    bitxhub key address --path ${CURRENT_PATH}/nodeSolo/certs/node.priv
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
    piers=$(docker ps -q -f name=pier-ethereum)
    for pier in ${piers[@]}; do
      print_blue "pier id of ethereum is as follow: "
      docker exec $pier pier --repo=/root/.pier id
    done
  fi

  if [ "$(docker ps -q -f name=pier-fabric)" ]; then
    print_blue "======> info about piers of fabric in docker"
    piers=$(docker ps -q -f name=pier-fabric)
    for pier in $piers; do
      print_blue "pier id of fabric is as follow: "
      docker exec $pier pier --repo=/root/.pier id
    done
  fi

  if [ -d ${CURRENT_PATH}/.pier_ethereum ]; then
    print_blue "======> info about piers of ethereum in binary"
    ${CURRENT_PATH}/bin/pier --repo=${CURRENT_PATH}/.pier_ethereum id
  fi

  if [ -d ${CURRENT_PATH}/.pier_fabric ]; then
    print_blue "======> info about piers of fabric in binary"
    ${CURRENT_PATH}/bin/pier --repo=${CURRENT_PATH}/.pier_fabric id
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
