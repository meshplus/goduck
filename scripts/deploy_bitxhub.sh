#!/usr/bin/env bash

### deploy_bitxhxub.sh -- Deploys bitxhub to remote server
###
### Usage:
###     deploy_bitxhxub.sh [server_ips server_username server_password version] [options]
###
###     e.g. deploy_bitxhxub.sh 188.0.0.1,188.0.0.2,188.0.0.3,188.0.0.4 bitxhub v1.0.0-rc1
###
### Options:
###     -h      Show this message.

set -e

IPS=188.0.0.1,188.0.0.2,188.0.0.3,188.0.0.4
VERSION=v1.0.0-rc1
USERNAME=bitxhub
PASSWORD=bitxhub
REPO_PATH=~/.goduck
BIN_PATH=$REPO_PATH/bin

ARCH=$(echo "$(uname -s | tr '[:upper:]' '[:lower:]' | sed 's/mingw64_nt.*/windows/')-$(uname -m | sed 's/x86_64/amd64/g')")
MARCH=$(uname -m)

help() {
  awk -F'### ' '/^###/ { print $2 }' "$0"
}

download() {
  local NAME=$1
  local URL=$2
  echo "===> Downloading $URL"
  mkdir -p $BIN_PATH
  cd $BIN_PATH
  if [ -f "$NAME" ]; then
    return 0
  fi
  wget $URL
}

deploy() {
  NAME=$1
  local ips=$(echo $IPS | tr "," "\n")
  for IP in $ips; do
    ssh $USERNAME@"$IP" "rm $NAME"
    scp $BIN_PATH/"$NAME" $USERNAME@"$IP":~/
    ssh $USERNAME@"$IP" "tar xf $NAME"
  done
}

if [ -z "$1" ]; then
  help
  exit 0
fi

# parses arguments
if [ -n "$1" ] && [ "${1:0:1}" != "-" ]; then
  IPS=$1
  shift
  if [ -n "$1" ] && [ "${1:0:1}" != "-" ]; then
    USERNAME=$1
    shift
  fi
  if [ -n "$1" ] && [ "${1:0:1}" != "-" ]; then
    PASSWORD=$1
    shift
  fi
  if [ -n "$1" ] && [ "${1:0:1}" != "-" ]; then
    VERSION=$1
    shift
  fi
fi

# parses opts
while getopts "h?" opt; do
  case "$opt" in
  h | \?)
    help
    exit 0
    ;;
  esac
done

DOWNLOAD_NAME=bitxhub_linux-amd64_"$VERSION".tar.gz
DOWNLOAD_URL=https://github.com/meshplus/bitxhub/releases/download/"$VERSION"/bitxhub_linux-amd64_"$VERSION".tar.gz
download "$DOWNLOAD_NAME" "$DOWNLOAD_URL"

echo "===> Deploying bitxhub($VERSION) to $IPS"
deploy "$DOWNLOAD_NAME"
