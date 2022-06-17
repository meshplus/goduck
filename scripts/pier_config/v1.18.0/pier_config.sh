#!/usr/bin/env bash

set -e

SYSTEM=$(uname -s)
if [ $SYSTEM == "Linux" ]; then
  SYSTEM="linux"
elif [ $SYSTEM == "Darwin" ]; then
  SYSTEM="darwin"
fi

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

function print_blue() {
  printf "${BLUE}%s${NC}\n" "$1"
}

function print_green() {
  printf "${GREEN}%s${NC}\n" "$1"
}

function print_red() {
  printf "${RED}%s${NC}\n" "$1"
}

# The sed commend with system judging
# Examples:
# sed -i 's/a/b/g' bob.txt => x_replace 's/a/b/g' bob.txt
function x_replace() {
  system=$(uname)

  if [ "${system}" = "Linux" ]; then
    sed -i "$@"
  else
    sed -i '' "$@"
  fi
}

printHelp() {
  awk -F'### ' '/^###/ { print $2 }' "$0"
}

function InitConfig() {
  readConfig
  generatePierConfig
  rewritePierConfig
}

function readConfig() {
  REWRITE=`sed '/^.*rewrite/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`

  HTTPPORT=`sed '/^.*httpPort/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
  PPROFPORT=`sed '/^.*pprofPort/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
  APIPORT=`sed '/^.*apiPort/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
  TLS=`sed '/^.*tls/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
  MODE=`sed '/^.*mode/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`

  case $MODE in
  relay)
    BITXHUBADDR=`sed '/^.*bitxhubAddr/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
    VALIDATORS=`sed '/^.*bitxhubValidators/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
    ;;
  direct)
    DIRECTPORT=`sed '/^.*directPort/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
    PEERSADDR=`sed '/^.*peersAddr/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
    ;;
  union)
    CONNECTORS=`sed '/^.*connectors/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
    PROVIDERS=`sed '/^.*providers/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
    ;;
  esac

  case $APPCHAINTYPE in
  ethereum)
    CONTRACTADDR=`sed '/^.*contractAddr/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
    ETHADDR=`sed '/^.*ethAddr/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
    ;;
  fabric)
    CRYPTOPATH=`sed '/^.*cryptoPath/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
    FABRICIP=`sed '/^.*fabricIP/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
    FABIRCP1=`sed '/^.*ordererP/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
    FABIRCP2=`sed '/^.*urlSubstitutionExpP1/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
    FABIRCP3=`sed '/^.*eventUrlSubstitutionExpP1/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
    FABIRCP4=`sed '/^.*urlSubstitutionExpP2/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
    FABIRCP5=`sed '/^.*eventUrlSubstitutionExpP2/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
    FABIRCP6=`sed '/^.*urlSubstitutionExpP3/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
    FABIRCP7=`sed '/^.*eventUrlSubstitutionExpP3/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
    FABIRCP8=`sed '/^.*urlSubstitutionExpP4/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
    FABIRCP9=`sed '/^.*eventUrlSubstitutionExpP4/!d;s/.*=//;s/[[:space:]]//g' ${CONFIGPATH}`
    ;;
  esac


}

