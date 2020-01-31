#!/bin/bash
set -Eeuxo pipefail

installed () {
    which $1
}

# for linux install go
if installed apt-get && ! installed go; then
    apt-get install curl;
    # install go
    curl https://dl.google.com/go/go1.13.4.linux-amd64.tar.gz > go.tar.gz
    tar -xf go.tar.gz
    rm go.tar.gz
    rm -Rf /usr/local/go
    mv go /usr/local
    export PATH=/usr/local/go/bin:$PATH
    fi
fi