set -e

source x.sh

CURRENT_PATH=$(pwd)
GODUCK_REPO_PATH=~/.goduck
PIER_CLIENT_FABRIC_VERSION=master
PIER_CLIENT_ETHEREUM_VERSION=master
FABRIC_RULE=fabric_rule.wasm
ETHEREUM_RULE=ethereum_rule.wasm
FABRIC_PLUGIN=fabric-client-1.4
ETHEREUM_PLUGIN=eth-client
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
      --appchainIP "${APPCHAINIP}" \
      --appchainPorts "${APPCHAINPORTS}" \
      --appchainAddr "${APPCHAINADDR}" \
      --contractAddr "${APPCHAINCONTRACTADDR}" \
      --target "${PIERREPO}" \
      --tls "${TLS}" \
      --httpPort "${HTTP}" \
      --pprofPort "${PPROF}" \
      --apiPort "${API}" \
      --cryptoPath "${CRYPTOPATH}" \
      --method "${METHOD}" \
      --version "${VERSION}"

    # copy appchain crypto-config and modify config.yaml
    print_blue "===> Copy fabric crypto-config"
    if [ ! -d "${CRYPTOPATH}" ]; then
      print_red "crypto-config ${CRYPTOPATH} not found, please start fabric network first"
      exit 1
    fi
    cp -r "${CRYPTOPATH}" "${PIERREPO}"/fabric/
    cp "${PIERREPO}"/fabric/crypto-config/peerOrganizations/org2.example.com/peers/peer1.org2.example.com/msp/signcerts/peer1.org2.example.com-cert.pem "${PIERREPO}"/fabric/fabric.validators

    # copy plugins file to pier root
    print_blue "===> Copy fabric plugin"
    mkdir -p "${PIERREPO}"/plugins
    if [[ "${VERSION}" == "v1.0.0-rc1" || "${VERSION}" == "v1.0.0" ]]; then
      PLUGIN="fabric-client-1.4.so"
      PLUGINNAME="fabric-client-1.4.so"
    else
      PLUGIN="fabric-client-1.4"
      PLUGINNAME="appchain_plugin"
    fi

    PIER_PLUGIN_PATH="${PIER_PATH}"
    if [ $TYPE == "docker" ]; then
      PIER_PLUGIN_PATH="${PIER_LINUX_PATH}"
    fi
    if [ ! -f "${PIER_PLUGIN_PATH}"/"${PLUGIN}" ]; then
      print_red "pier plugin binary is not downloaded, please download pier plugin for fabric first"
      exit 1
    fi
    cp "${PIER_PLUGIN_PATH}"/"${PLUGIN}" "${PIERREPO}"/plugins/"${PLUGINNAME}"

    cd "${PIERREPO}"/fabric
    if [ ! -f validating.wasm ]; then
      print_blue "===> Downloading validating.wasm"
      wget https://raw.githubusercontent.com/meshplus/pier-client-fabric/master/config/validating.wasm
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
      --appchainIP "${APPCHAINIP}" \
      --appchainAddr "${APPCHAINADDR}" \
      --appchainPorts "${APPCHAINPORTS}" \
      --contractAddr "${APPCHAINCONTRACTADDR}" \
      --target "${PIERREPO}" \
      --tls "${TLS}" \
      --httpPort "${HTTP}" \
      --pprofPort "${PPROF}"\
      --apiPort "${API}" \
      --method "${METHOD}" \
      --version "${VERSION}"

    # copy plugins file to pier root
    print_blue "===> Copy ethereum plugin"
    mkdir -p "${PIERREPO}"/plugins
    if [[ "${VERSION}" == "v1.0.0-rc1" || "${VERSION}" == "v1.0.0" ]]; then
      PLUGIN="eth-client.so"
      PLUGINNAME="eth-client.so"
    else
      PLUGIN="eth-client"
      PLUGINNAME="appchain_plugin"
    fi

    PIER_PLUGIN_PATH="${PIER_PATH}"
    if [ $TYPE == "docker" ]; then
      PIER_PLUGIN_PATH="${PIER_LINUX_PATH}"
    fi
    if [ ! -f "${PIER_PLUGIN_PATH}"/"${PLUGIN}" ]; then
      print_red "pier plugin binary is not downloaded, please download pier plugin for ethereum first"
      exit 1
    fi


    cp "${PIER_PLUGIN_PATH}"/"${PLUGIN}" "${PIERREPO}"/plugins/"${PLUGINNAME}"
    cd "${PIERREPO}"/ether
    if [ ! -f validating.wasm ]; then
      print_blue "===> Downloading validating.wasm"
      wget https://raw.githubusercontent.com/meshplus/pier-client-ethereum/master/config/validating.wasm
    fi
  fi
}

