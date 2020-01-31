#!/bin/bash
set -Eeuxo pipefail

installed () {
    which $1
}

require () {
    if ! installed $1; then echo "Please manually install $1"; exit 1; fi
}

#required for git.proxeus.com/web/channelhub dependency
export GOPRIVATE="git.proxeus.com"

# for linux install npm and curl
if installed apt-get; then
    require curl
    # install node
    if ! installed node; then
      curl -sL https://deb.nodesource.com/setup_11.x | sudo -E bash -
      apt-get install -y nodejs

      if ! installed yarn; then
        curl -sS https://dl.yarnpkg.com/debian/pubkey.gpg | sudo apt-key add -
        echo "deb https://dl.yarnpkg.com/debian/ stable main" | sudo tee /etc/apt/sources.list.d/yarn.list
        apt-get update && sudo apt-get install yarn
      fi
    fi

    apt-get install -y libgconf-2-4 #on some linux distributions missing package
fi

require go
require npm
require yarn

# cross-compiler for windows cgo and osx builder

# for linux use apt-get
if installed apt-get; then
    apt-get install mingw-w64 docker.io
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
