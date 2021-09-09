#!/usr/bin/env bash

set -e

SYSTEM=$(uname -s)
if [ $SYSTEM == "Linux" ]; then
  SYSTEM="linux"
elif [ $SYSTEM == "Darwin" ]; then
  SYSTEM="darwin"
fi

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

function print_blue() {
  printf "${BLUE}%s${NC}\n" "$1"
}

function print_green() {
  printf "${GREEN}%s${NC}\n" "$1"
}

function print_red() {
  printf "${RED}%s${NC}\n" "$1"
}

# The sed commend with system judging
# Examples:
# sed -i 's/a/b/g' bob.txt => x_replace 's/a/b/g' bob.txt
function x_replace() {
  system=$(uname)

  if [ "${system}" = "Linux" ]; then
    sed -i "$@"
  else
    sed -i '' "$@"
  fi
}

printHelp() {
  awk -F'### ' '/^###/ { print $2 }' "$0"
}

function InitConfig() {
  readConfig
  generateNodesConfig
  rewriteNodesConfig
} 

function readConfig() {
  print_blue "======> read config"
  MODE=`sed '/^.*mode/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
  NUM=`sed '/^.*num/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
  AGENCYPRIVPATH=`sed '/^.*agency_priv_path/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
  AGENCYCERTPATH=`sed '/^.*agency_cert_path/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
  CACERTPATH=`sed '/^.*ca_cert_path/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
  CONSENSUSTYPE=`sed '/^.*consensus_type/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`

  JSONRPCP=`sed '/^.*jsonrpc_port/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
  GRPCP=`sed '/^.*grpc_port/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
  GATEWAYP=`sed '/^.*gateway_port/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
  PPROFP=`sed '/^.*pprof_port/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
  MONITORP=`sed '/^.*monitor_port/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`

  NODEHOST=`sed '/^.*node_host/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`

  for a in $JSONRPCP ; do
    JSONRPCPS+=($a)
  done

  for a in $GRPCP ; do
    GRPCPS+=($a)
  done

  for a in $GATEWAYP ; do
    GATEWAYPS+=($a)
  done

  for a in $PPROFP ; do
    PPROFPS+=($a)
  done

  for a in $MONITORP ; do
    MONITORPS+=($a)
  done

  for a in $NODEHOST ; do
    host_lables+=($a)
  done

  NUM_1=0
  if [ $NUM != 1 ]; then
    NUM_1=`expr $NUM - 1`
  fi
  for (( i=0; i<=$NUM_1; i++ )); do
    host_lable=${host_lables[$i]}
    host_lable_start=`sed -n "/$host_lable/=" ${CONFIGPATH} | head -n 1` #要求配置文件中第一个配置项是关于host的配置
    host_ip=`sed -n "$host_lable_start,/ip/p" ${CONFIGPATH} | tail -n 1 | sed 's/[[:space:]]//g;s/ip=//g'`
    IPS+=($host_ip)
  done
}

function generateNodesConfig() {
  for (( i = 1; i <= ${NUM}; i++ )); do
    if [ ${NUM} == 1 ]; then
      generateNodeConfig $i "nodeSolo"
    else
      generateNodeConfig $i "node$i"
    fi
  done
}

