#!/bin/sh

proposalID=$1
voteOption=$2
voteReason=$3

bitxhub --repo /root/.bitxhub/ client  governance vote --id $1 --info $2 --reason $3
#--gateway "http://172.19.0.1:9091/v1/"
