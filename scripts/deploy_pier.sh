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
CHAINTYPE=$2
VERSION=$3
MODIFY_CONFIG_PATH=$4
TARGET=$5
SYSTEM=$(uname -s)
if [ $SYSTEM == "Linux" ]; then
  SYSTEM="linux"
elif [ $SYSTEM == "Darwin" ]; then
  SYSTEM="darwin"
fi
PIER_DEPLOY_PATH="${CURRENT_PATH}"/pier-deploy
BIN_FILE=pier_linux-amd64_"${VERSION}".tar.gz
PLUGIN_BIN_FILE=$CHAINTYPE-client
BIN_PATH="${CURRENT_PATH}/bin/pier_linux_${VERSION}"

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
  PPROFPORT=`sed '/^.*pprofPort/!d;s/.*=//;s/[[:space:]]//g' ${MODIFY_CONFIG_PATH}`

  PIERHOST=`sed '/^.*pier_host/!d;s/.*=//;s/[[:space:]]//g' ${MODIFY_CONFIG_PATH}`
  host_lable_start=`sed -n "/$PIERHOST/=" ${MODIFY_CONFIG_PATH} | head -n 1` #要求配置文件中第一个配置项是关于host的配置
  HOST_IP=`sed -n "$host_lable_start,/ip/p" ${MODIFY_CONFIG_PATH} | tail -n 1 | sed 's/[[:space:]]//g;s/ip=//g'`
  HOST_USER=`sed -n "$host_lable_start,/user/p" ${MODIFY_CONFIG_PATH} | tail -n 1 | sed 's/[[:space:]]//g;s/user=//g'`
}

function prepareConfig() {
  flag=true
  if [[ -d ${PIER_DEPLOY_PATH} ]]; then
    print_blue "pier_$CHAINTYPE configuration file already exists"
    print_blue "reinitializing would overwrite your configuration? ($REWRITE)"
    flag=$REWRITE
    if [ $REWRITE == true ]; then
      rm -r ${PIER_DEPLOY_PATH}
    fi
  fi

  if [ $flag == true ]; then
    goduck pier config \
        --appchain "${CHAINTYPE}" \
        --target "${PIER_DEPLOY_PATH}" \
        --configPath "${MODIFY_CONFIG_PATH}" \
        --upType "binary" \
        --version "${VERSION}"
  fi
}

function deploy() {
  prepareConfig
  chmod +x ${BIN_PATH}/${PLUGIN_BIN_FILE}
  cp ${BIN_PATH}/${PLUGIN_BIN_FILE} ${PIER_DEPLOY_PATH}/plugins/appchain_plugin
  print_green "==========> Prepare config successful"
  print_blue "==========> Deploying pier-$CHAINTYPE($VERSION)"
  readHostConfig

  WHO=$HOST_USER@$HOST_IP
  FULL_TARGET=$WHO:$TARGET

  print_blue "Operating at pier-$CHAINTYPE: $WHO "
  # 传bin文件
  scp ${BIN_PATH}/${BIN_FILE} $FULL_TARGET
  # 传配置文件
    scp -r ${PIER_DEPLOY_PATH} $FULL_TARGET/.pier_$CHAINTYPE

  # 解压bin
  ssh $WHO "if [[ -d ${TARGET}/.pier_$CHAINTYPE ]]; then tar xzf ${TARGET}/${BIN_FILE} -C ${TARGET}/.pier_$CHAINTYPE; else mkdir ${TARGET}/.pier_$CHAINTYPE && tar xzf ${TARGET}/${BIN_FILE} -C ${TARGET}/.pier_$CHAINTYPE; fi; "
  #scp ${BIN_PATH}/${PLUGIN_BIN_FILE} $FULL_TARGET/.pier_$CHAINTYPE/appchain_plugin
  # 配置插件
#  sleep 3
#  ssh $WHO "mv $TARGET/.pier_$CHAINTYPE/$CHAINTYPE-client $TARGET/.pier_$CHAINTYPE/plugins/appchain_plugin; chmod +x $TARGET/.pier_$CHAINTYPE/plugins/appchain_plugin"

  print_blue "Runing pier-$CHAINTYPE"
  # 启动
  ssh $WHO "export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${TARGET}/.pier_$CHAINTYPE; cd ${TARGET}/.pier_$CHAINTYPE; nohup ./pier --repo ${TARGET}/.pier_$CHAINTYPE start >/dev/null 2>&1 &"

  print_blue "Checking pier-$CHAINTYPE"
  error=false
  PIER_PID=`ssh $WHO "sleep 3 ; lsof -i:$PPROFPORT | grep LISTEN"` || error=true
  PIER_PID=${PIER_PID#* } || error=true
  PIER_PID=`echo ${PIER_PID} | awk '{print $1}'` || error=true
  echo $PIER_PID > ${PIER_DEPLOY_PATH}/pier-$CHAINTYPE.PID || error=true
  if [ ${error} == false ]; then
    print_green "Start pier-$CHAINTYPE end"
  else
    print_red "Start pier-$CHAINTYPE fail"
  fi
}

if [ "$OPT" == "up" ]; then
  deploy
else
  printHelp
  exit 1
fi