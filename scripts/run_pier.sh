set -e

source x.sh

CURRENT_PATH=$(pwd)
GODUCK_REPO_PATH=~/.goduck
PIER_CLIENT_FABRIC_VERSION=master
PIER_CLIENT_ETHEREUM_VERSION=master
SYSTEM=$(uname -s)

function printHelp() {
  print_blue "Usage:  "
  echo "  run_pier.sh <OPT>"
  echo "    <OPT> - one of 'up', 'down', 'restart'"
  echo "      - 'up' - bring up a new pier"
  echo "      - 'down' - clear a new pier"
  echo "    -t <mode> - pier type (default \"fabric\")"
  echo "    -r <pier_root> - pier repo path (default \".pier_fabric\")"
  echo "    -v <pier_version> - pier version (default \"v1.0.0\")"
  echo "    -b <bitxhub_addr> - bitxhub addr(default \"localhost:60011\")"
  echo "  run_pier.sh -h (print this message)"
}

function prepare() {
  cd "${PIER_PATH}"
  # download pier binary package and extract
  if [ ! -a "${PIER_PATH}"/pier ]; then
    if [ "${SYSTEM}" == "Linux" ]; then
      tar xf pier-linux-amd64.tar.gz
      export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${PIER_PATH}
    elif [ "${SYSTEM}" == "Darwin" ]; then
      tar xf pier-macos-x86-64.tar.gz
      install_name_tool -change @rpath/libwasmer.dylib "${PIER_PATH}"/libwasmer.dylib "${PIER_PATH}"/pier
    else
      print_red "Pier does not support the current operating system"
    fi
  fi

  if [ ! -f "${PIER_PATH}"/pier ]; then
    print_red "pier binary is not downloaded, please download pier first"
    exit 1
  fi

  if [ "$MODE" == "fabric" ]; then
    cd "${CURRENT_PATH}"
    print_blue "===> Generate fabric pier configure"
    goduck pier config \
      --mode "relay" \
      --bitxhub "localhost:60011" \
      --validators "0xe6f8c9cf6e38bd506fae93b73ee5e80cc8f73667" \
      --validators "0x8374bb1e41d4a4bb4ac465e74caa37d242825efc" \
      --validators "0x759801eab44c9a9bbc3e09cb7f1f85ac57298708" \
      --validators "0xf2d66e2c27e93ff083ee3999acb678a36bb349bb" \
      --appchain-type "fabric" \
      --appchain-IP "127.0.0.1" \
      --target "${CONFIG_PATH}"

    mkdir -p "${CONFIG_PATH}"/plugins

    # copy appchain crypto-config and modify config.yaml
    print_blue "===> Copy fabric crypto-config"
    if [ ! -d "${CURRENT_PATH}"/crypto-config ]; then
      print_red "crypto-config not found, please start fabric network first"
      exit 1
    fi
    cp -r "${CURRENT_PATH}"/crypto-config "${CONFIG_PATH}"/fabric/

    # copy plugins file to pier root
    print_blue "===> Copy fabric plugin"
    if [ ! -f "${PIER_PATH}"/fabric1.4-client.so ]; then
      print_red "pier plugin binary is not downloaded, please download pier plugin for fabric first"
      exit 1
    fi
    cp "${PIER_PATH}"/fabric1.4-client.so "${CONFIG_PATH}"/plugins/fabric-client-1.4.so

    cd "${PIER_PATH}"
    if [ ! -f fabric_rule.wasm ]; then
      print_blue "===> Downloading fabric_rule.wasm"
      wget https://github.com/meshplus/bitxhub/blob/master/scripts/quick_start/fabric_rule.wasm
    fi
  fi

  if [ "$MODE" == "ethereum" ]; then
    cd "${CURRENT_PATH}"
    print_blue "===> Generate ethereum pier configure"
    # generate config for ethereum pier
    goduck pier config \
      --mode "relay" \
      --bitxhub "localhost:60011" \
      --validators "0xe6f8c9cf6e38bd506fae93b73ee5e80cc8f73667" \
      --validators "0x8374bb1e41d4a4bb4ac465e74caa37d242825efc" \
      --validators "0x759801eab44c9a9bbc3e09cb7f1f85ac57298708" \
      --validators "0xf2d66e2c27e93ff083ee3999acb678a36bb349bb" \
      --appchain-type "ethereum" \
      --appchain-IP "127.0.0.1" \
      --target "${CONFIG_PATH}"

    mkdir -p "${CONFIG_PATH}"/plugins

    # copy plugins file to pier root
    print_blue "===> Copy ethereum plugin"
    if [ ! -f "${PIER_PATH}"/eth-client.so ]; then
      print_red "pier plugin binary is not downloaded, please download pier plugin for ethereum first"
      exit 1
    fi
    cp "${PIER_PATH}"/eth-client.so "${CONFIG_PATH}"/plugins/eth-client.so

    cd "${CONFIG_PATH}"
    if [ ! -f ethereum_rule.wasm ]; then
      print_blue "===> Downloading ethereum_rule.wasm"
      wget https://github.com/meshplus/pier-client-ethereum/blob/master/config/validating.wasm
      mv validating.wasm ethereum_rule.wasm
    fi
  fi
}

function appchain_register() {
  "${PIER_PATH}"/pier --repo "${CONFIG_PATH}" appchain register \
    --name $1 \
    --type $2 \
    --desc $3 \
    --version $4 \
    --validators "${CONFIG_PATH}"/$5
}

