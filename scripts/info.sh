#!/usr/bin/env bash

set -e
source compare.sh

CURRENT_PATH=$(pwd)
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'
BITXHUB_PATH=${CURRENT_PATH}/bitxhub
PIER_PATH=${CURRENT_PATH}/pier
SYSTEM=$(uname -s)
if [ $SYSTEM == "Linux" ]; then
  SYSTEM="linux"
elif [ $SYSTEM == "Darwin" ]; then
  SYSTEM="darwin"
fi

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
  bitxhubKeyName="certs/key.priv"
  if [ -e ${BITXHUB_PATH}/bitxhub.version ]; then
    version=`awk 'NR==1{print}' ${BITXHUB_PATH}/bitxhub.version`
    if [[ $version == "v1.0.0" || $version == "v1.0.0-rc1" || $version == "v1.1.0-rc1" ]]; then
      bitxhubKeyName="certs/node.priv"
    fi
    version1=$version
    version2=v1.11.0
    version_compare
    if [[ $versionComPareRes -ge 0 ]]; then
      bitxhubKeyName="key.json"
    fi
    binPath=${CURRENT_PATH}/bin/bitxhub_${SYSTEM}_$version
  fi

  if [ -e ${BITXHUB_PATH}/bitxhub.pid ]; then
    if [ -d ${BITXHUB_PATH}/.bitxhub/nodeSolo ]; then
      if [ "$(ps aux | grep bitxhub | grep -v grep | grep -v info)" ]; then
        print_blue "======> Show address of solo bitxhub node started in binary"
        $binPath/bitxhub key address --path ${BITXHUB_PATH}/.bitxhub/nodeSolo/$bitxhubKeyName
      fi
    fi

    nodes=$(ls ${BITXHUB_PATH}/.bitxhub | grep node | grep -v nodeSolo || true)
    if [ -n "$nodes" ]; then
      if [ "$(ps aux | grep bitxhub | grep -v grep | grep -v info)" ]; then
        print_blue "======> Show address of each bitxhub node started in binary"
        for n in $nodes ; do
          $binPath/bitxhub key address --path ${BITXHUB_PATH}/.bitxhub/$n/$bitxhubKeyName
        done
      fi
    fi
  fi

  if [ "$(docker ps -q -f name=bitxhub_solo)" ]; then
    print_blue "======> address of solo bitxhub node started in docker"
    docker exec bitxhub_solo bitxhub key address --path /root/.bitxhub/$bitxhubKeyName
  fi

  if [ "$(docker ps -q -f name=bitxhub_node)" ]; then
    print_blue "======> address of each bitxhub node started in docker"
    cids=$(docker ps -q -f name=bitxhub_node)
    i=1
    for container_id in $cids; do
      echo "node ${i} address:"
      docker exec $container_id bitxhub key address --path /root/.bitxhub/$bitxhubKeyName
      i=$((i+1))
    done
  fi
}

function showPierInfo() {
  if [ "$(docker ps -q -f name=pier-ethereum)" ]; then
    print_blue "======> info about piers of ethereum in docker"
    CIDS=`docker ps -qf "name=pier-ethereum"`
    echo $CIDS >"${CURRENT_PATH}/pier/pier-ethereum.cid"
    if [ -e "${CURRENT_PATH}"/pier/pier-ethereum-docker.addr ]; then
      rm "${CURRENT_PATH}"/pier/pier-ethereum-docker.addr
    fi
    for cid in $CIDS; do
      if [ cid != "" ]; then
        echo `docker exec $cid pier --repo=/root/.pier id` >>"${CURRENT_PATH}/pier/pier-ethereum-docker.addr"
      fi
    done
    cat ${CURRENT_PATH}/pier/pier-ethereum-docker.addr
  fi

  if [ "$(docker ps -q -f name=pier-fabric)" ]; then
    print_blue "======> info about piers of fabric in docker"
    CIDS=`docker ps -qf "name=pier-fabric"`
    echo $CIDS >"${CURRENT_PATH}/pier/pier-fabric.cid"
    if [ -e "${CURRENT_PATH}"/pier/pier-fabric-docker.addr ]; then
      rm "${CURRENT_PATH}"/pier/pier-fabric-docker.addr
    fi
    for cid in $CIDS; do
      if [ cid != "" ]; then
        echo `docker exec $cid pier --repo=/root/.pier id` >>"${CURRENT_PATH}/pier/pier-fabric-docker.addr"
      fi
    done
    cat ${CURRENT_PATH}/pier/pier-fabric-docker.addr
  fi

  if [ -e ${PIER_PATH}/pier-ethereum.pid ]; then
    if [ "$(ps aux | grep pier | grep -v grep | grep -v info)" ]; then
      print_blue "======> info about piers of ethereum in binary"
      cat ${CURRENT_PATH}/pier/pier-ethereum-binary.addr
    fi
  fi
  if [ -e ${PIER_PATH}/pier-fabric.pid ]; then
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