function appchain_register_binary() {
  if [[ "${VERSION}" < "v1.7.0" ]]; then
    "${PIER_PATH}"/pier --repo "${PIERREPO}" appchain register \
      --name $1 \
      --type $2 \
      --desc $3 \
      --version $4 \
      --validators "${PIERREPO}"/$5
  elif [[ "${VERSION}" == "v1.7.0" ]]; then
    "${PIER_PATH}"/pier --repo "${PIERREPO}" appchain register \
      --name $1 \
      --type $2 \
      --desc $3 \
      --version $4 \
      --validators "${PIERREPO}"/$5 \
      --consensusType ""
  else
    "${PIER_PATH}"/pier --repo "${PIERREPO}" appchain method register \
      --name $1 \
      --type $2 \
      --desc $3 \
      --version $4 \
      --validators "${PIERREPO}"/$5 \
      --admin-key $ADMINKEY \
      --consensus "consensusType" \
      --method $METHOD \
      --doc-addr "doc-addr" \
      --doc-hash "doc-hash"
  fi
}

function pier_docker_rule_deploy() {
  print_blue "======> Deploy rule in bitxhub"

  docker exec $CID /root/.pier/scripts/deployRule.sh /root/.pier/$MODE/validating.wasm $METHOD $VERSION

  if [[ "${VERSION}" < "v1.8.0" ]]; then
    print_blue "Please use the 'goduck pier start' command to start the PIER"
  else
    print_blue "Waiting for the administrators of BitXHub to vote for approval. If approved, use the 'goduck pier start' command to start PIER"
  fi
}

function pier_binary_rule_deploy() {
  print_blue "======> Deploy rule in bitxhub"
  if [ "${SYSTEM}" == "linux" ]; then
    export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${PIER_PATH}
  fi

  if [[ "${VERSION}" < "v1.8.0" ]]; then
    "${PIER_PATH}"/pier --repo "${PIERREPO}" rule deploy --path "${RULEREPO}"
    print_blue "Please use the 'goduck pier start' command to start the PIER"
  else
    deployret=`"${PIER_PATH}"/pier --repo "${PIERREPO}" rule deploy --path "${RULEREPO}" --method "${METHOD}" --admin-key "${ADMINKEY}"`
    echo $deployret
    rule_addr=`echo $deployret | grep -o "0x.*"`
    "${PIER_PATH}"/pier --repo "${PIERREPO}" rule bind --addr "${rule_addr}" --method "${METHOD}" --admin-key "${ADMINKEY}"
    print_blue "Waiting for the administrators of BitXHub to vote for approval. If approved, use the 'goduck pier start' command to start PIER"
  fi
}

function pier_docker_up() {
  print_blue "===> Start pier of ${MODE} in ${TYPE}..."

  docker exec $CID /root/.pier/scripts/startPier.sh
}

