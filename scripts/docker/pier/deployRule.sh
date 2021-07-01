#!/bin/sh

rulePath=$1
method=$2
pierVersion=$3

if [ "$pierVersion" \< "v1.8.0" ]; then
  pier --repo /root/.pier rule deploy --path $1
else
  command1=$(pier --repo /root/.pier rule deploy --path $1 --method $2 --admin-key /root/.pier/key.json)
  if [ "$pierVersion" == "v1.8.0" ]; then
    address=$(echo "$command1"|grep -o '0x.\{40\}')
    echo "ruleAddr: ${address}"
    pier --repo /root/.pier rule bind --addr ${address} --method $2 --admin-key /root/.pier/key.json
  fi
fi