set -e

source x.sh
source compare.sh

CURRENT_PATH=$(pwd)
SYSTEM=$(uname -s)
CONFIG_PATH="${CURRENT_PATH}"/bitxhub
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

function appchain_register_binary() {
  version1=${VERSION}
  version2="v1.8.0"
  version_compare
  if [[ $versionComPareRes -lt 0 ]]; then
    # < v1.8.0
    "${PIER_BIN_PATH}"/pier --repo "${PIERREPO}" appchain register \
      --name $1 \
      --type $2 \
      --desc $3 \
      --version $4 \
      --validators "${PIERREPO}"/$5 \
      --consensusType ""
  else
    # >= v1.8.0
    version1=${VERSION}
    version2="v1.11.3"
    version_compare
    if [[ $versionComPareRes -lt 0 ]]; then
      # v1.8.0 <= < v1.11.2
      "${PIER_BIN_PATH}"/pier --repo "${PIERREPO}" appchain method register \
            --name $1 \
            --type $2 \
            --desc $3 \
            --version $4 \
            --validators "${PIERREPO}"/$5 \
            --admin-key "${PIERREPO}/key.json" \
            --consensus "consensusType" \
            --method "$6" \
            --doc-addr "doc-addr" \
            --doc-hash "doc-hash"
    else
      # >= v1.11.2
      "${PIER_BIN_PATH}"/pier --repo "${PIERREPO}" appchain method register \
                  --name $1 \
                  --type $2 \
                  --desc $3 \
                  --version $4 \
                  --validators "${PIERREPO}"/$5 \
                  --admin-key "${PIERREPO}/key.json" \
                  --consensus "consensusType" \
                  --method "$6" \
                  --doc-addr "doc-addr" \
                  --doc-hash "doc-hash" \
                  --rule "0x00000000000000000000000000000000000000a2" \
                  --rule-url "url" \
                  --reason "reason"
    fi
  fi
}

function pier_docker_rule_deploy() {
  print_blue "======> Deploy rule in bitxhub"

  docker exec $PIERCID /root/.pier/scripts/deployRule.sh /root/.pier/$APPCHAINTYPE/validating.wasm $METHOD $VERSION
}

function pier_binary_rule_deploy() {
  print_blue "======> Deploy rule in bitxhub"
  if [ "${SYSTEM}" == "linux" ]; then
    export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:${PIER_BIN_PATH}
  fi

  version1=${VERSION}
  version2="v1.8.0"
  version_compare
  if [[ $versionComPareRes -lt 0 ]]; then
    #  if [[ "${VERSION}" < "v1.8.0" ]]; then
    "${PIER_BIN_PATH}"/pier --repo "${PIERREPO}" rule deploy --path "${RULEREPO}"
  else
    deployret=$("${PIER_BIN_PATH}"/pier --repo "${PIERREPO}" rule deploy --path "${RULEREPO}" --method "${METHOD}" --admin-key "${PIERREPO}/key.json")
    echo $deployret
    if [[ "${VERSION}" == "v1.8.0" ]]; then
      rule_addr=$(echo $deployret | grep -o "0x.*")
      "${PIER_BIN_PATH}"/pier --repo "${PIERREPO}" rule bind --addr "${rule_addr}" --method "${METHOD}" --admin-key "${PIERREPO}/key.json"
    fi
  fi
}

