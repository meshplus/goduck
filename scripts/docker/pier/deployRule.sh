#!/bin/sh

rulePath=$1
method=$2
pierVersion=$3

OLD_IFS="$IFS"
  IFS="."
  versionArr1=($pierVersion)
  versionArr2=("v1.8.0")
  IFS="$OLD_IFS"

  for ((i = 0; i < 3; i++)); do
    if [ ${#versionArr1[i]} \> ${#versionArr2[i]} ]; then
      versionComPareRes=1
    elif [ ${#versionArr1[i]} \< ${#versionArr2[i]} ]; then
      versionComPareRes=-1
    else
      if [ ${versionArr1[i]} \> ${versionArr2[i]} ]; then
        versionComPareRes=1
      elif [ ${versionArr1[i]} \< ${versionArr2[i]} ]; then
        versionComPareRes=-1
      else
        versionComPareRes=0
      fi
    fi
  done

  if [[ $versionComPareRes -lt 0 ]]; then
#if [ "$pierVersion" \< "v1.8.0" ]; then
  pier --repo /root/.pier rule deploy --path $1
else
  command1=$(pier --repo /root/.pier rule deploy --path $1 --method $2 --admin-key /root/.pier/key.json)
  if [ "$pierVersion" == "v1.8.0" ]; then
    address=$(echo "$command1"|grep -o '0x.\{40\}')
    echo "ruleAddr: ${address}"
    pier --repo /root/.pier rule bind --addr ${address} --method $2 --admin-key /root/.pier/key.json
  fi
fi