function pier_binary_up() {
  cd "${PIERREPO}"

  print_blue "===> Start pier of ${MODE} in ${TYPE}..."
  nohup "${PIER_PATH}"/pier --repo "${PIERREPO}" start >/dev/null 2>&1 &
  PID=$!

  echo ${PID} >"${CONFIG_PATH}"/pier-${MODE}.pid
  echo `"${PIER_PATH}"/pier --repo "${PIERREPO}" id` >"${CONFIG_PATH}"/pier-${MODE}-binary.addr
  print_blue "You can use the \"goduck status list\" command to check the status of the startup pier."
}

function copyPlugin() {
  cd $PIERREPO
  mkdir plugins
  if [ "$MODE" == "fabric" ]; then
    cp ${PIER_PATH}/$FABRIC_PLUGIN $PIERREPO/plugins/appchain_plugin
  elif [ "$MODE" == "ethereum" ]; then
    cp ${PIER_PATH}/$ETHEREUM_PLUGIN $PIERREPO/plugins/appchain_plugin
  else
    print_red "Not supported mode"
    exit 1
  fi
}

function start_pier_container() {
  print_blue "===> Start pier container of ${MODE}-${VERSION} in ${TYPE}..."
  if [ ! "$(docker ps -q -f name=pier-${MODE})" ]; then
    if [ "$(docker ps -aq -f name=pier-${MODE})" ]; then
      print_red "pier-${MODE} container already exists, please clean them first"
      exit 1
    fi

    print_blue "===> Start a new pier-${MODE} container"
    if [ "$MODE" == "fabric" ]; then
      if [ ! -d "${CRYPTOPATH}" ]; then
        print_red "crypto-config ${CRYPTOPATH} not found, please start fabric network first"
        exit 1
      fi
      
      if [ $SYSTEM == "linux" ]; then
        if [[ "${VERSION}" < "v1.6.0" ]]; then
          docker run -d --name pier-fabric \
            --add-host host.docker.internal:`hostname -I | awk '{print $1}'` \
            -v $PIERREPO/pier.toml:/root/.pier/pier.toml \
            -v $PIERREPO/fabric:/root/.pier/fabric \
            -v ${CRYPTOPATH}:/root/.pier/fabric/crypto-config \
            meshplus/pier-fabric:"${VERSION}"
        else
          docker run -d --name pier-fabric \
            --add-host host.docker.internal:`hostname -I | awk '{print $1}'` \
            -v $PIERREPO/pier.toml:/root/.pier/pier.toml \
            -v $PIERREPO/fabric:/root/.pier/fabric \
            -v $PIERREPO/plugins:/root/.pier/plugins \
            -v ${CRYPTOPATH}:/root/.pier/fabric/crypto-config \
            meshplus/pier-fabric:"${VERSION}"
        fi
      else
        if [[ "${VERSION}" < "v1.6.0" ]]; then
          docker run -d --net host --name pier-fabric \
            -v $PIERREPO/pier.toml:/root/.pier/pier.toml \
            -v $PIERREPO/fabric:/root/.pier/fabric \
            -v ${CRYPTOPATH}:/root/.pier/fabric/crypto-config \
            meshplus/pier-fabric:"${VERSION}"
        else
          docker run -d --net host --name pier-fabric \
            -v $PIERREPO/pier.toml:/root/.pier/pier.toml \
            -v $PIERREPO/fabric:/root/.pier/fabric \
            -v $PIERREPO/plugins:/root/.pier/plugins \
            -v ${CRYPTOPATH}:/root/.pier/fabric/crypto-config \
            meshplus/pier-fabric:"${VERSION}"
        fi
      fi
    elif [ "$MODE" == "ethereum" ]; then
      print_blue "===> Wait for ethereum-node container to start for seconds..."
      sleep 5
      if [ $SYSTEM == "linux" ]; then
        if [[ "${VERSION}" < "v1.6.0" ]]; then
          docker run -d --name pier-ethereum \
            --add-host host.docker.internal:`hostname -I | awk '{print $1}'` \
            -v $PIERREPO/pier.toml:/root/.pier/pier.toml \
            -v $PIERREPO/ether:/root/.pier/ether \
            meshplus/pier-ethereum:"${VERSION}"
        else
          docker run -d --name pier-ethereum \
            --add-host host.docker.internal:`hostname -I | awk '{print $1}'` \
            -v $PIERREPO/pier.toml:/root/.pier/pier.toml \
            -v $PIERREPO/ether:/root/.pier/ether \
            -v $PIERREPO/plugins:/root/.pier/plugins \
            meshplus/pier-ethereum:"${VERSION}"
        fi
      else
        if [[ "${VERSION}" < "v1.6.0" ]]; then
          docker run -d --net host --name pier-ethereum \
            -v $PIERREPO/pier.toml:/root/.pier/pier.toml \
            -v $PIERREPO/ether:/root/.pier/ether \
            meshplus/pier-ethereum:"${VERSION}"
        else
          docker run -d --net host --name pier-ethereum \
            -v $PIERREPO/pier.toml:/root/.pier/pier.toml \
            -v $PIERREPO/ether:/root/.pier/ether \
            -v $PIERREPO/plugins:/root/.pier/plugins \
            meshplus/pier-ethereum:"${VERSION}"
        fi
      fi
    else
      print_red "Not supported mode"
      exit 1
    fi

    startPierContainer=${PIERREPO}/scripts/docker-compose-pier.yaml
    x_replace "s/container_name: .*/container_name: pier-$MODE/g" "${startPierContainer}"
    x_replace "s/image: meshplus\/pier:.*/image: meshplus\/pier:${VERSION}/g" "${startPierContainer}"
    x_replace "s/\".*:34544\"/\"${HTTP}:34544\"/g" "${startPierContainer}"
    x_replace "s/\".*:34555\"/\"${PPROF}:34555\"/g" "${startPierContainer}"
    pierRepoTmp=$(echo "${PIERREPO}"|sed 's/\//\\\//g')
    x_replace "s/pier-fabric-repo/${pierRepoTmp}/g" "${startPierContainer}"

    docker-compose -f ${PIERREPO}/scripts/docker-compose-pier.yaml up -d
  else
    print_red "pier-${MODE} container already running, please stop them first"
    exit 1
  fi

  sleep 5
  if [ -z `docker ps -qf "name=pier-$MODE"` ]; then
    print_red "===> Start pier container fail"
  else
    print_green "===> Start pier container successfully"
    CID=`docker ps -qf "name=pier-$MODE"`
    echo $CID >"${CONFIG_PATH}"/pier-${MODE}.cid"
    echo `docker exec $CID pier --repo=/root/.pier id` >"${CONFIG_PATH}"/pier-${MODE}-docker.addr"
  fi
}

