
.PHONY: check-node-version init serve-main-hosted serve-dapp all main-hosted dapp

init:
	yarn install --pure-lockfile --cache-folder .yarn

check-node-version:
	yarn run check-node-version

validate: init check-node-version
	../build/validate-ui.sh

serve-main-hosted: check-node-version
	cd core && yarn run serve

serve-dapp: check-node-version
	cd dapp && yarn run serve

fmt: fmt-dapp

fmt-dapp:
	cd dapp && yarn run fmt

all: main-hosted dapp

main-hosted: init check-node-version
	cd core && yarn run build

dapp: init check-node-version
	cd dapp && yarn run build
