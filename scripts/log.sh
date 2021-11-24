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

function showBxhAll() {
  if [ $N = "-1" ]; then
    tail -n +1 $REPO_PATH/logs/bitxhub.log
  else
    tail -n -$N $REPO_PATH/logs/bitxhub.log
  fi
}

function showBxhNet() {
  if [ $N = "-1" ]; then
    tail -n +1 $REPO_PATH/logs/bitxhub.log | grep "module=p2p" || true
  else
    tail -n -$N $REPO_PATH/logs/bitxhub.log | grep "module=p2p" || true
  fi
}

function showBxhOrder() {
  if [ $N = "-1" ]; then
    tail -n +1 $REPO_PATH/logs/bitxhub.log | grep "module=order" || true
  else
    tail -n -$N $REPO_PATH/logs/bitxhub.log | grep "module=order" || true
  fi
}

function showPierLog() {
  if [ $N = "-1" ]; then
    tail -n +1 $REPO_PATH/logs/pier.log
  else
    tail -n -$N $REPO_PATH/logs/pier.log
  fi
}

MODE=$1
REPO_PATH=$2
N=$3

if [ "$MODE" == "bxhAll" ]; then
  showBxhAll
elif [ "$MODE" == "bxhNet" ]; then
  showBxhNet
elif [ "$MODE" == "bxhOrder" ]; then
  showBxhOrder
elif [ "$MODE" == "pierLog" ]; then
  showPierLog
else
  printHelp
  exit 1
fi
