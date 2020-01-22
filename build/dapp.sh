#!/bin/bash
set -Eeuxo pipefail

TARGET="$(printenv TARGET || true)" # linux, windows, darwin
if [[ "${TARGET}" == "" ]]; then
    TARGET=$(go env GOOS)
fi

echo "Building for target "${TARGET}"..."

# make sure go-bindata is up to date
go install ./vendor/github.com/asticode/go-bindata/go-bindata
# attach bindata with js code
go generate ./dapp

echo "Bundling electron..."

if [[ "${TARGET}" == "$(go env GOOS)" || "${TARGET}" == "windows" ]]; then
    # make sure electron bundler is synchronized
    go install ./vendor/github.com/asticode/go-astilectron-bundler/astilectron-bundler
    cd dapp && GOCACHE="$(go env GOCACHE)" astilectron-bundler -c bundler_${TARGET}.json -v && cd ..
else # native or cross compilation not possible for this GOOS and selected target, use docker
       ./build/run-in-docker.sh builder "
        # make sure electron bundler is synchronized
        go install ./vendor/github.com/asticode/go-astilectron-bundler/astilectron-bundler
        cd dapp && astilectron-bundler -c bundler_"${TARGET}".json -v
     "
fi