function generatePierConfig() {
  print_blue "======> Generate configuration files for pier_$APPCHAINTYPE"

  if [ -f ${TARGET}/pier.toml ]; then
    print_blue "whether rewrite the configuration files already exist"
      if [ $REWRITE == "false" ]; then
        print_blue "not rewrite configuration files"
        exit 1
      else
        print_blue "rewrite configuration files"
        rm -r ${TARGET}/*
      fi
  fi

  print_blue "【1】generate configuration files"
  if [ "${SYSTEM}" == "linux" ]; then
    export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:"${PIERBINPATH}"/
  elif [ "${SYSTEM}" == "darwin" ]; then
    install_name_tool -change @rpath/libwasmer.dylib "${PIERBINPATH}"/libwasmer.dylib "${PIERBINPATH}"/pier
  else
    print_red "Pier does not support the current operating system"
  fi
  ${PIERBINPATH}/pier --repo ${TARGET} init $MODE

  print_blue "【2】copy pier plugin and appchain config"
  mkdir ${TARGET}/plugins
  if [ $APPCHAINTYPE == "ethereum" ]; then
    cp ${PLUGINPATH}/ethereum-client ${TARGET}/plugins/appchain_plugin
  elif [ $APPCHAINTYPE == "fabric" ]; then
    cp ${PLUGINPATH}/fabric-client ${TARGET}/plugins/appchain_plugin
  else
    print_red "Not supported mode"
  fi
  cp -r ${APPCHAINCONFIGPATH} ${TARGET}/${APPCHAINTYPE}
}

function rewritePierConfig() {
  print_blue "======> Rewrite config for pier_$APPCHAINTYPE"

  print_blue "【1】rewrite pier.toml"
  # port
  x_replace "s/http.*= .*/http = $HTTPPORT/" ${TARGET}/pier.toml
  x_replace "s/pprof.*= .*/pprof = $PPROFPORT/" ${TARGET}/pier.toml
  # type
  x_replace "s/type.*= \".*\"/type = \"$MODE\"/" ${TARGET}/pier.toml
  case $MODE in
  relay)
    x_replace "s/addrs.*=.*/addrs = [\"$BITXHUBADDR\"]/" ${TARGET}/pier.toml
    x_replace "s/validators.*= .*/validators = $VALIDATORS/" ${TARGET}/pier.toml
    ;;
  direct)
    PID=`${PIERBINPATH}/pier --repo ${TARGET} p2p id`
    SELFADDR="\"/ip4/127.0.0.1/tcp/$DIRECTPORT/p2p/$PID\""
    PEERSADDR1=${PEERSADDR/\[/\[$SELFADDR,}
    x_replace "s/peers.*= .*/peers = $PEERSADDR1/" ${TARGET}/pier.toml
    ;;
  union)
    x_replace "s/connectors.*= .*/connectors = $CONNECTORS/" ${TARGET}/pier.toml
    x_replace "s/providers.*= .*/providers = $PROVIDERS/" ${TARGET}/pier.toml
    ;;
  esac
  # tls
  x_replace "s/enable_tls.*= .*/enable_tls = $TLS/" ${TARGET}/pier.toml
  # appchain
  x_replace "s/config.*= \".*\"/config = \"$APPCHAINTYPE\"/" ${TARGET}/pier.toml

  print_blue "【2】rewrite appchain config"
  if [ $APPCHAINTYPE == "ethereum" ]; then
    ETHADDRTMP=$(echo "${ETHADDR}" | sed 's/\//\\\//g')
    x_replace "s/{{.AppchainAddr}}/$ETHADDRTMP/" ${TARGET}/$APPCHAINTYPE/$APPCHAINTYPE.toml
    x_replace "s/{{.AppchainContractAddr}}/$CONTRACTADDR/" ${TARGET}/$APPCHAINTYPE/$APPCHAINTYPE.toml
  elif [ $APPCHAINTYPE == "fabric" ]; then
    FABRICADDR="$FABRICIP:$FABIRCP3"
    CRYPTOPATH_TMP=$(echo $CRYPTOPATH | sed 's/\//\\\//g')
    x_replace "s/{{.AppchainAddr}}/$FABRICADDR/" ${TARGET}/$APPCHAINTYPE/$APPCHAINTYPE.toml
    x_replace "s/{{.ConfigPath}}/$CRYPTOPATH_TMP/" ${TARGET}/$APPCHAINTYPE/config.yaml
    x_replace "s/{{.AppchainIP}}/$FABRICIP/" ${TARGET}/$APPCHAINTYPE/config.yaml
    x_replace "s/{{.Port1}}/$FABIRCP1/" ${TARGET}/$APPCHAINTYPE/config.yaml
    x_replace "s/{{.Port2}}/$FABIRCP2/" ${TARGET}/$APPCHAINTYPE/config.yaml
    x_replace "s/{{.Port3}}/$FABIRCP3/" ${TARGET}/$APPCHAINTYPE/config.yaml
    x_replace "s/{{.Port4}}/$FABIRCP4/" ${TARGET}/$APPCHAINTYPE/config.yaml
    x_replace "s/{{.Port5}}/$FABIRCP5/" ${TARGET}/$APPCHAINTYPE/config.yaml
    x_replace "s/{{.Port6}}/$FABIRCP6/" ${TARGET}/$APPCHAINTYPE/config.yaml
    x_replace "s/{{.Port7}}/$FABIRCP7/" ${TARGET}/$APPCHAINTYPE/config.yaml
    x_replace "s/{{.Port8}}/$FABIRCP8/" ${TARGET}/$APPCHAINTYPE/config.yaml
    x_replace "s/{{.Port9}}/$FABIRCP9/" ${TARGET}/$APPCHAINTYPE/config.yaml
  else
    print_red "Not supported mode"
  fi

  print_blue "【3】rewrite api"
  x_replace "s/{{.ApiPort}}/$APIPORT/" ${TARGET}/api
}

# parses opts
# PIERBINPATH: PIER binary and plug-in binary are in the directory
while getopts "h?p:b:c:a:g:f:" opt; do
  case "$opt" in
  h | \?)
    printHelp
    exit 0
    ;;
  p)
    TARGET=$OPTARG
    ;;
  b)
    PIERBINPATH=$OPTARG
    ;;
  c)
    CONFIGPATH=$OPTARG
    ;;
  a)
    APPCHAINTYPE=$OPTARG
    ;;
  g)
    PLUGINPATH=$OPTARG
    ;;
  f)
    APPCHAINCONFIGPATH=$OPTARG
    ;;
  esac
done

InitConfig

# 1.11.2 delete method