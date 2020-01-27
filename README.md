# Proxeus Storage App

This repository contains the code for the Proxeus Storage App.

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
```
sudo apt-get install make golang curl npm docker docker-compose
```

## Getting Started

Get repository:
```
git clone git@github.com:ProxeusApp/storage-app.git $(go env GOPATH)/src/git.proxeus.com/core/central
```

Change into directory:
```
cd $(go env GOPATH)/src/git.proxeus.com/core/central
```

On Windows you need to run, **as administrator**
`npm install -g --production windows-build-tools`

### Building
First to initialize dependencies run (sudo might be needed):
```
make init
```

To build the app with all dependencies execute this command. It will build the app, pgp-, and spp-server:
```
make all
```

### Building and running servers locally

To run the app in devMode execute:
```
./artifacts/dapp/darwin-amd64/Proxeus.app/Contents/MacOS/Proxeus --devMode
```
