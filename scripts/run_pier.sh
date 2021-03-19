set -e

source x.sh

CURRENT_PATH=$(pwd)
GODUCK_REPO_PATH=~/.goduck
PIER_CLIENT_FABRIC_VERSION=master
PIER_CLIENT_ETHEREUM_VERSION=master
FABRIC_RULE=fabric_rule.wasm
ETHEREUM_RULE=ethereum_rule.wasm
SYSTEM=$(uname -s)
if [ $SYSTEM == "Linux" ]; then
  SYSTEM="linux"
elif [ $SYSTEM == "Darwin" ]; then
  SYSTEM="darwin"
fi

function printHelp() {
  print_blue "Usage:  "
  echo "  run_pier.sh <OPT>"
  echo "    <OPT> - one of 'register', 'up', 'down', 'restart'"
  echo "      - 'register' - register pier to bitxhub"
  echo "      - 'up' - bring up a new pier"
  echo "      - 'down' - clear a new pier"
  echo "    -t <mode> - pier type (default \"fabric\")"
  echo "    -r <pier_root> - pier repo path (default \".pier_fabric\")"
  echo "    -v <pier_version> - pier version (default \"v1.1.0-rc1\")"
  echo "    -b <bitxhub_addr> - bitxhub addr(default \"localhost:60011\")"
  echo "  run_pier.sh -h (print this message)"
}

function extractBin() {
    cd "${PIER_PATH}"
  # download pier binary package and extract
  if [ ! -a "${PIER_PATH}"/pier ]; then
    if [ "${SYSTEM}" == "linux" ]; then
      tar xf pier_linux-amd64_$VERSION.tar.gz -C ${PIER_PATH} --strip-components 1
      export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${PIER_PATH}
    elif [ "${SYSTEM}" == "darwin" ]; then
      tar xf pier_darwin_x86_64_$VERSION.tar.gz -C ${PIER_PATH} --strip-components 1
      install_name_tool -change @rpath/libwasmer.dylib "${PIER_PATH}"/libwasmer.dylib "${PIER_PATH}"/pier
    else
      print_red "Pier does not support the current operating system"
    fi
  fi

  if [ ! -f "${PIER_PATH}"/pier ]; then
    print_red "pier binary is not downloaded, please download pier first"
    exit 1
  fi
}