function rule_deploy() {
  "${PIER_PATH}"/pier --repo "${CONFIG_PATH}" rule deploy --path "${CONFIG_PATH}"/$1_rule.wasm
}

function pier_docker_up() {
  print_blue "===> Start pier of ${MODE} in ${TYPE}..."
  if [ ! "$(docker ps -q -f name=pier-${MODE})" ]; then
    if [ "$(docker ps -aq -f status=exited -f name=pier-${MODE})" ]; then
      print_red "pier-${MODE} container already exists, please clean them first"
      exit 1
    fi

    print_blue "===> Start a new pier-${MODE} container"
    if [ "$MODE" == "fabric" ]; then
      docker run -d --name pier-fabric \
        -v "${CURRENT_PATH}"/crypto-config:/root/.pier/fabric/crypto-config \
        meshplus/pier-fabric:"${VERSION}"
    elif [ "$MODE" == "ethereum" ]; then
      print_blue "===> Wait for ethereum-node container to start for seconds..."
      sleep 5
      docker run -d --name pier-ethereum meshplus/pier-ethereum:"${VERSION}"
    else
      echo "Not supported mode"
    fi
  else
    print_red "pier-${MODE} container already running, please stop them first"
    exit 1
  fi

  print_green "===> Start pier successfully!!!"
}

function pier_binary_up() {
  prepare

  print_blue "===> pier_root: "${CONFIG_PATH}", bitxhub_addr: $BITXHUB_ADDR"

  cd "${CONFIG_PATH}"

  if [ "$MODE" == "fabric" ]; then
    print_blue "===> Register pier(fabric) to bitxhub"
    appchain_register chainA fabric chainA-description 1.4.3 fabric/fabric.validators
    print_blue "===> Deploy rule in bitxhub"
    rule_deploy fabric
    cd "${CURRENT_PATH}"
    export CONFIG_PATH="${CONFIG_PATH}"/fabric
  fi

  if [ "$MODE" == "ethereum" ]; then
    print_blue "===> Register pier(ethereum) to bitxhub"
    appchain_register chainB ether chainB-description 1.9.13 ethereum/ether.validators
    print_blue "===> Deploy rule in bitxhub"
    rule_deploy ethereum
    cd "${CURRENT_PATH}"
    export CONFIG_PATH="${CONFIG_PATH}"/ether
  fi

  print_blue "===> Start pier of ${MODE} in ${TYPE}..."
  nohup "${PIER_PATH}"/pier --repo "${CONFIG_PATH}" start >/dev/null 2>&1 &
  echo $! >"${CURRENT_PATH}/pier/pier-${MODE}.pid"
  print_green "===> Start pier successfully!!!"
}

function pier_up() {
  if [ "${TYPE}" == "docker" ]; then
    pier_docker_up
  elif [ "${TYPE}" == "binary" ]; then
    pier_binary_up
  else
    echo "Not supported up type "${TYPE}" for pier"
  fi
}

function pier_down() {
  set +e

  print_blue "===> Kill $MODE pier in binary"
  cd "${CURRENT_PATH}"/pier
  if [ -a pier-$MODE.pid ]; then
    pid=$(cat pier-$MODE.pid)
    kill "$pid"
    if [ $? -eq 0 ]; then
      echo "pier-$MODE pid:$pid exit"
    else
      print_red "pier exit fail, try use kill -9 $pid"
    fi
    rm pier-$MODE.pid
  else
    echo "pier-$MODE binary is not running"
  fi

  print_blue "===> Kill $MODE pier in docker"
  if [ "$(docker ps -q -f name=pier-$MODE)" ]; then
    docker stop pier-$MODE
  else
    echo "pier-$MODE container is not running"
  fi
}

function pier_clean() {
  set +e

  pier_down

  print_blue "===> Clean $MODE pier in binary"

  if [ -d "${CONFIG_PATH}" ]; then
    echo "remove $MODE pier configure"
    rm -r "${CONFIG_PATH}"
  else
    echo "pier-$MODE configure is not existed"
  fi

  print_blue "===> Clean $MODE pier in docker"
  if [ "$(docker ps -a -q -f name=pier-$MODE)" ]; then
    docker rm pier-$MODE
  else
    echo "pier-$MODE container is not existed"
  fi
}

PIER_ROOT=.pier_fabric
BITXHUB_ADDR="localhost:60011"
MODE="fabric"
TYPE="binary"
VERSION="v1.0.0-rc1"

CONFIG_PATH="${CURRENT_PATH}"/pier/${PIER_ROOT}
PIER_PATH="${CURRENT_PATH}/bin/pier_${VERSION}"


OPT=$1
shift

while getopts "h?t:m:r:b:v:" opt; do
  case "$opt" in
  h | \?)
    printHelp
    exit 0
    ;;
  t)
    TYPE=$OPTARG
    ;;
  m)
    MODE=$OPTARG
    ;;
  r)
    PIER_ROOT=$OPTARG
    ;;
  b)
    BITXHUB_ADDR=$OPTARG
    ;;
  v)
    VERSION=$OPTARG
    ;;
  esac
done

if [ "$OPT" == "up" ]; then
  pier_up
elif [ "$OPT" == "down" ]; then
  pier_down
elif [ "$OPT" == "clean" ]; then
  pier_clean
else
  printHelp
  exit 1
fi
