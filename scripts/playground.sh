#!/usr/bin/env bash

set -e
source x.sh
source compare.sh

CURRENT_PATH=$(pwd)
OPT=$1
VERSION=$2
TYPE=$3
MODIFY_CONFIG_PATH=$4
TARGET=$5
MODE=$(sed '/^.*mode/!d;s/.*=//;s/[[:space:]]//g' ${MODIFY_CONFIG_PATH})
NUM=$(sed '/^.*num/!d;s/.*=//;s/[[:space:]]//g' ${MODIFY_CONFIG_PATH})
REWRITE=$(sed '/^.*rewrite/!d;s/.*=//;s/[[:space:]]//g' ${MODIFY_CONFIG_PATH})
CONFIG_PATH="${CURRENT_PATH}"/bitxhub

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

function check_bitxhub() {
  if [ -a "${CONFIG_PATH}"/bitxhub.pid ]; then
    print_red "Bitxhub already run in daemon processes"
    exit 1
  fi

  if [ -a "${CONFIG_PATH}"/bitxhub.cid ]; then
    print_red "Bitxhub already run in daemon processes"
    exit 1
  fi

  flag=true
  if [[ -d ${TARGET}/nodeSolo && $MODE == "solo" ]]; then
    print_blue "BitXHub solo configuration file already exists"
    print_blue "reinitializing would overwrite your configuration? ($REWRITE)"
    flag=$REWRITE
    if [ $REWRITE == true ]; then
      rm -r ${TARGET}
    fi
  fi
  if [[ -d ${TARGET}/node1 && $MODE == "cluster" ]]; then
    print_blue "BitXHub cluster configuration file already exists"
    print_blue "reinitializing would overwrite your configuration? ($REWRITE)"
    flag=$REWRITE
    if [ $REWRITE == true ]; then
      rm -r ${TARGET}
    fi
  fi

  if [ $flag == true ]; then
    goduck bitxhub config --version $VERSION --target "${TARGET}" --configPath "${MODIFY_CONFIG_PATH}"
  fi
}

function binary_prepare() {
  cd "${BXH_PATH}"
  if [ ! -a "${BXH_PATH}"/bitxhub ]; then
    if [ "${SYSTEM}" == "linux" ]; then
      tar xf bitxhub_linux-amd64_$VERSION.tar.gz
      export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:"${BXH_PATH}"/
    elif [ "${SYSTEM}" == "darwin" ]; then
      tar xf bitxhub_darwin_x86_64_$VERSION.tar.gz
      install_name_tool -change @rpath/libwasmer.dylib "${BXH_PATH}"/libwasmer.dylib "${BXH_PATH}"/bitxhub
    else
      print_red "Bitxhub does not support the current operating system"
    fi
  fi

  check_bitxhub
}

function bitxhub_binary_solo() {
  binary_prepare
  cd "${TARGET}"
  version1=$VERSION
  version2=v1.11.0
  version_compare
  if [[ $versionComPareRes -lt 0 ]]; then
    if [ ! -d nodeSolo/plugins ]; then
      mkdir "${TARGET}"/nodeSolo/plugins
      cp -r "${BXH_PATH}"/solo.so nodeSolo/plugins
    fi
  fi
  print_blue "======> Start bitxhub solo by binary"
  nohup "${BXH_PATH}"/bitxhub --repo "${TARGET}"/nodeSolo start >/dev/null 2>&1 &
  PID=$!
  echo ${VERSION} >>"${CONFIG_PATH}"/bitxhub.version
  "${BXH_PATH}"/bitxhub version >"${CONFIG_PATH}"/bitxhub-binary.version
  echo ${PID} >>"${CONFIG_PATH}"/bitxhub.pid

  print_blue "You can use the \"goduck status list\" command to check the status of the startup BitXHub node."
  if [ ${VERSION} == "v1.8.0" ]; then
    print_blue "Note: To register the appchain, you need to execute the \"bitxhub client did init\" command."
  fi
}

