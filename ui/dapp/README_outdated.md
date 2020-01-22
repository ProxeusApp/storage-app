## Dapp Information ##

To see wallet, on browser console write:
this.web3.eth.accounts.wallet

## Mostly outdated from here on forward ##
See README in root directory

# Native Proxeus
is the main repository for building the system for operating systems like Windows, Linux or Mac.

## **Prerequisite**
> <a href="https://git-scm.com/downloads">install</a> and <a href="https://git.proxeus.com/snippets/1">configure</a> git

## **Build**
>#### get/update dependencies
>```sh
>$ go get -v -u github.com/asticode/go-astilectron-bundler/...
>$ go get -v -u git.proxeus.com/core/central/dapp/...
>$ sudo npm install -g yarn
>.../dapp yarn install
>```
>#### Build assets
The assets are going to be embedded in the executable on the next step.
>```sh
>.../dapp yarn run build
>.../dapp go run bindata_builder/main.go
>```
>#### Bundle and build the native Proxeus
>```sh
>.../dapp astilectron-bundler -v
>```
If you get the following error message

`bindata.go:1:1: expected 'package', found 'EOF'`

remove `.../dapp/bindata.go` file

Now checkout the binary under `.../dapp/output/*`
For example to run on linux do: ./output/linux-amd64/Proxeus

## **Web Server**
#### Run server without embedded browser

`go run api/main/service.go`

#### Available flag

- **serverAddress** - to change the host and port where the server will run (default is ':8081')

# Frontend

# blockfile

> A Vue.js project

## Build Setup

``` bash
# install dependencies
npm install

# serve with hot reload at localhost:8080
npm run dev

# build for production with minification
npm run build

# build for production and view the bundle analyzer report
npm run build --report

# run unit tests
npm run unit

# run e2e tests
npm run e2e

# run all tests
npm test
```

For a detailed explanation on how things work, check out the [guide](http://vuejs-templates.github.io/webpack/) and [docs for vue-loader](http://vuejs.github.io/vue-loader).

# **Known issues**

# **Notes**

- PSPP Allowance should be bigger than 1 XES Token
- Use npm link for local wallet development https://medium.com/@alexishevia/the-magic-behind-npm-link-d94dcb3a81af