function copyScripts() {
  cd $PIERREPO
  cp -r ${CURRENT_PATH}/docker/pier ${PIERREPO}/scripts

  cd $PIERREPO/scripts
  chmod +x registerAppchain.sh
  chmod +x deployRule.sh
  chmod +x startPier.sh
  chmod +x vote.sh
}

function pier_docker_register() {
  generateConfig

  copyScripts

  start_pier_container

docker exec
  if [ "$MODE" == "fabric" ]; then
    print_blue "===> Register pier(fabric) to bitxhub"
    docker exec $CID scripts/registerAppchain.sh $METHOD chainA fabric chainA-description 1.4.3 /root/.pier/fabric/fabric.validators consensusType $VERSION
  fi

  if [ "$MODE" == "ethereum" ]; then
    print_blue "===> Register pier(ethereum) to bitxhub"
    docker exec $CID scripts/registerAppchain.sh $METHOD chainB ether chainB-description 1.9.13 /root/.pier/ethereum/ether.validators consensusType $VERSION
  fi

  print_blue "Waiting for the administrators of BitXHub to vote for approval. If approved, use the 'goduck pier rule' command to deploy rule to bitxhub"
}

function pier_binary_register() {
  extractBin

  generateConfig

  print_blue "===> pier_root: "${PIERREPO}", bitxhub_addr: $BITXHUB_ADDR"

  cd "${PIERREPO}"

  if [ "$MODE" == "fabric" ]; then
    print_blue "===> Register pier(fabric) to bitxhub"
    appchain_register_binary chainA fabric chainA-description 1.4.3 fabric/fabric.validators
  fi

  if [ "$MODE" == "ethereum" ]; then
    print_blue "===> Register pier(ethereum) to bitxhub"
    appchain_register_binary chainB ether chainB-description 1.9.13 ethereum/ether.validators
  fi

  if [[ "${VERSION}" < "v1.6.0" ]]; then
    print_blue "Please use the 'goduck pier start' command to start the PIER"
  else
    print_blue "Waiting for the administrators of BitXHub to vote for approval. If approved, use the 'goduck pier rule' command to deploy rule to bitxhub"
  fi
}