function bitxhub_docker_solo() {
  docker_prepare
  print_blue "======> Start bitxhub solo mode by docker compose"
  docker-compose -f "${CONFIG_PATH}"/docker-compose-bitxhub-solo.yaml up -d

  echo v${VERSION} >>"${CONFIG_PATH}"/bitxhub.version
  sleep 1
  CID=$(docker container ls | grep bitxhub-solo)
  echo ${CID:0:12} >>"${CONFIG_PATH}"/bitxhub.cid
  docker exec ${CID:0:12} bitxhub version >"${CONFIG_PATH}"/bitxhub-docker.version
  print_blue "You can use the \"goduck status list\" command to check the status of the startup BitXHub node."
  if [ ${VERSION} == "v1.8.0" ]; then
    print_blue "Note: To register the appchain, you need to execute the \"bitxhub client did init\" command."
  fi
}

function bitxhub_binary_cluster() {
  binary_prepare
  declare -a PIDS
  cd "${TARGET}"
  print_blue "======> Start bitxhub cluster"
  for ((i = 1; i < $NUM + 1; i = i + 1)); do
    version1=$VERSION
    version2=v1.11.0
    version_compare
    if [[ $versionComPareRes -lt 0 ]]; then
      if [ ! -d node${i}/plugins ]; then
        mkdir node${i}/plugins
        cp -r "${BXH_PATH}"/raft.so node${i}/plugins
      fi
    fi
    echo "Start bitxhub node${i}"
    nohup "${BXH_PATH}"/bitxhub --repo="${TARGET}"/node${i} start >/dev/null 2>&1 &
    PIDS[${i}]=$!
  done

  for ((i = 1; i < $NUM + 1; i = i + 1)); do
    PID=${PIDS[${i}]}
    echo ${VERSION} >>"${CONFIG_PATH}"/bitxhub.version
    echo ${PID} >>"${CONFIG_PATH}"/bitxhub.pid
  done
  "${BXH_PATH}"/bitxhub version >"${CONFIG_PATH}"/bitxhub-binary.version
  print_blue "You can use the \"goduck status list\" command to check the status of the startup BitXHub node."
  if [ ${VERSION} == "v1.8.0" ]; then
    print_blue "Note: To register the appchain, you need to execute the \"bitxhub client did init\" command."
  fi
}

