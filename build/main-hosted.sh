#!/bin/bash
set -Eeuxo pipefail

DEBUG_FLAG=""
if [[ "$(printenv DEBUG || true)" == "1" ]]
then
    echo "bindata set to DEBUG mode"
    DEBUG_FLAG="-debug"
fi

LDFLAGS=""
if [[ -n "${BUILD_ID:-}" ]]
then
    LDFLAGS="-X main.ServerVersion=build-${BUILD_ID}"
fi

# make sure go-bindata is up to date
go install ./vendor/github.com/asticode/go-bindata/go-bindata

go-bindata ${DEBUG_FLAG} -pkg assets -o ./main/handlers/assets/bindata.go -prefix ./artifacts/main-hosted/dist ./artifacts/main-hosted/dist/...
export CGO_ENABLED=0
go build -ldflags="${LDFLAGS}" -tags nocgo -o ./artifacts/main-hosted/server ./main
# internal connectors
function build-connector {
    go build -ldflags="${LDFLAGS}" -tags nocgo -o ./artifacts/main-hosted/connectors/${1} ./main/connectors/${1}
}
build-connector http-plug
build-connector weather-plug
