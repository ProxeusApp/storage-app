# Proxeus Storage App

This repository contains the code for the Proxeus Storage App.

## Documentation / Architectural Overview
For more info check the [documentation](docs/overview.md).

## Getting Started

### Prerequisites
+ make
+ go (1.13+, 64bit for Windows)
+ GOBIN added to your PATH (to check your GOBIN: `echo $(go env GOPATH)/bin`)
+ curl
+ yarn (1.12.3+)
+ node (8.11.3+)
+ vue-cli
+ docker

Command:
If you run linux run.
```
sudo apt-get install make golang curl npm docker docker-compose
```

### Get repository

Get repository:
```
git clone git@github.com:ProxeusApp/storage-app.git ~/workspace/storage-app
```

Change into directory:
```
cd ~/workspace/storage-app
```

Windows:
Run **as administrator**
`npm install -g --production windows-build-tools`

OSX:
Some commands will require brew:
`/usr/bin/ruby -e "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/master/install)"`

### Building
Before starting make sure you have the above listed prerequisites installed 
and that your Go Bin directory has been added to your PATH

On Linux to install golang you can also simply run:
```
make go-init-linux
``` 

First to initialize dependencies run (sudo might be needed):
```
make init
```

To build the app with all dependencies execute this command. It will build the app, pgp-, and spp-server:
```
make all
```

### Building and running servers locally

In the `docker-compose.yml` set the variables for `ETHCLIENTURL` and `ETHWEBSOCKETURL.
You can get one here https://infura.io

Start spp and pgp-server (sudo might be needed)
```
docker-compose up spp pgp
```

To run the app in devMode add the "devMode"-flag:
E.g. on OSX:
```
ETHCLIENTURL=https://ropsten.infura.io/v3/YOURAPIKEY ETHWEBSOCKETURL=wss://ropsten.infura.io/ws/v3/YOURAPIKEY ./artifacts/dapp/darwin-amd64/Proxeus.app/Contents/MacOS/Proxeus --devMode
```
