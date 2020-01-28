#!/bin/bash
set -Eeuxo pipefail

installed () {
    which $1
}

#required for git.proxeus.com/web/channelhub dependency
export GOPRIVATE="git.proxeus.com"

# for linux install npm and curl
if installed apt-get; then
    apt-get install curl;
    # install go
    curl https://dl.google.com/go/go1.11.2.linux-amd64.tar.gz > go.tar.gz
    tar -xf go.tar.gz
    rm go.tar.gz
    rm -Rf /usr/local/go
    mv go /usr/local
    export PATH=/usr/local/go/bin:$PATH
    echo "----- Please add /usr/local/go/bin to your PATH -----"
    # install node
    curl -sL https://deb.nodesource.com/setup_11.x | sudo -E bash -
    apt-get install -y nodejs
    curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | sudo apt-key add -
    echo "deb https://dl.yarnpkg.com/debian/ stable main" | sudo tee /etc/apt/sources.list.d/yarn.list
    apt-get update && sudo apt-get install yarn
    apt-get install -y libgconf-2-4 #on some linux distributions missing package
fi


require () {
    if ! installed $1; then echo "Please manually install $1"; exit 1; fi
}
require go
require curl
require npm
require yarn

# install golang's dep
mkdir -p $(go env GOPATH)/bin
export PATH=$(go env GOPATH)/bin:$PATH
curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

go get golang.org/x/tools/cmd/goimports

# cross-compiler for windows cgo and osx builder

# for linux use apt-get
if installed apt-get; then
    apt-get install mingw-w64 docker.io
    # TODO: embed osxcross and osx sdk in our infra/repo
fi

docker build --tag=builder build/builder

# on macs use brew
if [[ "$OSTYPE" == "darwin"* ]]; then
  require brew

  if ! brew ls --versions mingw-w64 > /dev/null; then
      brew install mingw-w64
  fi
  clang=$(go env GOPATH)/bin/o64-clang
    if ! [[ -e "$clang" ]]; then
        ln -s /usr/bin/clang "$clang" # symlink to be compatible with dapp/bundler_darwin.json
    fi
fi