function docker_prepare() {
  check_bitxhub

  if [ $flag == true ]; then

    for ((i = 1; i <= $NUM; i++)); do
      nodeName="node$i"
      if [ $MODE == "solo" ]; then
        nodeName="nodeSolo"
      fi
      for ((j = 1; j <= $NUM; j++)); do
        j_tmp=$(expr $j + 1)
        x_replace "s/\"\/ip4\/127.0.0.1\/tcp\/400$j\/p2p\/\"/\"\/ip4\/172.19.0.$j_tmp\/tcp\/400$j\/p2p\/\"/g" "${TARGET}"/$nodeName/network.toml
      done
    done

    DOCKER_COMPOSE_FILE=docker-compose-bitxhub.yaml
    if [ $MODE == "cluster" ]; then
      if [[ -z "$(docker images -q meshplus/bitxhub:$VERSION 2>/dev/null)" ]]; then
        docker pull meshplus/bitxhub:$VERSION
      fi

      cp "${CURRENT_PATH}"/docker/bitxhub/docker-compose-bitxhub.yaml "${CONFIG_PATH}"
      x_replace "s/image: meshplus\/bitxhub:.*/image: meshplus\/bitxhub:$VERSION/g" "${CONFIG_PATH}"/"${DOCKER_COMPOSE_FILE}"
    else
      if [[ -z "$(docker images -q meshplus/bitxhub-solo:$VERSION 2>/dev/null)" ]]; then
        docker pull meshplus/bitxhub-solo:$VERSION
      fi

      DOCKER_COMPOSE_FILE=docker-compose-bitxhub-solo.yaml
      cp "${CURRENT_PATH}"/docker/bitxhub/docker-compose-bitxhub-solo.yaml "${CONFIG_PATH}"
      x_replace "s/image: meshplus\/bitxhub-solo:.*/image: meshplus\/bitxhub-solo:$VERSION/g" "${CONFIG_PATH}"/"${DOCKER_COMPOSE_FILE}"
    fi

    bitxhubRepoTmp=$(echo "${TARGET}" | sed 's/\//\\\//g')

    # read port
    JSONRPCP=$(sed '/^.*jsonrpc_port/!d;s/.*=//;s/[[:space:]]//g' ${MODIFY_CONFIG_PATH})
    GRPCP=$(sed '/^.*grpc_port/!d;s/.*=//;s/[[:space:]]//g' ${MODIFY_CONFIG_PATH})
    GATEWAYP=$(sed '/^.*gateway_port/!d;s/.*=//;s/[[:space:]]//g' ${MODIFY_CONFIG_PATH})
    PPROFP=$(sed '/^.*pprof_port/!d;s/.*=//;s/[[:space:]]//g' ${MODIFY_CONFIG_PATH})
    MONITORP=$(sed '/^.*monitor_port/!d;s/.*=//;s/[[:space:]]//g' ${MODIFY_CONFIG_PATH})
    NUM=$(sed '/^.*num/!d;s/.*=//;s/[[:space:]]//g' ${MODIFY_CONFIG_PATH})
    for a in $JSONRPCP; do
      JSONRPCPS+=($a)
    done
    for a in $GRPCP; do
      GRPCPS+=($a)
    done
    for a in $GATEWAYP; do
      GATEWAYPS+=($a)
    done
    for a in $PPROFP; do
      PPROFPS+=($a)
    done
    for a in $MONITORP; do
      MONITORPS+=($a)
    done

    if [ $NUM -gt 4 ]; then
      for ((i = 5; i <= $NUM; i++)); do
        jsonrpcP=${JSONRPCPS[$i - 1]}
        grpcP=${GRPCPS[$i - 1]}
        gatewayP=${GATEWAYPS[$i - 1]}
        pprofP=${PPROFPS[$i - 1]}
        monitorP=${MONITORPS[$i - 1]}
        i_tmp=$(expr $i + 1)
        version1=${VERSION}
        version2="v1.8.0"
        version_compare
        if [[ $versionComPareRes -lt 0 ]]; then
#        if [ "${VERSION}" \< "v1.8.0" ]; then
          echo "
bitxhub_node$i:
    restart: always
          image: meshplus/bitxhub:$VERSION
    container_name: bitxhub-0$i
    tty: true
    volumes:
      - /var/run/:/host/var/run/
      - ${bitxhubRepoTmp}\/node$i/bitxhub.toml:/root/.bitxhub/bitxhub.toml
      - ${bitxhubRepoTmp}\/node$i/network.toml:/root/.bitxhub/network.toml
      - ${bitxhubRepoTmp}\/node$i/key.json:/root/.bitxhub/key.json
      - ${bitxhubRepoTmp}\/node$i/certs:/root/.bitxhub/certs
    ports:
      - \"$grpcP:$grpcP\"
      - \"$gatewayP:$gatewayP\"
      - \"$pprofP:$pprofP\"
      - \"$monitorP:$monitorP\"
      - \"400$i:400$i\"
    working_dir: /root/.bitxhub
    networks:
      p2p:
        ipv4_address: 172.19.0.$i_tmp" >>"${TARGET}"/"${DOCKER_COMPOSE_FILE}"
        else
          echo "
bitxhub_node$i:
    restart: always
          image: meshplus/bitxhub:$VERSION
    container_name: bitxhub-0$i
    tty: true
    volumes:
      - /var/run/:/host/var/run/
      - ${bitxhubRepoTmp}\/node$i/bitxhub.toml:/root/.bitxhub/bitxhub.toml
      - ${bitxhubRepoTmp}\/node$i/network.toml:/root/.bitxhub/network.toml
      - ${bitxhubRepoTmp}\/node$i/key.json:/root/.bitxhub/key.json
      - ${bitxhubRepoTmp}\/node$i/certs:/root/.bitxhub/certs
    ports:
      - \"$grpcP:$grpcP\"
      - \"$gatewayP:$gatewayP\"
      - \"$pprofP:$pprofP\"
      - \"$monitorP:$monitorP\"
      - \"400$i:400$i\"
    working_dir: /root/.bitxhub
    networks:
      p2p:
        ipv4_address: 172.19.0.$i_tmp" >>"${TARGET}"/"${DOCKER_COMPOSE_FILE}"
        fi
      done
    fi

    for ((i = 1; i <= $NUM; i++)); do
      nodeName="node$i"
      if [ $MODE == "solo" ]; then
        nodeName="nodeSolo"
      fi
      x_replace "s/- .\/bitxhub\/.bitxhub\/$nodeName\/bitxhub.toml:\/root\/.bitxhub\/bitxhub.toml/- ${bitxhubRepoTmp}\/$nodeName\/bitxhub.toml:\/root\/.bitxhub\/bitxhub.toml/g" "${CONFIG_PATH}"/"${DOCKER_COMPOSE_FILE}"
      x_replace "s/- .\/bitxhub\/.bitxhub\/$nodeName\/network.toml:\/root\/.bitxhub\/network.toml/- ${bitxhubRepoTmp}\/$nodeName\/network.toml:\/root\/.bitxhub\/network.toml/g" "${CONFIG_PATH}"/"${DOCKER_COMPOSE_FILE}"
      x_replace "s/- .\/bitxhub\/.bitxhub\/$nodeName\/order.toml:\/root\/.bitxhub\/order.toml/- ${bitxhubRepoTmp}\/$nodeName\/order.toml:\/root\/.bitxhub\/order.toml/g" "${CONFIG_PATH}"/"${DOCKER_COMPOSE_FILE}"
      x_replace "s/- .\/bitxhub\/.bitxhub\/$nodeName\/key.json:\/root\/.bitxhub\/key.json/- ${bitxhubRepoTmp}\/$nodeName\/key.json:\/root\/.bitxhub\/key.json/g" "${CONFIG_PATH}"/"${DOCKER_COMPOSE_FILE}"
      x_replace "s/- .\/bitxhub\/.bitxhub\/$nodeName\/certs:\/root\/.bitxhub\/certs/- ${bitxhubRepoTmp}\/$nodeName\/certs:\/root\/.bitxhub\/certs/g" "${CONFIG_PATH}"/"${DOCKER_COMPOSE_FILE}"

      jsonrpcP=${JSONRPCPS[$i - 1]}
      grpcP=${GRPCPS[$i - 1]}
      gatewayP=${GATEWAYPS[$i - 1]}
      pprofP=${PPROFPS[$i - 1]}
      monitorP=${MONITORPS[$i - 1]}

      if [ "${VERSION}" != "v1.6.1" ] && [ "${VERSION}" != "v1.6.2" ]; then
        x_replace "s/\".*:788$i\"/\"$jsonrpcP:$jsonrpcP\"/" "${CONFIG_PATH}"/"${DOCKER_COMPOSE_FILE}"
      fi

      x_replace "s/\".*:5001$i\"/\"$grpcP:$grpcP\"/" "${CONFIG_PATH}"/"${DOCKER_COMPOSE_FILE}"
      x_replace "s/\".*:809$i\"/\"$gatewayP:$gatewayP\"/" "${CONFIG_PATH}"/"${DOCKER_COMPOSE_FILE}"
      x_replace "s/\".*:4312$i\"/\"$pprofP:$pprofP\"/" "${CONFIG_PATH}"/"${DOCKER_COMPOSE_FILE}"
      x_replace "s/\".*:3001$i\"/\"$monitorP:$monitorP\"/" "${CONFIG_PATH}"/"${DOCKER_COMPOSE_FILE}"

      x_replace "s/ipv4_address: 172.19.0.$i/ipv4_address: 172.19.0.$i/" "${CONFIG_PATH}"/"${DOCKER_COMPOSE_FILE}"
    done

  fi
}

