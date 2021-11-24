#!/usr/bin/env bash

set -e
source x.sh
source retry.sh

CURRENT_PATH=$(pwd)
SYSTEM=$(uname -s)
if [ $SYSTEM == "Linux" ]; then
  SYSTEM="linux"
elif [ $SYSTEM == "Darwin" ]; then
  SYSTEM="darwin"
fi

function showBxhAllBinary() {
  if [ $N = "-1" ]; then
    tail -n +1 $REPO_PATH/logs/bitxhub.log
  else
    tail -n -$N $REPO_PATH/logs/bitxhub.log
  fi
}

function showBxhAllDocker() {
  CID=$(docker container ls | grep meshplus/bitxhub | awk '{print $1}' | head -n $NUM | tail -n 1)
  if [ $N = "-1" ]; then
    docker logs --tail -1 $CID
  else
    docker logs --tail $N $CID
  fi
}

function showBxhNetBinary() {
  if [ $N = "-1" ]; then
    tail -n +1 $REPO_PATH/logs/bitxhub.log | grep "module=p2p" || true
  else
    tail -n -$N $REPO_PATH/logs/bitxhub.log | grep "module=p2p" || true
  fi
}

function showBxhNetDocker() {
  CID=$(docker container ls | grep meshplus/bitxhub | awk '{print $1}' | head -n $NUM | tail -n 1)
  if [ $N = "-1" ]; then
    docker logs --tail -1 $CID | grep "p2p" || true
  else
    docker logs --tail $N $CID | grep "p2p" || true
  fi
}

function showBxhOrderBinary() {
  if [ $N = "-1" ]; then
    tail -n +1 $REPO_PATH/logs/bitxhub.log | grep "module=order" || true
  else
    tail -n -$N $REPO_PATH/logs/bitxhub.log | grep "module=order" || true
  fi
}

function showBxhOrderDocker() {
  CID=$(docker container ls | grep meshplus/bitxhub | awk '{print $1}' | head -n $NUM | tail -n 1)
  if [ $N = "-1" ]; then
    docker logs --tail -1 $CID | grep "order" || true
  else
    docker logs --tail $N $CID | grep "order" || true
  fi
}

function showPierLogBinary() {
  if [ $N = "-1" ]; then
    tail -n +1 $REPO_PATH/logs/pier.log
  else
    tail -n -$N $REPO_PATH/logs/pier.log
  fi
}

function showPierLogDocker() {
  CID=$(docker container ls | grep pier-${APPCHAIN} | awk '{print $1}'| head -n $NUM | tail -n 1)
  if [ $N = "-1" ]; then
    docker logs --tail -1 $CID
  else
    docker logs --tail $N $CID
  fi
}

MODE=$1
UP_TYPE=$2
REPO_PATH=$3
N=$4
NUM=$5
APPCHAIN=$6

if [ "$MODE" == "bxhAll" ]; then
  if [ "$UP_TYPE" == "binary" ]; then
    showBxhAllBinary
  else
    showBxhAllDocker
  fi
elif [ "$MODE" == "bxhNet" ]; then
  if [ "$UP_TYPE" == "binary" ]; then
    showBxhNetBinary
  else
    showBxhNetDocker
  fi
elif [ "$MODE" == "bxhOrder" ]; then
  if [ "$UP_TYPE" == "binary" ]; then
    showBxhOrderBinary
  else
    showBxhOrderDocker
  fi
elif [ "$MODE" == "pierLog" ]; then
  if [ "$UP_TYPE" == "binary" ]; then
    showPierLogBinary
  else
    showPierLogDocker
  fi
else
  printHelp
  exit 1
fi
