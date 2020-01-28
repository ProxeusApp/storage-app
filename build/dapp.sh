#!/bin/bash
set -Eeuxo pipefail

TARGET="$(printenv TARGET || true)" # linux, windows, darwin
if [[ "${TARGET}" == "" ]]; then
    TARGET=$(go env GOOS)
fi

echo "Building for target "${TARGET}"..."

# attach bindata with js code
go generate ./dapp

echo "Bundling electron..."

 ./build/run-in-docker.sh builder "
        # make sure electron bundler is synchronized
        go install ./vendor/github.com/asticode/go-astilectron-bundler/astilectron-bundler
        cd dapp && GOCACHE=$(go env GOCACHE) astilectron-bundler -c bundler_"${TARGET}".json -v
     "
