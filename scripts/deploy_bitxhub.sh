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
source x.sh
source compare.sh

CURRENT_PATH=$(pwd)
OPT=$1
VERSION=$2
MODIFY_CONFIG_PATH=$3
TARGET=$4
SYSTEM=$(uname -s)
if [ $SYSTEM == "Linux" ]; then
  SYSTEM="linux"
elif [ $SYSTEM == "Darwin" ]; then
  SYSTEM="darwin"
fi
BXH_DEPLOY_PATH="${CURRENT_PATH}"/bxh-deploy
BIN_FILE=bitxhub_linux-amd64_"${VERSION}".tar.gz
BIN_PATH="${CURRENT_PATH}/bin/bitxhub_linux_${VERSION}"

MODE=$(sed '/^.*mode/!d;s/.*=//;s/[[:space:]]//g' ${MODIFY_CONFIG_PATH})
NUM=$(sed '/^.*num/!d;s/.*=//;s/[[:space:]]//g' ${MODIFY_CONFIG_PATH})
REWRITE=$(sed '/^.*rewrite/!d;s/.*=//;s/[[:space:]]//g' ${MODIFY_CONFIG_PATH})

ARCH=$(echo "$(uname -s | tr '[:upper:]' '[:lower:]' | sed 's/mingw64_nt.*/windows/')-$(uname -m | sed 's/x86_64/amd64/g')")
MARCH=$(uname -m)

function printHelp() {
  awk -F'### ' '/^###/ { print $2 }' "$0"
}

function readHostConfig() {
  print_blue "read host config"
  GRPCP=`sed '/^.*grpc_port/!d;s/.*=//;s/[[:space:]]//g' ${MODIFY_CONFIG_PATH}`
  for a in $GRPCP ; do
    GRPCPS+=($a)
  done

  NODEHOST=`sed '/^.*node_host/!d;s/.*=//;s/[[:space:]]//g' ${MODIFY_CONFIG_PATH}`
  for a in $NODEHOST ; do
    host_lables+=($a)
  done

  NUM_1=0
  if [ $NUM != 1 ]; then
    NUM_1=`expr $NUM - 1`
  fi
  for (( i=0; i<=$NUM_1; i++ )); do
    host_lable=${host_lables[$i]}
    host_lable_start=`sed -n "/$host_lable/=" ${MODIFY_CONFIG_PATH} | head -n 1` #要求配置文件中第一个配置项是关于host的配置
    host_ip=`sed -n "$host_lable_start,/ip/p" ${MODIFY_CONFIG_PATH} | tail -n 1 | sed 's/[[:space:]]//g;s/ip=//g'`
    host_user=`sed -n "$host_lable_start,/user/p" ${MODIFY_CONFIG_PATH} | tail -n 1 | sed 's/[[:space:]]//g;s/user=//g'`
    IPS+=($host_ip)
    USERS+=($host_user)
  done
}

function prepareConfig() {
  flag=true
  if [[ -d ${BXH_DEPLOY_PATH}/nodeSolo && $MODE == "solo" ]]; then
    print_blue "BitXHub solo configuration file already exists"
    print_blue "reinitializing would overwrite your configuration? ($REWRITE)"
    flag=$REWRITE
    if [ $REWRITE == true ]; then
      rm -r ${BXH_DEPLOY_PATH}
    fi
  fi
  if [[ -d ${BXH_DEPLOY_PATH}/node1 && $MODE == "cluster" ]]; then
    print_blue "BitXHub cluster configuration file already exists"
    print_blue "reinitializing would overwrite your configuration? ($REWRITE)"
    flag=$REWRITE
    if [ $REWRITE == true ]; then
      rm -r ${BXH_DEPLOY_PATH}
    fi
  fi

  if [ $flag == true ]; then
    goduck bitxhub config --version $VERSION --target "${BXH_DEPLOY_PATH}" --configPath "${MODIFY_CONFIG_PATH}"
  fi
}

function check(){
  print_blue "==========> Checking"
  echo "You need to wait more than 5 seconds for each node"
  for (( i = 1; i <= ${NUM}; i++ )); do
    ip=${IPS[$i-1]}
    user=${USERS[$i-1]}
    WHO=$user@$ip
    FULL_TARGET=$WHO:$TARGET

    PID=`cat ${BXH_DEPLOY_PATH}/bitxhub$i.PID`
    error=false
    ssh $WHO "ps aux | grep $PID | grep -v grep | grep node$i" || error=true
    if [ ${error} == false ]; then
      print_green "start node$i successful"
    else
      print_red "start node$i fail"
    fi
  done
}

function deploy() {
  prepareConfig
  print_green "==========> Prepare config successful"
  print_blue "==========> Deploying bitxhub($VERSION)"
  readHostConfig
  for (( i = 1; i <= ${NUM}; i++ )); do
    ip=${IPS[$i-1]}
    user=${USERS[$i-1]}
    WHO=$user@$ip
    FULL_TARGET=$WHO:$TARGET
    echo "Operating at node$i: $user@$ip "

    # 传bin文件
    scp ${BIN_PATH}/${BIN_FILE} $FULL_TARGET
    # 解压bin
    ssh $WHO "if [[ -d ${TARGET}/.bitxhub ]]; then tar xzf ${TARGET}/${BIN_FILE} -C ${TARGET}/.bitxhub; else mkdir ${TARGET}/.bitxhub && tar xzf ${TARGET}/${BIN_FILE} -C ${TARGET}/.bitxhub; fi; "
  done

  for (( i = 1; i <= ${NUM}; i++ )); do
    ip=${IPS[$i-1]}
    user=${USERS[$i-1]}
    WHO=$user@$ip
    FULL_TARGET=$WHO:$TARGET
    echo "Operating at node$i: $user@$ip "
    # 传配置文件
    scp -r ${BXH_DEPLOY_PATH}/node$i $FULL_TARGET/.bitxhub/
  done

  for (( i = 1; i <= ${NUM}; i++ )); do
    echo "Runing node$i"
    ip=${IPS[$i-1]}
    user=${USERS[$i-1]}
    WHO=$user@$ip
    FULL_TARGET=$WHO:$TARGET

    # 启动
    ssh $WHO "export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${TARGET}/.bitxhub; cd ${TARGET}/.bitxhub; nohup ./bitxhub --repo=${TARGET}/.bitxhub/node$i start >/dev/null 2>&1 &"
    sleep 1

    grpc_port=${GRPCPS[$i-1]}
    error=false
    NODE=`ssh $WHO "sleep 3 ; lsof -i:$grpc_port | grep LISTEN"` || error=true
    NODE=${NODE#* } || error=true
    nodePid=`echo ${NODE} | awk '{print $1}'` || error=true
    echo $nodePid > ${BXH_DEPLOY_PATH}/bitxhub$i.PID || error=true
    if [ ${error} == false ]; then
      print_green "Start BitXHub node$i end"
    else
      print_red "Start BitXHub node$i fail"
    fi
  done

  check
}

if [ "$OPT" == "up" ]; then
  deploy
else
  printHelp
  exit 1
fi