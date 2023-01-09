#!/usr/bin/env bash

BLUE='\033[0;34m'
NC='\033[0m'

function print_blue() {
  printf "${BLUE}%s${NC}\n" "$1"
}

function go_install() {
  version=$(go env GOVERSION)
  if [[ ! "$version" < "go1.16" ]];then
      go install "$@"
  else
      go get "$@"
  fi
}

print_blue "===> 1. Install packr2"
if ! type packr2 >/dev/null 2>&1; then
  go_install github.com/gobuffalo/packr/v2/packr2@v2.8.3
fi