function pier_register() {
  if [ "${TYPE}" == "docker" ]; then
    pier_docker_register
  elif [ "${TYPE}" == "binary" ]; then
    pier_binary_register
  else
    echo "Not supported up type "${TYPE}" for pier"
  fi
}

function pier_rule_deploy() {
  if [ "${TYPE}" == "docker" ]; then
    pier_docker_rule_deploy
  elif [ "${TYPE}" == "binary" ]; then
    pier_binary_rule_deploy
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
  if [ -a "${CONFIG_PATH}"/pier-$MODE.pid ]; then
    list=$(cat "${CONFIG_PATH}"/pier-$MODE.pid)
    for pid in $list; do
      kill "$pid"
      if [ $? -eq 0 ]; then
        echo "pier-$MODE pid:$pid exit"
      else
        print_red "pier exit fail, try use kill -9 $pid"
      fi
    done
    rm "${CONFIG_PATH}"/pier-$MODE.pid
  else
    echo "pier-$MODE binary is not running"
  fi

  print_blue "===> Kill $MODE pier in docker"
  if [ -a "${CONFIG_PATH}"/pier-$MODE.cid ]; then
    list=$(cat "${CONFIG_PATH}"/pier-$MODE.cid)
    for cid in $list; do
      docker kill "$cid"
      if [ $? -eq 0 ]; then
        echo "pier-$MODE container id:$cid exit"
      else
        print_red "pier exit fail"
      fi
    done
    rm "${CONFIG_PATH}"/pier-$MODE.cid
  else
    echo "pier-$MODE docker is not running"
  fi
}

function pier_clean() {
  set +e

  pier_down

  print_blue "===> Clean $MODE pier in docker"
  if [ "$(docker ps -a -q -f name=pier-$MODE)" ]; then
    docker rm pier-$MODE
  else
    echo "pier-$MODE container is not existed"
  fi


  print_blue "===> Clean $MODE pier config"
  if [ -d "${CONFIG_PATH}"/.pier_$MODE ]; then
    echo "remove $MODE pier configure"
    rm -r "${CONFIG_PATH}"/.pier_$MODE
  else
    echo "pier-$MODE configure is not existed"
  fi

  if [[ ! -z `ps | grep "${CONFIG_PATH}"/pier/.pier_$MODE/plugins/appchain_plugin | grep -v "grep"` ]]; then
    echo "clean the plugin process for $MODE pier"
    list=`ps aux| grep "${CONFIG_PATH}"/pier/.pier_$MODE/plugins/appchain_plugin | grep -v "grep" | awk '{print $2}'`
    for pluginPID in $list ; do
      kill $pluginPID
      if [ $? -eq 0 ]; then
        echo "pier-$MODE-plugin pid:$pluginPID exit"
      else
        print_red "pier plugin exit fail, try use kill -9 $pluginPID"
      fi
    done
  fi

  cleanPierInfoFile
}

