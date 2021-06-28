#!/bin/sh

proposalID=$1
voteOption=$2
voteReason=$3
# index=$4

# indexTmp=`expr $index + 1`
bitxhub --repo /root/.bitxhub/ client --gateway "http://172.19.0.1:9091/v1/" governance vote --id $1 --info $2 --reason $3