function bitxhub_docker_cluster() {
  docker_prepare
  print_blue "======> Start bitxhub cluster mode by docker compose"
  docker-compose -f "${CONFIG_PATH}"/docker-compose-bitxhub.yaml up -d

  for ((i = 1; i < $NUM + 1; i = i + 1)); do
    echo v${VERSION} >>"${CONFIG_PATH}"/bitxhub.version
    sleep 1
    CID=$(docker container ls | grep bitxhub_node$i)
    echo ${CID:0:12} >>"${CONFIG_PATH}"/bitxhub.cid
    docker exec ${CID:0:12} bitxhub version >"${CONFIG_PATH}"/bitxhub-docker.version
  done

  print_blue "You can use the \"goduck status list\" command to check the status of the startup BitXHub node."
  if [ ${VERSION} == "v1.8.0" ]; then
    print_blue "Note: To register the appchain, you need to execute the \"bitxhub client did init\" command."
  fi
}

function bitxhub_down() {
  set +e
  print_blue "======> Stop bitxhub"
  BITXHUB_CONFIG_PATH="${CURRENT_PATH}"/bitxhub
  if [ -e "${BITXHUB_CONFIG_PATH}"/bitxhub.pid ]; then
    for (( i = 1; ; i++ )); do
      pid=$(cat "${BITXHUB_CONFIG_PATH}"/bitxhub.pid | sed -n $i"p")
      if [ -z $pid ]; then
        break
      fi

      kill "$pid"
      if [ $? -eq 0 ]; then
        echo "node pid:$pid exit"
      else
        print_red "program exit fail, try use kill -9 $pid"
      fi
    done
    rm "${BITXHUB_CONFIG_PATH}"/bitxhub.pid
  fi

  if [ "$(docker ps | grep -c bitxhub_node)" -ge 1 ]; then
    docker-compose -f "${CONFIG_PATH}"/docker-compose-bitxhub.yaml stop
    echo "bitxhub docker cluster stop"
  fi

  if [ "$(docker ps | grep -c bitxhub_solo)" -ge 1 ]; then
    docker-compose -f "${CONFIG_PATH}"/docker-compose-bitxhub-solo.yaml stop
    echo "bitxhub docker solo stop"
  fi

  cleanBxhInfoFile
}

