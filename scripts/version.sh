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

MODE=$1
UPTYPE=$2
APPCHAINTYPE=$3
BXH_V_PATH="${CURRENT_PATH}/bitxhub/bitxhub-${UPTYPE}.version"
PIER_V_PATH="${CURRENT_PATH}/pier/pier-${APPCHAINTYPE}-${UPTYPE}.version"

function showBxhVersion() {
  cat $BXH_V_PATH
}

function showPierVersion() {
  cat $PIER_V_PATH
}

if [ "$MODE" == "bxh" ]; then
  showBxhVersion
elif [ "$MODE" == "pier" ]; then
  showPierVersion
else
  printHelp
  exit 1
fi
