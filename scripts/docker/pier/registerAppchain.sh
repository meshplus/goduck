#!/bin/sh

methodName=$1 # Please note: methodName must be consistent with pier.toml
appchainName=$2
appchainType=$3 # Please note: appchainType must be consistent with pier.toml
appchainDesc=$4
appchainVersion=$5
appchainValidators=$6
appchainConsensus=$7
pierVersion=$8

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
#if [ "${pierVersion}" \< "v1.8.0" ]; then
  pier --repo /root/.pier appchain register \
    --name $2 \
    --type $3 \
    --desc $4 \
    --version $5 \
    --validators $6 \
    --consensusType $7
else
  pier --repo /root/.pier appchain method register \
    --admin-key /root/.pier/key.json \
    --method $1 \
    --doc-addr /ipfs/QmQVxzUqN2Yv2UHUQXYwH8dSNkM8ReJ9qPqwJsf8zzoNUi \
    --doc-hash QmQVxzUqN2Yv2UHUQXYwH8dSNkM8ReJ9qPqwJsf8zzoNUi \
    --name $2 \
    --type $3 \
    --desc $4 \
    --version $5 \
    --validators $6 \
    --consensus $7
fi