function bitxhub_up() {
  case $TYPE in
  "docker")
    case $MODE in
    "solo")
      bitxhub_docker_solo
      ;;
    "cluster")
      bitxhub_docker_cluster
      ;;
    *)
      print_red "MODE should be solo or cluster"
      exit 1
      ;;
    esac
    ;;
  "binary")
    case $MODE in
    "solo")
      bitxhub_binary_solo
      ;;
    "cluster")
      bitxhub_binary_cluster
      ;;
    *)
      print_red "MODE should be solo or cluster"
      exit 1
      ;;
    esac
    ;;
  *)
    print_red "TYPE should be docker or binary"
    exit 1
    ;;
  esac
}

function bitxhub_clean() {
  set +e

  bitxhub_down

  print_blue "======> Clean bitxhub"
  if [ "$(docker ps -a | grep -c bitxhub_node)" -ge 1 ]; then
    docker-compose -f "${CONFIG_PATH}"/docker-compose-bitxhub.yaml rm -f
    echo "bitxhub docker cluster clean"
  fi

  if [ "$(docker ps -a | grep -c bitxhub_solo)" -ge 1 ]; then
    docker-compose -f "${CONFIG_PATH}"/docker-compose-bitxhub-solo.yaml rm -f
    echo "bitxhub docker solo clean"
  fi

  file_list=$(ls ${CONFIG_PATH}/.bitxhub 2>/dev/null | grep -v '^$')
  for file_name in $file_list; do
    if [ "${file_name:0:4}" == "node" ]; then
      rm -r ${CONFIG_PATH}/.bitxhub/"$file_name"
      echo "remove bitxhub configure $file_name"
    fi
  done
}

function cleanBxhInfoFile() {
  BITXHUB_CONFIG_PATH="${CURRENT_PATH}"/bitxhub
  if [ -e "${BITXHUB_CONFIG_PATH}"/bitxhub.pid ]; then
    rm "${BITXHUB_CONFIG_PATH}"/bitxhub.pid
  fi
  if [ -e "${BITXHUB_CONFIG_PATH}"/bitxhub.cid ]; then
    rm "${BITXHUB_CONFIG_PATH}"/bitxhub.cid
  fi
  if [ -e "${BITXHUB_CONFIG_PATH}"/bitxhub.version ]; then
    rm "${BITXHUB_CONFIG_PATH}"/bitxhub.version
  fi
  if [ -e "${BITXHUB_CONFIG_PATH}"/bitxhub-docker.version ]; then
    rm "${BITXHUB_CONFIG_PATH}"/bitxhub-docker.version
  fi
  if [ -e "${BITXHUB_CONFIG_PATH}"/bitxhub-binary.version ]; then
    rm "${BITXHUB_CONFIG_PATH}"/bitxhub-binary.version
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