# $1 : node number
# $2 : node startup repo
function generateNodeConfig() {
  if [ "${SYSTEM}" == "linux" ]; then
    export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:"${BITXHUBBINPATH}"/
  elif [ "${SYSTEM}" == "darwin" ]; then
    install_name_tool -change @rpath/libwasmer.dylib "${BITXHUBBINPATH}"/libwasmer.dylib "${BITXHUBBINPATH}"/bitxhub
  else
    print_red "Bitxhub does not support the current operating system"
  fi

  print_blue "======> Generate configuration files for node $1"

  print_blue "【1】generate certs"
  ${BITXHUBBINPATH}/bitxhub cert priv gen --name node --target ${TARGET}/$2/certs
  ${BITXHUBBINPATH}/bitxhub cert csr --key ${TARGET}/$2/certs/node.priv --org $2 --target ${TARGET}/$2/certs
  ${BITXHUBBINPATH}/bitxhub cert issue --csr ${TARGET}/$2/certs/node.csr --is_ca false --key ${AGENCYPRIVPATH} --cert ${AGENCYCERTPATH} --target ${TARGET}/$2/certs
  rm ${TARGET}/$2/certs/node.csr
  cp ${AGENCYCERTPATH} ${TARGET}/$2/certs
  cp ${CACERTPATH} ${TARGET}/$2/certs

  print_blue "【2】generate key"
  ${BITXHUBBINPATH}/bitxhub key gen --name key --target ${TARGET}/$2/certs
  ${BITXHUBBINPATH}/bitxhub --repo ${TARGET}/$2 key convert --priv ${TARGET}/$2/certs/key.priv --save

  print_blue "【3】generate configuration files"
  ${BITXHUBBINPATH}/bitxhub --repo ${TARGET}/$2 init

  # Obtain node pid and address
  PID=`${BITXHUBBINPATH}/bitxhub cert priv pid --path ${TARGET}/$2/certs/node.priv`
  pid_array+=(${PID})
  ADRR=`${BITXHUBBINPATH}/bitxhub key address --path ${TARGET}/$2/certs/key.priv`
  addr_array+=(${ADRR})

  print_blue "【4】copy consensus plugin"
  mkdir ${TARGET}/$2/plugins
  if [ $CONSENSUSTYPE == "solo" ]; then
    cp ${BITXHUBBINPATH}/solo.so ${TARGET}/$2/plugins/solo.so
  elif [ $CONSENSUSTYPE == "raft" ]; then
    cp ${BITXHUBBINPATH}/raft.so ${TARGET}/$2/plugins/raft.so
  else
    print_red "Plugins of the type are not provided. You need to copy the plugin into the corresponding directory after the command is executed"
  fi
}

function rewriteNodesConfig() {
  for (( n = 1; n <= ${NUM}; n++ )); do
    if [ ${NUM} == 1 ]; then
      rewriteNodeConfig $n "nodeSolo"
    else
      rewriteNodeConfig $n "node$n"
    fi
  done
}