function generateConfig() {
  if [ "${TYPE}" == "docker" ]; then
    BITXHUB_ADDR="host.docker.internal:60011"
  fi

  if [ "$MODE" == "fabric" ]; then
    cd "${CURRENT_PATH}"
    print_blue "===> Generate fabric pier configure"
    overW='Y'
    if [ "$OVERWRITE" == "false" ]; then
        overW='N'
    fi
    echo $overW|goduck pier config \
      --mode "relay" \
      --type $TYPE \
      --bitxhub ${BITXHUB_ADDR} \
      --validators "0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013" \
      --validators "0x79a1215469FaB6f9c63c1816b45183AD3624bE34" \
      --validators "0x97c8B516D19edBf575D72a172Af7F418BE498C37" \
      --validators "0xc0Ff2e0b3189132D815b8eb325bE17285AC898f8" \
      --appchainType "fabric" \
      --appchainIP ${APPCHAINIP} \
      --target ${CONFIG_PATH} \
      --tls ${TLS} \
      --httpPort ${HTTP} \
      --pprofPort ${PPROF} \
      --apiPort ${API} \
      --cryptoPath ${CRYPTOPATH} \
      --version ${VERSION}

    # copy appchain crypto-config and modify config.yaml
    print_blue "===> Copy fabric crypto-config"
    if [ ! -d "${CRYPTOPATH}" ]; then
      print_red "crypto-config ${CRYPTOPATH} not found, please start fabric network first"
      exit 1
    fi
    cp -r "${CRYPTOPATH}" "${CONFIG_PATH}"/fabric/
    cp "${CONFIG_PATH}"/fabric/crypto-config/peerOrganizations/org2.example.com/peers/peer1.org2.example.com/msp/signcerts/peer1.org2.example.com-cert.pem "${CONFIG_PATH}"/fabric/fabric.validators

    # copy plugins file to pier root
    print_blue "===> Copy fabric plugin"
    mkdir -p "${CONFIG_PATH}"/plugins
    if [[ "${VERSION}" == "v1.0.0-rc1" || "${VERSION}" == "v1.0.0" ]]; then
      PLUGIN="fabric-client-1.4.so"
    else
      PLUGIN="fabric-client-1.4"
    fi

    if [ ! -f "${PIER_PATH}"/"${PLUGIN}" ]; then
      print_red "pier plugin binary is not downloaded, please download pier plugin for fabric first"
      exit 1
    fi
    cp "${PIER_PATH}"/"${PLUGIN}" "${CONFIG_PATH}"/plugins/"${PLUGIN}"

    cd "${CONFIG_PATH}"
    if [ ! -f fabric_rule.wasm ]; then
      print_blue "===> Downloading fabric_rule.wasm"
      wget https://github.com/meshplus/bitxhub/blob/master/scripts/quick_start/fabric_rule.wasm
    fi
  fi

  if [ "$MODE" == "ethereum" ]; then
    cd "${CURRENT_PATH}"
    print_blue "===> Generate ethereum pier configure"
    # generate config for ethereum pier
    overW='Y'
    if [ "$OVERWRITE" == "false" ]; then
        overW='N'
    fi
    echo $overW|goduck pier config \
      --mode "relay" \
      --type $TYPE \
      --bitxhub ${BITXHUB_ADDR} \
      --validators "0xc7F999b83Af6DF9e67d0a37Ee7e900bF38b3D013" \
      --validators "0x79a1215469FaB6f9c63c1816b45183AD3624bE34" \
      --validators "0x97c8B516D19edBf575D72a172Af7F418BE498C37" \
      --validators "0xc0Ff2e0b3189132D815b8eb325bE17285AC898f8" \
      --appchainType "ethereum" \
      --appchainIP ${APPCHAINIP} \
      --target "${CONFIG_PATH}" \
      --tls "${TLS}" \
      --httpPort "${HTTP}" \
      --pprofPort "${PPROF}"\
      --apiPort "${API}" \
      --version "${VERSION}"

    # copy plugins file to pier root
    print_blue "===> Copy ethereum plugin"
    mkdir -p "${CONFIG_PATH}"/plugins
    if [[ "${VERSION}" == "v1.0.0-rc1" || "${VERSION}" == "v1.0.0" ]]; then
      PLUGIN="eth-client.so"
    else
      PLUGIN="eth-client"
    fi

    if [ ! -f "${PIER_PATH}"/"${PLUGIN}" ]; then
      print_red "pier plugin binary is not downloaded, please download pier plugin for ethereum first"
      exit 1
    fi
    cp "${PIER_PATH}"/"${PLUGIN}" "${CONFIG_PATH}"/plugins/"${PLUGIN}"

    cd "${CONFIG_PATH}"
    if [ ! -f ethereum_rule.wasm ]; then
      print_blue "===> Downloading ethereum_rule.wasm"
      wget https://github.com/meshplus/pier-client-ethereum/blob/master/config/validating.wasm
      mv validating.wasm ethereum_rule.wasm
    fi
  fi

  if [[ "${VERSION}" != "v1.0.0-rc1" && "${VERSION}" != "v1.0.0" ]]; then
    mv "${CONFIG_PATH}"/plugins/"${PLUGIN}" "${CONFIG_PATH}"/plugins/appchain_plugin
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
  generateConfig

  print_blue "===> Start pier of ${MODE}-${VERSION} in ${TYPE}..."
  if [ ! "$(docker ps -q -f name=pier-${MODE})" ]; then
#    if [ "$(docker ps -aq -f status=exited -f name=pier-${MODE})" ]; then
    if [ "$(docker ps -aq -f name=pier-${MODE})" ]; then
      print_red "pier-${MODE} container already exists, please clean them first"
      exit 1
    fi

    print_blue "===> Start a new pier-${MODE} container"
    VERSION=${VERSION:1}
    if [ "$MODE" == "fabric" ]; then
      if [ ! -d "${CRYPTOPATH}" ]; then
        print_red "crypto-config ${CRYPTOPATH} not found, please start fabric network first"
        exit 1
      fi
      if [ $SYSTEM == "linux" ]; then
        docker run -d --name pier-fabric \
        --add-host host.docker.internal:`hostname -I | awk '{print $1}'` \
        -v $CONFIG_PATH/$FABRIC_RULE:/root/.pier/fabric/validating.wasm \
        -v $CONFIG_PATH/pier.toml:/root/.pier/pier.toml \
        -v $CONFIG_PATH/fabric:/root/.pier/fabric \
        -v ${CRYPTOPATH}:/root/.pier/fabric/crypto-config \
        meshplus/pier-fabric:"${VERSION}"
      else
        docker run -d --name pier-fabric \
        -v $CONFIG_PATH/$FABRIC_RULE:/root/.pier/fabric/validating.wasm \
        -v $CONFIG_PATH/pier.toml:/root/.pier/pier.toml \
        -v $CONFIG_PATH/fabric:/root/.pier/fabric \
        -v ${CRYPTOPATH}:/root/.pier/fabric/crypto-config \
        meshplus/pier-fabric:"${VERSION}"
      fi
    elif [ "$MODE" == "ethereum" ]; then
      print_blue "===> Wait for ethereum-node container to start for seconds..."
      sleep 5
      if [ $SYSTEM == "linux" ]; then
        docker run -d --name pier-ethereum \
          --add-host host.docker.internal:`hostname -I | awk '{print $1}'` \
          -v $CONFIG_PATH/$ETHEREUM_RULE:/root/.pier/ethereum/validating.wasm \
          -v $CONFIG_PATH/pier.toml:/root/.pier/pier.toml \
          -v $CONFIG_PATH/ethereum:/root/.pier/ethereum \
          meshplus/pier-ethereum:"${VERSION}"
      else
          docker run -d --name pier-ethereum \
          -v $CONFIG_PATH/$ETHEREUM_RULE:/root/.pier/ethereum/validating.wasm \
          -v $CONFIG_PATH/pier.toml:/root/.pier/pier.toml \
          -v $CONFIG_PATH/ethereum:/root/.pier/ethereum \
          meshplus/pier-ethereum:"${VERSION}"
      fi
    else
      print_red "Not supported mode"
      exit 1
    fi
  else
    print_red "pier-${MODE} container already running, please stop them first"
    exit 1
  fi

  sleep 5
  if [ -z `docker ps -qf "name=pier-$MODE"` ]; then
    print_red "===> Start pier fail"
  else
    print_green "===> Start pier successfully"
    CID=`docker ps -qf "name=pier-$MODE"`
    echo $CID >"${CURRENT_PATH}/pier/pier-${MODE}.cid"
    echo `docker exec $CID pier --repo=/root/.pier id` >"${CURRENT_PATH}/pier/pier-${MODE}-docker.addr"
  fi

}

