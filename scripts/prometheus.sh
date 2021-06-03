#!/usr/bin/env bash

set -e
source x.sh

OPT=$1
ADDR1=$2
ADDR2=$3
ADDR3=$4
ADDR4=$5
GRAFANA_HOST=127.0.0.1
CURRENT_PATH=$(pwd)
PROM_PATH="${CURRENT_PATH}/docker/prometheus"

function printHelp() {
  print_blue "Usage:  "
  echo "  prometheus.sh <mode>"
  echo "    <OPT> - one of 'up', 'down', 'restart'"
  echo "      - 'up' - start prometheus"
  echo "      - 'down' - stop prometheus"
  echo "      - 'restart' - restart prometheus"
  echo "  prometheus.sh -h (print this message)"
}

function prometheus_up() {
  print_blue "====> Start prometheus to monitoring bitxhub"
  echo "bitxhub info: [$ADDR1] [$ADDR2] [$ADDR3] [$ADDR4]"
  prometheusConfig=$PROM_PATH/prometheus.yml
  x_replace "s/host.docker.internal:40011/$ADDR1/g" "${prometheusConfig}"
  x_replace "s/host.docker.internal:40012/$ADDR2/g" "${prometheusConfig}"
  x_replace "s/host.docker.internal:40013/$ADDR3/g" "${prometheusConfig}"
  x_replace "s/host.docker.internal:40014/$ADDR4/g" "${prometheusConfig}"

  print_blue "====> Start prometheus and grafana"
  docker-compose -f $PROM_PATH/docker-prom-compose.yml up -d
  echo "grafana host: $GRAFANA_HOST"

  sleep 5
  print_blue "====> Create datasource"
  curl -X POST \
    http://${GRAFANA_HOST}:3000/api/datasources \
    -H "Content-Type:application/json" \
    -d '{"name":"Prometheus","type":"prometheus","url":"http://prom:9090","access":"proxy", "isDefault":true}' \
    2>${PROM_PATH}/datasources2.log 1>${PROM_PATH}/datasources1.log

  print_blue "====> Create host dashboard"
  curl -X POST \
    http://${GRAFANA_HOST}:3000/api/dashboards/db \
    -H 'Accept: application/json' \
    -H 'Content-Type: application/json' \
    -H 'cache-control: no-cache' \
    -d @$PROM_PATH/Go_Processes.json \
    2>${PROM_PATH}/dashboards2.log 1>${PROM_PATH}/dashboards1.log

  echo ""
  prometheus_check
  print_green "Start prometheus successful!"
  print_green "You can access to \"http://${GRAFANA_HOST}:3000/d/HaYqdcgGk/go-processes\" to get prometheus information."
}

function prometheus_check() {
  if [[ $(cat $PROM_PATH/datasources1.log | grep '"message":"Datasource added"') ]]; then
    print_green "===> Add Create dashboards to grafana successfully!!!"
  else
    echo ""
    cat $PROM_PATH/datasources2.log
    cat $PROM_PATH/datasources1.log
    echo ""
    print_red "===> Fail to add datasource!!!"
    exit 0
  fi

  if [[ $(cat $PROM_PATH/dashboards1.log | grep '"status":"success"') ]]; then
    print_green "===> Create dashboards of grafana successfully!!!"
  else
    echo ""
    cat $PROM_PATH/dashboards2.log
    cat $PROM_PATH/dashboards1.log
    echo ""
    print_red "===> Fail to create dashboards!!!"
    exit 0
  fi
}

function prometheus_down() {
  print_blue "====> Stop prometheus"
  docker-compose -f $PROM_PATH/docker-prom-compose.yml down
}

function prometheus_restart() {
  prometheus_down
  prometheus_up
}

if [ "$OPT" == "up" ]; then
  prometheus_up
elif [ "$OPT" == "down" ]; then
  prometheus_down
elif [ "$OPT" == "restart" ]; then
  prometheus_restart
else
  printHelp
  exit 1
fi
