#!/bin/sh


appchainId=$1
serviceId=$2
name=$3
pierVersion=$4

if [ "${pierVersion}" = "v1.23.0" ] || [ "${pierVersion}" = "v2.8.0" ] || [ "${pierVersion}" = "v2.10.0" ]; then
    pier --repo /root/.pier appchain service register \
    --appchain-id $appchainId \
    --service-id $serviceId\
    --name $name \
    --intro "" \
    --type CallContract \
    --permit "" \
    --details "test" \
    --reason "reason" \
    --ordered "1"

fi
