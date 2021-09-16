#!/bin/sh

rulePath=$1
method=$2
pierVersion=$3

if [ "${pierVersion}" = "v1.6.1" ] || [ "${pierVersion}" = "v1.6.2" ] || [ "${pierVersion}" = "v1.7.0" ]; then

  commandStr="pier --repo /root/.pier rule deploy --path $1"
  error=false
  commandRes=$($commandStr || error=true)
  while [ ${error} == true ] || [ $(expr match "$commandRes" 'error') != 0 ]; do
    echo $commandRes
    error=false
    commandRes=$($commandStr || error=true)
  done
  echo $commandRes

else

  commandStr="pier --repo /root/.pier rule deploy --path $1 --method $2 --admin-key /root/.pier/key.json"
  error=false
  commandRes=$($commandStr || error=true)
  while [ ${error} == true ] || [ $(expr match "$commandRes" 'error') != 0 ]; do
    echo $commandRes
    error=false
    commandRes=$($commandStr || error=true)
  done
  echo $commandRes

  if [ "$pierVersion" == "v1.8.0" ]; then
    address=$(echo "$commandRes"|grep -o '0x.\{40\}')
    echo "ruleAddr: ${address}"

    commandStr="pier --repo /root/.pier rule bind --addr ${address} --method $2 --admin-key /root/.pier/key.json"
    error=false
    commandRes=$($commandStr || error=true)
    while [ ${error} == true ] || [ $(expr match "$commandRes" 'error') != 0 ]; do
      echo $commandRes
      error=false
      commandRes=$($commandStr || error=true)
    done
    echo $commandRes
  fi
fi