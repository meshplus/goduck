function echo_err() {
  printf "%s\n" "$*" 1>&2
}

function printHelp() {
  echo "Usage:  "
  echo "  wait-for.sh [host:port | url] [-p protocol] [-t timeout] [-r response] [command]"
  echo "      - 'host:port | url' - Like host:port or url, e.g 127.0.0.1:8080 or 127.0.0.9091/v1/chain_meta"
  echo "      - '-p protocol' - The protocol to make the request with, either tcp or http, default is tcp"
  echo "      - '-t timeout' - Timeout in seconds, default in 15 seconds"
  echo "      - '-r response' - If response check is required, use this argument. Until the response is consistent"
  echo "      - 'command' - The command after wait for server to do"
  echo "  wait-for.sh -h (print this message)"
}

REQUEST=$1
shift
args=0

while getopts "h?p:t:r:" opt; do
  case "$opt" in
  h | \?)
    printHelp
    exit 0
    ;;
  p)
    PROTOCOL="$OPTARG"
    args=$((args + 2))
    ;;
  t)
    TIMEOUT="$OPTARG"
    args=$((args + 2))
    ;;
  r)
    RESPONSE="$OPTARG"
    args=$((args + 2))
    ;;
  esac
done

if [ -z "$REQUEST" ] || [ $# == 0 ]; then
  printHelp
  exit 0
fi
if [ -z "$PROTOCOL" ]; then
  PROTOCOL="tcp"
fi
if [ -z "$TIMEOUT" ]; then
  TIMEOUT=15
fi

shift $args

function wait_for() {
  case $PROTOCOL in
  "tcp")
    if ! command -v nc >/dev/null; then
      echo_err 'nc command is missing!'
      exit 1
    fi
    ;;
  "http")
    if ! command -v wget >/dev/null; then
      echo_err 'wget command is missing!'
      exit 1
    fi
    ;;
  esac
  while true; do
    case $PROTOCOL in
    "tcp")
      HOST=$(echo "$REQUEST" | cut -d ':' -f1)
      PORT=$(echo "$REQUEST" | cut -d ':' -f2)
      nc -z "$HOST" "$PORT" >/dev/null 2>&1
      result=$?
      if [ $result -eq 0 ]; then
        exec "$@"
      fi
      ;;
    "http")
      time=$(date +%s)
      wget -T 1 "$REQUEST" -q -O response_"$time"
      resp=$(cat response_"$time")
      rm -rf response_"$time"
      result=$?
      if [ $result -eq 0 ] && [ -z "$RESPONSE" ] || [ "$RESPONSE" == "$resp" ]; then
        exec "$@"
      fi
      ;;
    esac

    if [ $TIMEOUT -eq 0 ]; then
      break
    fi
    TIMEOUT=$((TIMEOUT - 1))
    sleep 1
  done
  echo_err "Operation timed out"
  exit 1
}

wait_for "$@"