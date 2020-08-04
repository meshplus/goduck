set -e

source x.sh

CURRENT_PATH=$(pwd)
GODUCK_REPO_PATH=~/.goduck
PIER_CLIENT_FABRIC_VERSION=master
PIER_CLIENT_ETHEREUM_VERSION=master
SYSTEM=$(uname -s)

RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m'

function print_blue() {
  printf "${BLUE}%s${NC}\n" "$1"
}

function print_red() {
  printf "${RED}%s${NC}\n" "$1"
}

function printHelp() {
  print_blue "Usage:  "
  echo "  run_pier.sh <OPT>"
  echo "    <OPT> - one of 'up', 'down', 'restart'"
  echo "      - 'up' - bring up a new pier"
  echo "      - 'down' - clear a new pier"
  echo "    -t <mode> - pier type (default \"fabric\")"
  echo "    -r <pier_root> - pier repo path (default \".pier_fabric\")"
  echo "    -b <bitxhub_addr> - bitxhub addr(default \"localhost:60011\")"
  echo "  run_pier.sh -h (print this message)"
}

function prepare() {
  # download pier binary package and extract
  if [ ! -a bin/bitxhub ]; then
    mkdir -p bin && cd bin
    if [ "${SYSTEM}" == "Linux" ]; then
      tar xf pier-linux-amd64.tar.gz
      export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${CURRENT_PATH}/bin/
    elif [ "${SYSTEM}" == "Darwin" ]; then
      tar xf pier-macos-x86-64.tar.gz
      install_name_tool -change @rpath/libwasmer.dylib "${CURRENT_PATH}"/bin/libwasmer.dylib "${CURRENT_PATH}"/bin/pier
    else
      print_red "Pier does not support the current operating system"
    fi
  fi

  cd "${CURRENT_PATH}"/bin
  if [ ! -f pier ]; then
    print_red "pier binary is not downloaded, please download pier first"
    exit 1
  fi

  # judge whether to clean old storage of pier
  if [ "$CLEAN_DATA" == "true" ]; then
    if [ -d "${CURRENT_PATH}"/"${PIER_ROOT}"/store ]; then
      print_blue "===> remove old storage in "${PIER_ROOT}"/store"
      rm -rf "${CURRENT_PATH}"/"${PIER_ROOT}"/store
    fi
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
      --target "${PIER_ROOT}"

    mkdir -p "${PIER_ROOT}"/plugins

    # copy appchain crypto-config and modify config.yaml
    print_blue "===> Copy fabric crypto-config"
    if [ ! -d "${CURRENT_PATH}"/crypto-config ]; then
        print_red "crypto-config not found, please start fabric network first"
        exit 1
    fi
    cp -r "${CURRENT_PATH}"/crypto-config "${PIER_ROOT}"/fabric/

    # copy plugins file to pier root
    print_blue "===> Copy fabric plugin"
    if [ ! -f "${CURRENT_PATH}"/bin/fabric1.4-client.so ]; then
      print_red "pier plugin binary is not downloaded, please download pier plugin for fabric first"
      exit 1
    fi
    cp "${CURRENT_PATH}"/bin/fabric1.4-client.so "${PIER_ROOT}"/plugins/fabric-client-1.4.so

    cd "${CURRENT_PATH}"/"${PIER_ROOT}"
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
      --target "${PIER_ROOT}"

    mkdir -p "${PIER_ROOT}"/plugins

    # copy plugins file to pier root
    print_blue "===> Copy ethereum plugin"
    if [ ! -f "${CURRENT_PATH}"/bin/fabric1.4-client.so ]; then
      print_red "pier plugin binary is not downloaded, please download pier plugin for ethereum first"
      exit 1
    fi
    cp "${CURRENT_PATH}"/bin/eth-client.so "${PIER_ROOT}"/plugins/eth-client.so

    cd "${CURRENT_PATH}"/"${PIER_ROOT}"
    if [ ! -f ethereum_rule.wasm ]; then
        print_blue "===> Downloading ethereum_rule.wasm"
        wget https://github.com/meshplus/pier-client-ethereum/blob/master/config/validating.wasm
        mv validating.wasm ethereum_rule.wasm
    fi
  fi
}

