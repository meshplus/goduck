#!/usr/bin/env bash

set -e
source x.sh

CURRENT_PATH=$(pwd)
CONFIG_PATH="${CURRENT_PATH}"/bitxhub

OPT=$1
VERSION=$2
TYPE=$3
MODE=$4
N=$5
SYSTEM=$(uname -s)
if [ $SYSTEM == "Linux" ]; then
  SYSTEM="linux"
elif [ $SYSTEM == "Darwin" ]; then
  SYSTEM="darwin"
fi 
BXH_PATH="${CURRENT_PATH}/bin/bitxhub_${SYSTEM}_${VERSION}"

function printHelp() {
  print_blue "Usage:  "
  echo "  playground.sh <mode>"
  echo "    <OPT> - one of 'up', 'down', 'restart'"
  echo "      - 'up' - bring up the bitxhub network"
  echo "      - 'down' - clear the bitxhub network"
  echo "      - 'restart' - restart the bitxhub network"
  echo "  playground.sh -h (print this message)"
}

function binary_prepare() {
  cd "${BXH_PATH}"
  if [ ! -a "${BXH_PATH}"/bitxhub ]; then
    if [ "${SYSTEM}" == "linux" ]; then
      tar xf bitxhub_linux-amd64_$VERSION.tar.gz
      cp ./build/* . && rm -r build
      export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:"${BXH_PATH}"/
    elif [ "${SYSTEM}" == "darwin" ]; then
      tar xf bitxhub_macos_x86_64_$VERSION.tar.gz
      cp ./build/* . && rm -r build
      install_name_tool -change @rpath/libwasmer.dylib "${BXH_PATH}"/libwasmer.dylib "${BXH_PATH}"/bitxhub
    else
      print_red "Bitxhub does not support the current operating system"
    fi
  fi

  if [ -a "${CONFIG_PATH}"/bitxhub.pid ]; then
    print_red "Bitxhub already run in daemon processes"
    exit 1
  fi
}

function bitxhub_binary_check() {
  if [ -n "$(ps -p ${PID} -o pid=)" ]; then
    if [ -n "$(tail -n 50 ${NODEPATH}/logs/bitxhub.log | grep "Order is ready")" ]; then
      checkRet=1
    else
      checkRet=0
    fi;
  else
    checkRet=0
  fi
}

function bitxhub_docker_check() {
  if [ "$(docker container ls | grep -c ${CONTAINER})" -ge 1 ]; then
    if [ -n "$(docker logs --tail 50 ${CONTAINER} | grep "Order is ready")" ]; then
      checkRet=1
    else
      checkRet=0
    fi;
  else
    checkRet=0
  fi
}

function bitxhub_binary_solo() {
  binary_prepare

  cd "${CONFIG_PATH}"
  if [ ! -d nodeSolo/plugins ]; then
    mkdir nodeSolo/plugins
    cp -r "${BXH_PATH}"/solo.so nodeSolo/plugins
  fi
  print_blue "Start bitxhub solo by binary"
  nohup "${BXH_PATH}"/bitxhub --repo "${CONFIG_PATH}"/nodeSolo start >/dev/null 2>&1 &
  PID=$!
  NODEPATH="${CONFIG_PATH}"/nodeSolo
  sleep 3
  bitxhub_binary_check
  if [ ${checkRet} == "1" ]; then
    print_green "===> Start bitxhub solo successful"
    echo ${VERSION} >>"${CONFIG_PATH}"/bitxhub.version
    echo ${PID} >>"${CONFIG_PATH}"/bitxhub.pid
  else
    print_red "===> Start bitxhub solo fail"
    echo ${VERSION} >>"${CONFIG_PATH}"/bitxhub.version
    echo ${PID} >>"${CONFIG_PATH}"/bitxhub.pid
  fi
  print_blue "The above result check information may not be correct, you can use the \"goduck status list\" command to check the status of the startup BitXHub node."
}

function bitxhub_docker_solo() {
  VERSION=${VERSION:1}
  if [[ -z "$(docker images -q meshplus/bitxhub-solo:$VERSION 2>/dev/null)" ]]; then
    docker pull meshplus/bitxhub-solo:$VERSION
  fi

  print_blue "Start bitxhub solo mode by docker"
  if [ "$(docker container ls -a | grep -c bitxhub_solo)" -ge 1 ]; then
    docker start bitxhub_solo
  else
    docker run -d --name bitxhub_solo \
      -p 60011:60011 -p 9091:9091 -p 53121:53121 -p 40011:40011 \
      -v "${CONFIG_PATH}"/nodeSolo/api:/root/.bitxhub/api \
      -v "${CONFIG_PATH}"/nodeSolo/bitxhub.toml:/root/.bitxhub/bitxhub.toml \
      -v "${CONFIG_PATH}"/nodeSolo/genesis.json:/root/.bitxhub/genesis.json \
      -v "${CONFIG_PATH}"/nodeSolo/network.toml:/root/.bitxhub/network.toml \
      -v "${CONFIG_PATH}"/nodeSolo/order.toml:/root/.bitxhub/order.toml \
      -v "${CONFIG_PATH}"/nodeSolo/certs:/root/.bitxhub/certs \
      meshplus/bitxhub-solo:$VERSION
  fi

  CONTAINER=bitxhub_solo
  sleep 3
  bitxhub_docker_check
  if [ ${checkRet} == "1" ]; then
    print_green "===> Start bitxhub solo successful"
    echo v${VERSION} >>"${CONFIG_PATH}"/bitxhub.version
    CID=`docker container ls | grep bitxhub_solo`
    echo ${CID:0:12} >>"${CONFIG_PATH}"/bitxhub.cid
  else
    print_red "===> Start bitxhub solo fail"
    echo v${VERSION} >>"${CONFIG_PATH}"/bitxhub.version
    CID=`docker container ls | grep bitxhub_solo`
    echo ${CID:0:12} >>"${CONFIG_PATH}"/bitxhub.cid
  fi
  print_blue "The above result check information may not be correct, you can use the \"goduck status list\" command to check the status of the startup BitXHub node."
}

function bitxhub_binary_cluster() {
  binary_prepare
  declare -a PIDS
  cd "${CONFIG_PATH}"
  print_blue "Start bitxhub cluster"

  for ((i = 1; i < N + 1; i = i + 1)); do
    if [ ! -d node${i}/plugins ]; then
      mkdir node${i}/plugins
      cp -r "${BXH_PATH}"/raft.so node${i}/plugins
    fi
    echo "Start bitxhub node${i}"
    nohup "${BXH_PATH}"/bitxhub --repo="${CONFIG_PATH}"/node${i} start >/dev/null 2>&1 &
    PIDS[${i}]=$!
  done

  sleep 4
  for ((i = 1; i < N + 1; i = i + 1)); do
    NODEPATH="${CONFIG_PATH}"/node${i}
    PID=${PIDS[${i}]}
    bitxhub_binary_check
    if [ ${checkRet} == "1" ]; then
      print_green "===> Start bitxhub node${i} successful"
      echo ${VERSION} >>"${CONFIG_PATH}"/bitxhub.version
      echo ${PID} >>"${CONFIG_PATH}"/bitxhub.pid
    else
      print_red "===> Start bitxhub node${i} fail"
      echo ${VERSION} >>"${CONFIG_PATH}"/bitxhub.version
      echo ${PID} >>"${CONFIG_PATH}"/bitxhub.pid
    fi
  done
  print_blue "The above result check information may not be correct, you can use the \"goduck status list\" command to check the status of the startup BitXHub node."
}

function bitxhub_docker_cluster() {
  VERSION=${VERSION:1}
  if [[ -z "$(docker images -q meshplus/bitxhub:$VERSION 2>/dev/null)" ]]; then
    docker pull meshplus/bitxhub:$VERSION
  fi
  print_blue "Start bitxhub cluster mode by docker compose"
  x_replace "s/bitxhub:latest/bitxhub:$VERSION/g" "${CURRENT_PATH}"/docker/docker-compose.yml
  docker-compose -f "${CURRENT_PATH}"/docker/docker-compose.yml up -d

  sleep 4
  for ((i = 1; i < N + 1; i = i + 1)); do
    CONTAINER=bitxhub_node${i}
    bitxhub_docker_check
    if [ ${checkRet} == "1" ]; then
      print_green "===> Start bitxhub node${i} successful"
      echo v${VERSION} >>"${CONFIG_PATH}"/bitxhub.version
      CID=`docker container ls | grep bitxhub_node$i`
      echo ${CID:0:12} >> "${CONFIG_PATH}"/bitxhub.cid
    else
      print_red "===> Start bitxhub node${i} fail"
      echo v${VERSION} >>"${CONFIG_PATH}"/bitxhub.version
      CID=`docker container ls | grep bitxhub_node$i`
      echo ${CID:0:12} >> "${CONFIG_PATH}"/bitxhub.cid
    fi
  done
  print_blue "The above result check information may not be correct, you can use the \"goduck status list\" command to check the status of the startup BitXHub node."
}

function bitxhub_down() {
  set +e
  print_blue "===> Stop bitxhub"

  if [ -a "${CONFIG_PATH}"/bitxhub.pid ]; then
    list=$(cat "${CONFIG_PATH}"/bitxhub.pid)
    for pid in $list; do
      kill "$pid"
      if [ $? -eq 0 ]; then
        echo "node pid:$pid exit"
      else
        print_red "program exit fail, try use kill -9 $pid"
      fi
    done
    rm "${CONFIG_PATH}"/bitxhub.pid
  fi

 if [ -a "${CONFIG_PATH}"/bitxhub.cid ]; then
    list=$(cat "${CONFIG_PATH}"/bitxhub.cid)
    for cid in $list; do
      docker kill "$cid"
      if [ $? -eq 0 ]; then
        echo "container id:$cid exit"
      else
        print_red "container exit fail"
      fi
    done
    rm "${CONFIG_PATH}"/bitxhub.cid
  fi
}

function bitxhub_up() {
  case $MODE in
  "docker")
    case $TYPE in
    "solo")
      bitxhub_docker_solo
      ;;
    "cluster")
      bitxhub_docker_cluster
      ;;
    *)
      print_red "TYPE should be solo or cluster"
      exit 1
      ;;
    esac
    ;;
  "binary")
    case $TYPE in
    "solo")
      bitxhub_binary_solo
      ;;
    "cluster")
      bitxhub_binary_cluster
      ;;
    *)
      print_red "TYPE should be solo or cluster"
      exit 1
      ;;
    esac
    ;;
  *)
    print_red "MODE should be docker or binary"
    exit 1
    ;;
  esac
}

function bitxhub_clean() {
  set +e

  bitxhub_down

  print_blue "===> Clean bitxhub"

  file_list=$(ls ${CONFIG_PATH} 2>/dev/null | grep -v '^$')
  for file_name in $file_list; do
    if [ "${file_name:0:4}" == "node" ]; then
      rm -r "${CONFIG_PATH}"/"$file_name"
      echo "remove bitxhub configure $file_name"
    fi
  done

  if [ "$(docker ps -a | grep -c bitxhub_node)" -ge 1 ]; then
    docker-compose -f "${CURRENT_PATH}"/docker/docker-compose.yml rm -f
    echo "bitxhub docker cluster clean"
  fi

  if [ "$(docker ps -a | grep -c bitxhub_solo)" -ge 1 ]; then
    docker rm bitxhub_solo
    echo "bitxhub docker solo clean"
  fi

  if [ -e "${CONFIG_PATH}"/bitxhub.pid ]; then
    rm "${CONFIG_PATH}"/bitxhub.pid
  fi
  if [ -e "${CONFIG_PATH}"/bitxhub.cid ]; then
    rm "${CONFIG_PATH}"/bitxhub.cid
  fi
  if [ -e "${CONFIG_PATH}"/bitxhub.version ]; then
    rm "${CONFIG_PATH}"/bitxhub.version
  fi
}

function bitxhub_restart() {
  bitxhub_down
  bitxhub_up
}

if [ "$OPT" == "up" ]; then
  bitxhub_up
elif [ "$OPT" == "down" ]; then
  bitxhub_down
elif [ "$OPT" == "clean" ]; then
  bitxhub_clean
elif [ "$OPT" == "restart" ]; then
  bitxhub_restart
else
  printHelp
  exit 1
fi