function cleanPierInfoFile(){
  PIER_CONFIG_PATH="${CURRENT_PATH}"/pier

  if [ -e "${PIER_CONFIG_PATH}"/pier-ethereum.pid ]; then
    rm "${PIER_CONFIG_PATH}"/pier-ethereum.pid
  fi
  if [ -e "${PIER_CONFIG_PATH}"/pier-ethereum-binary.addr ]; then
    rm "${PIER_CONFIG_PATH}"/pier-ethereum-binary.addr
  fi
  if [ -e "${PIER_CONFIG_PATH}"/pier-fabric.pid ]; then
    rm "${PIER_CONFIG_PATH}"/pier-fabric.pid
  fi
  if [ -e "${PIER_CONFIG_PATH}"/pier-fabric-binary.addr ]; then
    rm "${PIER_CONFIG_PATH}"/pier-fabric-binary.addr
  fi


cleanPierInfoFile
}

function cleanPierInfoFile(){
  PIER_CONFIG_PATH="${CURRENT_PATH}"/pier

  if [ -e "${PIER_CONFIG_PATH}"/pier-ethereum.pid ]; then
    rm "${PIER_CONFIG_PATH}"/pier-ethereum.pid
  fi
  if [ -e "${PIER_CONFIG_PATH}"/pier-ethereum-binary.addr ]; then
    rm "${PIER_CONFIG_PATH}"/pier-ethereum-binary.addr
  fi
  if [ -e "${PIER_CONFIG_PATH}"/pier-fabric.pid ]; then
    rm "${PIER_CONFIG_PATH}"/pier-fabric.pid
  fi
  if [ -e "${PIER_CONFIG_PATH}"/pier-fabric-binary.addr ]; then
    rm "${PIER_CONFIG_PATH}"/pier-fabric-binary.addr
  fi

  if [ -e "${PIER_CONFIG_PATH}"/pier-ethereum.cid ]; then
    rm "${PIER_CONFIG_PATH}"/pier-ethereum.cid
  fi
  if [ -e "${PIER_CONFIG_PATH}"/pier-ethereum-docker.addr ]; then
    rm "${PIER_CONFIG_PATH}"/pier-ethereum-docker.addr
  fi
  if [ -e "${PIER_CONFIG_PATH}"/pier-fabric.cid ]; then
    rm "${PIER_CONFIG_PATH}"/pier-fabric.cid
  fi

  if [ -e "${PIER_CONFIG_PATH}"/pier-fabric-docker.addr ]; then
    rm "${PIER_CONFIG_PATH}"/pier-fabric-docker.addr
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

while getopts "h?t:m:b:v:c:f:a:l:p:o:i:d:s:n:r:u:k:e:" opt; do
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
  d)
    APPCHAINADDR=$OPTARG
    ;;
  s)
    APPCHAINPORTS=$OPTARG
    ;;
  n)
    APPCHAINCONTRACTADDR=$OPTARG
    ;;
  r)
    PIERREPO=$OPTARG
    ;;
  u)
    RULEREPO=$OPTARG
    ;;
  k)
    ADMINKEY=$OPTARG
    ;;
  e)
    METHOD=$OPTARG
    ;;
  esac
done

CONFIG_PATH="${CURRENT_PATH}"/pier
PIER_PATH="${CURRENT_PATH}/bin/pier_${SYSTEM}_${VERSION}"
PIER_LINUX_PATH="${CURRENT_PATH}/bin/pier_linux_${VERSION}"

if [ "$OPT" == "register" ]; then
  pier_register
elif [ "$OPT" == "up" ]; then
  pier_up
elif [ "$OPT" == "down" ]; then
  pier_down
elif [ "$OPT" == "clean" ]; then
  pier_clean
elif [ "$OPT" == "rule" ]; then
  pier_rule_deploy
else
  printHelp
  exit 1
fi