function appchain_register(){
    pier --repo "${CURRENT_PATH}"/"${PIER_ROOT}" appchain register \
    --name $1 \
    --type $2 \
    --desc $3 \
    --version $4 \
    --validators "${CURRENT_PATH}"/"${PIER_ROOT}/$5"
}

function rule_deploy() {
    print_blue "===> deploy path: ${CURRENT_PATH}/"${PIER_ROOT}"/$1_rule.wasm"
    pier --repo "${CURRENT_PATH}"/"${PIER_ROOT}" rule deploy --path "${CURRENT_PATH}/"${PIER_ROOT}"/$1_rule.wasm"
}

function pier_docker_up() {
  print_blue "===> Start pier of ${MODE} in ${TYPE}..."
  if [ ! "$(docker ps -q -f name=pier-${MODE})" ]; then
    if [ "$(docker ps -aq -f status=exited -f name=pier-${MODE})" ]; then
        # restart your container
        print_red "===> Remove old pier-${MODE} container"
        docker rm -f pier-${MODE}
    fi

    print_blue "===> Start a new pier-${MODE} container"
    if [ "$MODE" == "fabric" ]; then
      docker run -d --name pier-fabric \
      -v "${CURRENT_PATH}"/crypto-config:/root/.pier/fabric/crypto-config \
      meshplus/pier-fabric:1.0.0-rc1
    elif [ "$MODE" == "ethereum" ]; then
      print_blue "===> Wait for ethereum-node container to start for seconds..."
      sleep 5
      docker run -d --name pier-ethereum meshplus/pier-ethereum:1.0.0-rc1
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

  print_blue "===> pier_root: $PIER_ROOT, bitxhub_addr: $BITXHUB_ADDR"

  cd "${CURRENT_PATH}"/"${PIER_ROOT}"

  if [ "$MODE" == "fabric" ]; then
    print_blue "===> Register pier(fabric) to bitxhub"
    appchain_register chainA fabric chainA-description 1.4.3 fabric/fabric.validators
    print_blue "===> Deploy rule in bitxhub"
    rule_deploy fabric
    cd "${CURRENT_PATH}"
    export CONFIG_PATH="${PIER_ROOT}"/fabric
  fi

  if [ "$MODE" == "ethereum" ]; then
    print_blue "===> Register pier(ethereum) to bitxhub"
    appchain_register chainB ether chainB-description 1.9.13 ethereum/ether.validators
    print_blue "===> Deploy rule in bitxhub"
    rule_deploy ethereum
    cd "${CURRENT_PATH}"
    export CONFIG_PATH="${PIER_ROOT}"/ether
  fi
  
  print_blue "===> Start pier of ${MODE} in ${TYPE}..."
  nohup "${CURRENT_PATH}"/bin/pier --repo "${CURRENT_PATH}"/"${PIER_ROOT}" start >/dev/null 2>&1 &
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
  fi

  print_blue "===> Kill $MODE pier in docker"
  if [ "$(docker ps -q -f name=pier-$MODE)" ]; then
    docker rm -f pier-$MODE
    exit 0
  fi
  echo "pier-$MODE container is not running"
}

function pier_id() {
  if [ "${TYPE}" == "docker" ]; then
    print_blue "===> pier id of pier-$MODE in docker is as follow"
    docker exec pier-"$MODE" pier --repo=/root/.pier id
  elif [ "${TYPE}" == "binary" ]; then
    print_blue "===> pier id of $MODE is as follow"
    "${CURRENT_PATH}"/bin/pier --repo "${CURRENT_PATH}"/"${PIER_ROOT}"
  else
   echo "Not supported up type "${TYPE}" for pier"
  fi
}

PIER_ROOT=.pier_fabric
BITXHUB_ADDR="localhost:60011"
CLEAN_DATA="true"
MODE="fabric"
TYPE="binary"

OPT=$1
shift

while getopts "h?t:m:r:b:c:" opt; do
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
  c)
    CLEAN_DATA=$OPTARG
    ;;
  esac
done

if [ "$OPT" == "up" ]; then
  pier_up
elif [ "$OPT" == "down" ]; then
  pier_down
elif [ "$OPT" == "id" ]; then
  pier_id
else
  printHelp
  exit 1
fi