function pier_binary_up() {
  cd "${CONFIG_PATH}"

  if [ "$MODE" == "fabric" ]; then
    print_blue "===> Deploy rule in bitxhub"
    rule_deploy fabric
    export START_PATH="${CONFIG_PATH}" && export CONFIG_PATH="${CONFIG_PATH}"/fabric
  fi

  if [ "$MODE" == "ethereum" ]; then
    print_blue "===> Deploy rule in bitxhub"
    rule_deploy ethereum
    export START_PATH="${CONFIG_PATH}"
  fi

  print_blue "===> Start pier of ${MODE} in ${TYPE}..."
  nohup "${PIER_PATH}"/pier --repo "${START_PATH}" start >/dev/null 2>&1 &
  PID=$!

  sleep 10
  if [ -n "$(ps -p ${PID} -o pid=)" ]; then
    print_green "===> Start pier successfully!!!"
    echo ${PID} >"${CURRENT_PATH}/pier/pier-${MODE}.pid"
    echo `"${PIER_PATH}"/pier --repo "${START_PATH}" id` >"${CURRENT_PATH}/pier/pier-${MODE}-binary.addr"
  else
    print_red "===> Start pier fail"
  fi
}

function pier_binary_register() {
  extractBin

  generateConfig

  print_blue "===> pier_root: "${CONFIG_PATH}", bitxhub_addr: $BITXHUB_ADDR"

  cd "${CONFIG_PATH}"

  if [ "$MODE" == "fabric" ]; then
    print_blue "===> Register pier(fabric) to bitxhub"
    appchain_register chainA fabric chainA-description 1.4.3 fabric/fabric.validators
  fi

  if [ "$MODE" == "ethereum" ]; then
    print_blue "===> Register pier(ethereum) to bitxhub"
    appchain_register chainB ether chainB-description 1.9.13 ethereum/ether.validators
  fi

  if [[ "${VERSION}" < "v1.6.0" ]]; then
    print_blue "Please use the 'goduck start' command to start the PIER"
  else
    print_blue "Waiting for the administrators of BitXHub to vote for approval. If approved, use the 'goduck start' command to start the PIER"
  fi
}