function pier_docker_up() {
  cp -r ${CURRENT_PATH}/docker/pier ${PIERREPO}/scripts
  cd $PIERREPO/scripts
  chmod +x registerAppchain.sh
  chmod +x deployRule.sh

  print_blue "======> Start pier of ${APPCHAINTYPE}-${VERSION} in ${UPTYPE}..."
  if [ ! "$(docker ps -q -f name=pier-${APPCHAINTYPE})" ]; then

    print_blue "======> Start a new pier-${APPCHAINTYPE}"

    startPierContainer=${PIERREPO}/scripts/docker-compose-pier.yaml
    x_replace "s/container_name: .*/container_name: pier-$APPCHAINTYPE/g" "${startPierContainer}"
    x_replace "s/image: meshplus\/pier:.*/image: meshplus\/pier:${VERSION}/g" "${startPierContainer}"
    HTTPPORT=$(sed '/^.*httpPort/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH})
    PPROFPORT=$(sed '/^.*pprofPort/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH})
    x_replace "s/\".*:34544\"/\"${HTTPPORT}:34544\"/g" "${startPierContainer}"
    x_replace "s/\".*:34555\"/\"${PPROFPORT}:34555\"/g" "${startPierContainer}"
    pierRepoTmp=$(echo "${PIERREPO}" | sed 's/\//\\\//g')
    x_replace "s/pier-fabric-repo/${pierRepoTmp}/g" "${startPierContainer}"

    cp "${startPierContainer}" ${PIER_CONFIG_PATH}/docker-compose-pier.yaml
    docker-compose -f ${PIER_CONFIG_PATH}/docker-compose-pier.yaml up -d
    sleep 1
    CID=`docker ps | grep pier-${APPCHAINTYPE} | awk '{print $1}'`
    echo $CID >"${PIER_CONFIG_PATH}"/pier-${APPCHAINTYPE}.cid
    echo $(docker exec $CID pier id) >"${PIER_CONFIG_PATH}"/pier-${APPCHAINTYPE}-docker.addr
    docker exec $CID pier version >"${PIER_CONFIG_PATH}"/pier-${APPCHAINTYPE}-docker.version
  else
    print_red "pier-${APPCHAINTYPE} container already running, please stop them first"
    exit 1
  fi

  print_blue "You can use the \"goduck status list\" command to check the status of the startup pier."
}

function pier_binary_up() {
  cd "${PIERREPO}"

  print_blue "======> Start pier of ${APPCHAINTYPE} in ${UPTYPE}..."
  nohup "${PIER_BIN_PATH}"/pier --repo "${PIERREPO}" start >/dev/null 2>&1 &
  PID=$!
  echo ${PID} >"${PIER_CONFIG_PATH}"/pier-${APPCHAINTYPE}.pid
  echo $("${PIER_BIN_PATH}"/pier --repo "${PIERREPO}" id) >"${PIER_CONFIG_PATH}"/pier-${APPCHAINTYPE}-binary.addr
  "${PIER_BIN_PATH}"/pier version >"${PIER_CONFIG_PATH}"/pier-${APPCHAINTYPE}-binary.version

  print_blue "You can use the \"goduck status list\" command to check the status of the startup pier."
}

function pier_docker_register() {
  if [ "$APPCHAINTYPE" == "fabric" ]; then
    print_blue "======> Register pier(fabric) to bitxhub"
    docker exec $PIERCID scripts/registerAppchain.sh $METHOD chainA fabric chainA-description 1.4.3 /root/.pier/fabric/fabric.validators consensusType $VERSION
  fi

  if [ "$APPCHAINTYPE" == "ethereum" ]; then
    print_blue "======> Register pier(ethereum) to bitxhub"
    docker exec $PIERCID scripts/registerAppchain.sh "${METHOD}" chainB ether chainB-description 1.9.13 /root/.pier/ethereum/ether.validators consensusType "${VERSION}"
  fi

  print_blue "Waiting for the administrators of BitXHub to vote for approval."
}

function pier_binary_register() {
  if [ ! -f "${PIER_BIN_PATH}"/pier ]; then
    print_red "pier binary is not downloaded, please download pier first"
    exit 1
  fi

  print_green "======> pier_root: ${PIERREPO}"

  # register pier
  if [ "$APPCHAINTYPE" == "fabric" ]; then
    print_blue "======> Register pier(fabric) to bitxhub"
    appchain_register_binary chainA fabric chainA-description 1.4.3 fabric/fabric.validators $METHOD
  fi

  if [ "$APPCHAINTYPE" == "ethereum" ]; then
    print_blue "======> Register pier(ethereum) to bitxhub"
    appchain_register_binary chainB ether chainB-description 1.9.13 ethereum/ether.validators $METHOD
  fi

  print_blue "Waiting for the administrators of BitXHub to vote for approval."
}

function pier_register() {
  if [ "${UPTYPE}" == "docker" ]; then
    pier_docker_register
  elif [ "${UPTYPE}" == "binary" ]; then
    pier_binary_register
  else
    echo "Not supported up type "${UPTYPE}" for pier"
  fi
}

function pier_rule_deploy() {
  if [ "${UPTYPE}" == "docker" ]; then
    pier_docker_rule_deploy
  elif [ "${UPTYPE}" == "binary" ]; then
    pier_binary_rule_deploy
  else
    echo "Not supported up type "${UPTYPE}" for pier"
  fi
}

function pier_up() {
  # generate config
  goduck pier config \
    --appchain "${APPCHAINTYPE}" \
    --target "${PIERREPO}" \
    --configPath "${CONFIGPATH}" \
    --upType "${UPTYPE}" \
    --version "${VERSION}"

  if [ "${UPTYPE}" == "docker" ]; then
    x_replace "s/localhost/host.docker.internal/g" "${PIERREPO}"/pier.toml
    x_replace "s/127.0.0.1/host.docker.internal/g" "${PIERREPO}"/"${APPCHAINTYPE}"/"${APPCHAINTYPE}".toml
    pier_docker_up
  elif [ "${UPTYPE}" == "binary" ]; then
    pier_binary_up
  else
    echo "Not supported up type "${UPTYPE}" for pier"
  fi
}

function pier_down() {
  set +e

  print_blue "======> Kill $APPCHAINTYPE pier in binary"
  if [ -e "${PIER_CONFIG_PATH}"/pier-$APPCHAINTYPE.pid ]; then
    for ((i = 1; ; i++)); do
      pid=$(cat "${PIER_CONFIG_PATH}"/pier-$APPCHAINTYPE.pid | sed -n $i"p")
      if [ -z $pid ]; then
        break
      fi

      kill "$pid"
      if [ $? -eq 0 ]; then
        echo "pier-$APPCHAINTYPE pid:$pid exit"
      else
        print_red "pier exit fail, try use kill -9 $pid"
      fi
    done
  fi

  print_blue "======> Kill $APPCHAINTYPE pier in docker"
  if [ "$(docker ps | grep -c pier-$APPCHAINTYPE)" -ge 1 ]; then
    docker-compose -f ${PIER_CONFIG_PATH}/docker-compose-pier.yaml stop
    echo "pier-$APPCHAINTYPE docker stop"
  fi

  if [ $APPCHAINTYPE == "fabric" ]; then
    cleanPierFabricInfoFile
  else
    cleanPierEtherInfoFile
  fi
}

function pier_clean() {
  set +e

  pier_down

  print_blue "======> Clean $APPCHAINTYPE pier in docker"
  if [ "$(docker ps | grep -c pier-$APPCHAINTYPE)" -ge 1 ]; then
    docker-compose -f ${PIER_CONFIG_PATH}/docker-compose-pier.yaml rm -f
    echo "pier-$APPCHAINTYPE docker clean"
  fi

  print_blue "======> Clean $APPCHAINTYPE pier config"
  if [ -d "${PIER_CONFIG_PATH}"/.pier_$APPCHAINTYPE ]; then
    echo "remove $APPCHAINTYPE pier configure"
    rm -r "${PIER_CONFIG_PATH}"/.pier_$APPCHAINTYPE
  fi

  if [[ ! -z $(ps | grep "${PIER_CONFIG_PATH}"/.pier_$APPCHAINTYPE/plugins/appchain_plugin | grep -v "grep") ]]; then
    echo "clean the plugin process for $APPCHAINTYPE pier"
    list=$(ps aux | grep "${PIER_CONFIG_PATH}"/.pier_$APPCHAINTYPE/plugins/appchain_plugin | grep -v "grep" | awk '{print $2}')
    for pluginPID in $list; do
      kill $pluginPID
      if [ $? -eq 0 ]; then
        echo "pier-$APPCHAINTYPE-plugin pid:$pluginPID exit"
      else
        print_red "pier plugin exit fail, try use kill -9 $pluginPID"
      fi
    done
  fi
}

function cleanPierFabricInfoFile() {
  PIER_CONFIG_PATH="${CURRENT_PATH}"/pier

  if [ -e "${PIER_CONFIG_PATH}"/pier-fabric.pid ]; then
    rm "${PIER_CONFIG_PATH}"/pier-fabric.pid
  fi
  if [ -e "${PIER_CONFIG_PATH}"/pier-fabric-binary.addr ]; then
    rm "${PIER_CONFIG_PATH}"/pier-fabric-binary.addr
  fi
  if [ -e "${PIER_CONFIG_PATH}"/pier-fabric-binary.version ]; then
    rm "${PIER_CONFIG_PATH}"/pier-fabric-binary.version
  fi

  if [ -e "${PIER_CONFIG_PATH}"/pier-fabric.cid ]; then
    rm "${PIER_CONFIG_PATH}"/pier-fabric.cid
  fi
  if [ -e "${PIER_CONFIG_PATH}"/pier-fabric-docker.addr ]; then
    rm "${PIER_CONFIG_PATH}"/pier-fabric-docker.addr
  fi
  if [ -e "${PIER_CONFIG_PATH}"/pier-fabric-docker.addr ]; then
    rm "${PIER_CONFIG_PATH}"/pier-fabric-docker.addr
  fi
}

function cleanPierEtherInfoFile() {
  PIER_CONFIG_PATH="${CURRENT_PATH}"/pier

  if [ -e "${PIER_CONFIG_PATH}"/pier-ethereum.pid ]; then
    rm "${PIER_CONFIG_PATH}"/pier-ethereum.pid
  fi
  if [ -e "${PIER_CONFIG_PATH}"/pier-ethereum-binary.addr ]; then
    rm "${PIER_CONFIG_PATH}"/pier-ethereum-binary.addr
  fi
  if [ -e "${PIER_CONFIG_PATH}"/pier-ethereum-binary.version ]; then
      rm "${PIER_CONFIG_PATH}"/pier-ethereum-binary.version
    fi

  if [ -e "${PIER_CONFIG_PATH}"/pier-ethereum.cid ]; then
    rm "${PIER_CONFIG_PATH}"/pier-ethereum.cid
  fi
  if [ -e "${PIER_CONFIG_PATH}"/pier-ethereum-docker.addr ]; then
    rm "${PIER_CONFIG_PATH}"/pier-ethereum-docker.addr
  fi
  if [ -e "${PIER_CONFIG_PATH}"/pier-ethereum-docker.addr ]; then
    rm "${PIER_CONFIG_PATH}"/pier-ethereum-docker.addr
  fi
}

METHOD=""
OPT=$1
shift

while getopts "h?a:p:c:u:v:r:m:i:" opt; do
  case "$opt" in
  h | \?)
    printHelp
    exit 0
    ;;
  a)
    APPCHAINTYPE=$OPTARG
    ;;
  p)
    PIERREPO=$OPTARG
    ;;
  c)
    CONFIGPATH=$OPTARG
    ;;
  u)
    UPTYPE=$OPTARG
    ;;
  v)
    VERSION=$OPTARG
    ;;
  r)
    RULEREPO=$OPTARG
    ;;
  m)
    METHOD=$OPTARG
    ;;
  i)
    PIERCID=$OPTARG
    ;;
  esac
done

PIER_CONFIG_PATH="${CURRENT_PATH}"/pier
PIER_BIN_PATH="${CURRENT_PATH}/bin/pier_${SYSTEM}_${VERSION}"

if [ "$OPT" == "up" ]; then
  pier_up
elif [ "$OPT" == "register" ]; then
  pier_register
elif [ "$OPT" == "rule" ]; then
  pier_rule_deploy
elif [ "$OPT" == "down" ]; then
  pier_down
elif [ "$OPT" == "clean" ]; then
  pier_clean
else
  printHelp
  exit 1
fi
