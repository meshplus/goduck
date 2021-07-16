APP_NAME = goduck
APP_VERSION = 1.0.0

# build with version infos
VERSION_DIR = github.com/meshplus/${APP_NAME}
BUILD_DATE = $(shell date +%FT%T)
GIT_COMMIT = $(shell git log --pretty=format:'%h' -n 1)
GIT_BRANCH = $(shell git rev-parse --abbrev-ref HEAD)

GO_LDFLAGS += -X "${VERSION_DIR}.BuildDate=${BUILD_DATE}"
GO_LDFLAGS += -X "${VERSION_DIR}.CurrentCommit=${GIT_COMMIT}"
GO_LDFLAGS += -X "${VERSION_DIR}.CurrentBranch=${GIT_BRANCH}"
GO_LDFLAGS += -X "${VERSION_DIR}.CurrentVersion=${APP_VERSION}"

TEST_PKGS := $(shell go list ./... | grep -v 'mock_*' | grep -v 'tester')

RED=\033[0;31m
GREEN=\033[0;32m
BLUE=\033[0;34m
NC=\033[0m

GO = go

help: Makefile
	@printf "${BLUE}Choose a command run:${NC}\n"
	@sed -n 's/^##//p' $< | column -t -s ':' | sed -e 's/^/    /'

## make prepare: Preparation before development
prepare:
	@cd scripts && bash prepare.sh

## make test: Run go unittest
test:
	$(GO) generate ./...
	@$(GO) test ${TEST_PKGS} -count=1

## make test-coverage: Test project with cover
test-coverage:
	$(GO) generate ./...
	@$(GO) test -short -coverprofile cover.out -covermode=atomic ${TEST_PKGS}
	@cat cover.out >> coverage.txt

## make install: Go install the project
install:
	cd cmd/goduck && packr
	$(GO) install -ldflags '${GO_LDFLAGS}' ./cmd/${APP_NAME}
	@printf "${GREEN}Build ${APP_NAME} successfully!${NC}\n"

build:
	cd cmd/goduck && packr
	@mkdir -p bin
	$(GO) build -ldflags '${GO_LDFLAGS}' ./cmd/${APP_NAME}
	@mv ./${APP_NAME} bin
	@printf "${GREEN}Build ${APP_NAME} successfully!${NC}\n"

## make linter: Run golanci-lint
linter:
	golangci-lint run