function pier_register() {
  if [ "${TYPE}" == "docker" ]; then
    print_blue "===> For docker type, versions below v1.6.0 can be registered directly in the START command, version v1.6.0 and above do not support separate PIER registration currently."
  elif [ "${TYPE}" == "binary" ]; then
    pier_binary_register
  else
    echo "Not supported up type "${TYPE}" for pier"
  fi
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
    list=$(cat pier-$MODE.pid)
    for pid in $list; do
      kill "$pid"
      if [ $? -eq 0 ]; then
        echo "pier-$MODE pid:$pid exit"
      else
        print_red "pier exit fail, try use kill -9 $pid"
      fi
    done
    rm pier-$MODE.pid
    rm ${CURRENT_PATH}/pier/pier-${MODE}-binary.addr
  else
    echo "pier-$MODE binary is not running"
  fi

  print_blue "===> Kill $MODE pier in docker"
  cd "${CURRENT_PATH}"/pier
  if [ -a pier-$MODE.cid ]; then
    list=$(cat pier-$MODE.cid)
    for cid in $list; do
      docker kill "$cid"
      if [ $? -eq 0 ]; then
        echo "pier-$MODE container id:$cid exit"
      else
        print_red "pier exit fail"
      fi
    done
    rm pier-$MODE.cid
    rm ${CURRENT_PATH}/pier/pier-${MODE}-docker.addr
  else
    echo "pier-$MODE docker is not running"
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

  if [ -e "${CURRENT_PATH}"/pier/pier-$MODE.pid ]; then
    rm "${CURRENT_PATH}"/pier/pier-$MODE.pid
  fi
  if [ -e "${CURRENT_PATH}"/pier/pier-$MODE-binary.addr ]; then
    rm "${CURRENT_PATH}"/pier/pier-$MODE-binary.addr
  fi

  if [[ ! -z `ps | grep ${CURRENT_PATH}/pier/.pier_$MODE/plugins/appchain_plugin | grep -v "grep"` ]]; then
    echo "clean the plugin process for $MODE pier"
    list=`ps aux| grep ${CURRENT_PATH}/pier/.pier_$MODE/plugins/appchain_plugin | grep -v "grep" | awk '{print $2}'`
    for pluginPID in $list ; do
      kill $pluginPID
      if [ $? -eq 0 ]; then
        echo "pier-$MODE-plugin pid:$pluginPID exit"
      else
        print_red "pier plugin exit fail, try use kill -9 $pluginPID"
      fi
    done
    IFS=$OLD_IFS
  fi

  print_blue "===> Clean $MODE pier in docker"
  if [ "$(docker ps -a -q -f name=pier-$MODE)" ]; then
    docker rm pier-$MODE
  else
    echo "pier-$MODE container is not existed"
  fi

  if [ -e "${CURRENT_PATH}"/pier/pier-$MODE.cid ]; then
    rm "${CURRENT_PATH}"/pier/pier-$MODE.cid
  fi
  if [ -e "${CURRENT_PATH}"/pier/pier-$MODE-docker.addr ]; then
    rm "${CURRENT_PATH}"/pier/pier-$MODE-docker.addr
  fi
}

PIER_ROOT=.pier_fabric
BITXHUB_ADDR="localhost:60011"
MODE="fabric"
TYPE="binary"
VERSION="v1.1.0-rc1"
CRYPTOPATH="$HOME/crypto-config"
PPORT="44550"
APORT="8080"

OPT=$1
shift

while getopts "h?t:m:b:v:c:f:a:l:p:o:i:" opt; do
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
  b)
    BITXHUB_ADDR=$OPTARG
    ;;
  v)
    VERSION=$OPTARG
    ;;
  c)
    CRYPTOPATH=$OPTARG
    ;;
  f)
    PPROF=$OPTARG
    ;;
  a)
    API=$OPTARG
    ;;
  l)
    TLS=$OPTARG
    ;;
  p)
    HTTP=$OPTARG
    ;;
  o)
    OVERWRITE=$OPTARG
    ;;
  i)
    APPCHAINIP=$OPTARG
    ;;
  esac
done

CONFIG_PATH="${CURRENT_PATH}"/pier/.pier_${MODE}
PIER_PATH="${CURRENT_PATH}/bin/pier_${SYSTEM}_${VERSION}"

if [ "$OPT" == "register" ]; then
  pier_register
elif [ "$OPT" == "up" ]; then
  pier_up
elif [ "$OPT" == "down" ]; then
  pier_down
elif [ "$OPT" == "clean" ]; then
  pier_clean
else
  printHelp
  exit 1
fi
