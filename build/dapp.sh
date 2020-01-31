#!/bin/bash
set -Eeuxo pipefail

TARGET="$(printenv TARGET || true)" # linux, windows, darwin
if [[ "${TARGET}" == "" ]]; then
    TARGET=$(go env GOOS)
fi

echo "Building for target "${TARGET}"..."
echo "Bundling electron..."

 ./build/run-in-docker.sh builder "
        # make sure electron bundler is synchronized
        go install github.com/asticode/go-astilectron-bundler/astilectron-bundler
        # attach bindata with js code
        go get -u github.com/asticode/go-bindata/...
        go generate ./dapp
        cd dapp && astilectron-bundler -c bundler_"${TARGET}".json
     "