# $1 : node number
# $2 : node startup repo
function rewriteNodeConfig() {
  print_blue "======> Rewrite config for node $1"

  print_blue "【1】rewrite bitxhub.toml"
  # port
  jsonrpc=${JSONRPCPS[$1-1]}
  x_replace "s/jsonrpc.*= .*/jsonrpc = $jsonrpc/" ${TARGET}/$2/bitxhub.toml
  grpc=${GRPCPS[$1-1]}
  x_replace "s/grpc.*= .*/grpc = $grpc/" ${TARGET}/$2/bitxhub.toml
  gateway=${GATEWAYPS[$1-1]}
  x_replace "s/gateway.*= .*/gateway = $gateway/" ${TARGET}/$2/bitxhub.toml
  pprof=${PPROFPS[$1-1]}
  x_replace "s/pprof.*= .*/pprof = $pprof/" ${TARGET}/$2/bitxhub.toml
  monitor=${MONITORPS[$1-1]}
  x_replace "s/monitor.*= .*/monitor = $monitor/" ${TARGET}/$2/bitxhub.toml
  # mode
  if [ $MODE == "solo" ]; then
    x_replace "s/solo.*= .*/solo = true/" ${TARGET}/$2/bitxhub.toml
  else
    x_replace "s/solo.*= .*/solo = false/" ${TARGET}/$2/bitxhub.toml
  fi
  # order
  x_replace "s/plugin.*= .*/plugin = \"plugins\/$CONSENSUSTYPE\.so\"/" ${TARGET}/$2/bitxhub.toml

  # gas_price
  x_replace "s/bvm_gas_price.*= .*/bvm_gas_price= 0/" ${TARGET}/$2/bitxhub.toml

  # genesis
  dider_addr=${addr_array[0]}
  x_replace "s/dider.*= .*/dider = \"$dider_addr\"/" ${TARGET}/$2/bitxhub.toml
  if [ $NUM -gt 4 ]; then
    admin_start=`sed -n '/\[\[genesis.admins\]\]/=' ${TARGET}/$2/bitxhub.toml | head -n 1`
    for (( i = 4; i < $NUM; i++ )); do
      x_replace "$admin_start i\\
    weight = 2
" ${TARGET}/$2/bitxhub.toml
      x_replace "$admin_start i\\
    address = \" \"
" ${TARGET}/$2/bitxhub.toml
      x_replace "$admin_start i\\
  [[genesis.admins]]
" ${TARGET}/$2/bitxhub.toml
    done
  fi

  if [ $NUM -lt 4 ]; then
    NUM_ADD=$(expr $NUM + 1)
    delete_line_start=$(sed -n "/genesis.admins/=" ${TARGET}/$2/bitxhub.toml | head -n $NUM_ADD | tail -n 1)
    delete_line_end=$(sed -n "/weight/=" ${TARGET}/$2/bitxhub.toml | head -n 4 | tail -n 1)
    x_replace "${delete_line_start},${delete_line_end}d" ${TARGET}/$2/bitxhub.toml
  fi

  for (( i = 1; i <= ${NUM}; i++ )); do
    addr=${addr_array[$i-1]}
    addr_line=`sed -n "/address = \".*\"/=" ${TARGET}/$2/bitxhub.toml | head -n $i | tail -n 1`
    x_replace "$addr_line s/address = \".*\"/address = \"$addr\"/" ${TARGET}/$2/bitxhub.toml
  done

  if [ $MODE == "cluster" ]; then
    print_blue "【2】rewrite network.toml"
    x_replace "1 s/id = .*/id = $1/" ${TARGET}/$2/network.toml #要求第一行配置是自己的id
    x_replace "s/n = .*/n = $NUM/" ${TARGET}/$2/network.toml
    # nodes
    if [ $NUM -gt 4 ]; then
      nodes_start=`sed -n '/\[\[nodes\]\]/=' ${TARGET}/$2/network.toml | head -n 1`
      for (( i = 4; i < $NUM; i++ )); do
        x_replace "$nodes_start i\\
    pid = \" \"
" ${TARGET}/$2/network.toml
        x_replace "$nodes_start i\\
    id = 1
" ${TARGET}/$2/network.toml
        x_replace "$nodes_start i\\
    hosts = [\"\/\ip4\/127.0.0.1\/tcp\/4001\/p2p\/\"]
" ${TARGET}/$2/network.toml
        x_replace "$nodes_start i\\
    account = \" \"
" ${TARGET}/$2/network.toml
        x_replace "$nodes_start i\\
  [[nodes]]
" ${TARGET}/$2/network.toml
      done
    fi

    for (( i = 1; i <= ${NUM}; i++ )); do
      account=${addr_array[$i-1]}
      ip=${IPS[$i-1]}
      pid=${pid_array[$i-1]}

      # 要求配置项顺序一定
      a_line=`sed -n "/account = \".*\"/=" ${TARGET}/$2/network.toml | head -n $i | tail -n 1`
      x_replace "$a_line s/account = \".*\"/account = \"$account\"/" ${TARGET}/$2/network.toml
      host_line=`expr $a_line + 1`
      x_replace "$host_line s/hosts = .*/hosts = [\"\/\ip4\/$ip\/tcp\/400$i\/p2p\/\"]/" ${TARGET}/$2/network.toml
      id_line=`expr $a_line + 2`
      x_replace "$id_line s/id = .*/id = $i/" ${TARGET}/$2/network.toml
      pid_line=`expr $a_line + 3`
      x_replace "$pid_line s/pid = \".*\"/pid = \"$pid\"/" ${TARGET}/$2/network.toml
    done
  fi
}

# parses opts
while getopts "h?t:b:p:" opt; do
  case "$opt" in
  h | \?)
    printHelp
    exit 0
    ;;
  t)
    TARGET=$OPTARG
    ;;
  b)
    BITXHUBBINPATH=$OPTARG
    ;;
  p)
    CONFIGPATH=$OPTARG
    ;;
  esac
done

InitConfig

# 相比v1.9.0, 为了便于快速体验需要将gas_price设为0