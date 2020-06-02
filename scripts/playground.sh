set -e

source x.sh

VERSION=1.0
CURRENT_PATH=$(pwd)

OPT=$1
TYPE=$2
MODE=$3
N=$4

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
  cd "${CURRENT_PATH}"
  if [ -a bin/bitxhub ]; then
    print_blue "Install bitxhub binary"
    goduck install bitxhub "${CURRENT_PATH}"/bin
    cd bin
    if [ "${uname}" == "Linux" ]; then
      tar xf bitxhub_linux_amd64.tar.gz
    elif [ "${uname}" == "Darwin" ]; then
      tar xf bitxhub_macos_x86_64.tar.gz
    else
      echo "Bitxhub does not support the current operating system"
    fi
  fi
}

function bitxhub_binary_solo() {
  binary_prepare
  cd "${CURRENT_PATH}"
  if [ ! -d nodeSolo/plugins ]; then
    mkdir nodeSolo/plugins
    cp -r bin/plugins/solo.so nodeSolo/plugins
  fi
  print_blue "Start bitxhub solo by binary"
  nohup "${CURRENT_PATH}"/bin/bitxhub --repo "${CURRENT_PATH}"/nodeSolo start &
  echo $! >bitxhub.pid
}

function bitxhub_docker_solo() {
  if [[ -z "$(docker images -q meshplus/bitxhub-solo:latest 2>/dev/null)" ]]; then
    docker pull meshplus/bitxhub-solo:latest
  fi

  print_blue "Start bitxhub solo mode by docker"
  docker run -d --name bitxhub_solo \
    -p 60011:60011 -p 9091:9091 -p 53121:53121 -p 40011:40011 \
    -v "${CURRENT_PATH}"/nodeSolo/api:/root/.bitxhub/api \
    -v "${CURRENT_PATH}"/nodeSolo/bitxhub.toml:/root/.bitxhub/bitxhub.toml \
    -v "${CURRENT_PATH}"/nodeSolo/genesis.json:/root/.bitxhub/genesis.json \
    -v "${CURRENT_PATH}"/nodeSolo/network.toml:/root/.bitxhub/network.toml \
    -v "${CURRENT_PATH}"/nodeSolo/order.toml:/root/.bitxhub/order.toml \
    -v "${CURRENT_PATH}"/nodeSolo/certs:/root/.bitxhub/certs \
    meshplus/bitxhub-solo
}

function bitxhub_binary_cluster() {
  binary_prepare

  cd "${CURRENT_PATH}"
  print_blue "Start bitxhub cluster"
  for ((i = 1; i < N + 1; i = i + 1)); do
    if [ ! -d node${i}/plugins ]; then
      mkdir node${i}/plugins
      cp -r bin/plugins/raft.so node${i}/plugins
    fi
    echo "Start bitxhub node${i}"
    nohup "${CURRENT_PATH}"/bin/bitxhub --repo="${CURRENT_PATH}"/node${i} start &
    echo $!+"/n" >"${CURRENT_PATH}"/bitxhub.pid
  done
}

function bitxhub_docker_cluster() {
  if [[ -z "$(docker images -q meshplus/bitxhub:latest 2>/dev/null)" ]]; then
    docker pull meshplus/bitxhub:latest
  fi
  print_blue "Start bitxhub cluster mode by docker compose"
  docker-compose -f "${CURRENT_PATH}"/docker/docker-compose.yml up -d
}

function bitxhub_down() {
  if [ -a "${CURRENT_PATH}"/bitxhub.pid ]; then
    list=$(cat "${CURRENT_PATH}"/bitxhub.pid)
    for pid in $list; do
      kill "$pid"
      if [ $? -eq 0 ]; then
        echo "program exit."
      else
        echo "program exit fail, try use kill -9 $pid"
      fi
    done
    rm "${CURRENT_PATH}"/bitxhub.pid
  fi

  if [ "$(docker container ls -a | grep -c bitxhub_node)" -ge 1 ]; then
    docker-compose -f "${CURRENT_PATH}"/docker/docker-compose.yml stop
  fi

  if [ "$(docker container ls -a | grep -c bitxhub_solo)" -ge 1 ]; then
    docker stop bitxhub_solo
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
      echo "TYPE should be solo or cluster"
      exit 1
      ;;
    esac
    ;;
  "binary")
    case $TYPE in
    "solo")
      bitxhub_binary_cluster
      ;;
    "cluster")
      bitxhub_binary_cluster
      ;;
    *)
      echo "TYPE should be solo or cluster"
      exit 1
      ;;
    esac
    ;;
  *)
    print_blue "MODE should be docker or binary"
    exit 1
    ;;
  esac
}

function bitxhub_restart() {
  bitxhub_down
  bitxhub_up
}

print_blue "===> Script version: $VERSION"

if [ "$OPT" == "up" ]; then
  bitxhub_up
elif [ "$OPT" == "down" ]; then
  bitxhub_down
elif [ "$OPT" == "restart" ]; then
  bitxhub_restart
else
  printHelp
  exit 1
fi
