#!/usr/bin/env bash

source x.sh

CURRENT_PATH=$(pwd)
MODE=$1

function init_deploy() {

  chmod +x $CURRENT_PATH/deploy_contracts.sh

  print_blue "init ether1 http:8545 ws:8546 port:30303"
  goduck ether start --httpport 8545 --wsport 8546 --port 30303

  print_blue "init ether1 http:8547 ws:8548 port:30304"
  goduck ether start --httpport 8547 --wsport 8548 --port 30304

  $CURRENT_PATH/deploy_contracts.sh deploy http://localhost:8545 http://localhost:8547

}

if [ "$MODE" == "start" ]; then
  init_deploy
fi