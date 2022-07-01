#!/bin/sh

methodName=$1 # Please note: methodName must be consistent with pier.toml
appchainName=$2
appchainType=$3 # Please note: appchainType must be consistent with pier.toml
appchainDesc=$4
appchainVersion=$5
appchainValidators=$6
appchainConsensus=$7
pierVersion=$8
broker=$9
admin=${10}

if [ "${pierVersion}" = "v1.6.1" ] || [ "${pierVersion}" = "v1.6.2" ] || [ "${pierVersion}" = "v1.7.0" ]; then
  pier --repo /root/.pier appchain register \
    --name $appchainName \
    --type $appchainType \
    --desc $appchainDesc \
    --version $appchainVersion \
    --validators $appchainValidators \
    --consensusType $appchainConsensus
elif [ "${pierVersion}" = "v1.8.0" ] || [ "${pierVersion}" = "v1.9.0" ] || [ "${pierVersion}" = "v1.11.0" ] || [ "${pierVersion}" = "v1.11.1" ]; then
  pier --repo /root/.pier appchain method register \
    --admin-key /root/.pier/key.json \
    --method $methodName \
    --doc-addr /ipfs/QmQVxzUqN2Yv2UHUQXYwH8dSNkM8ReJ9qPqwJsf8zzoNUi \
    --doc-hash QmQVxzUqN2Yv2UHUQXYwH8dSNkM8ReJ9qPqwJsf8zzoNUi \
    --name $appchainName \
    --type $appchainType \
    --desc $appchainDesc \
    --version $appchainVersion \
    --validators $appchainValidators \
    --consensus $appchainConsensus
elif [ "${pierVersion}" = "v1.23.0" ]; then
  pier --repo /root/.pier  appchain register \
  --appchain-id $methodName \
  --name $appchainName \
  --type $appchainType \
  --trustroot $appchainValidators \
  --broker $broker \
  --desc "desc" \
  --master-rule "0x00000000000000000000000000000000000000a2" \
  --rule-url "http://github.com" \
  --admin $admin \
  --reason "reason"
else
  pier --repo /root/.pier appchain method register \
    --admin-key /root/.pier/key.json \
    --method $methodName \
    --doc-addr /ipfs/QmQVxzUqN2Yv2UHUQXYwH8dSNkM8ReJ9qPqwJsf8zzoNUi \
    --doc-hash QmQVxzUqN2Yv2UHUQXYwH8dSNkM8ReJ9qPqwJsf8zzoNUi \
    --name $appchainName \
    --type $appchainType \
    --desc $appchainDesc \
    --version $appchainVersion \
    --validators $appchainValidators \
    --consensus $appchainConsensus \
    --rule 0x00000000000000000000000000000000000000a2 \
    --rule-url "url"
fi