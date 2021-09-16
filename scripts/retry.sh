function command_retry() {
  commandStr=$1
  error=false
  commandRes=$($commandStr || error=true)
  while [ ${error} == true ] || [[ "$commandRes" =~ error ]]; do
    echo $commandRes
    error=false
    commandRes=$($commandStr || error=true)
  done
  echo $commandRes
}