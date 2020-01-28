# required tooling
init:
	./build/deps.sh

dapp: dapp-ui
	./build/dapp.sh

dapp-all-platforms: dapp-ui dapp-all-platforms-go-only

dapp-all-platforms-go-only:
	TARGET=linux ./build/dapp.sh
	TARGET=windows ./build/dapp.sh
	TARGET=darwin ./build/dapp.sh

dapp-ui:
	make -C ui dapp

spp:
	make -C spp spp

pgp:
	make -C pgp-server pgp

validate:
	./build/validate.sh

validate-ui:
	make -C ui validate

fmt:
	goimports -w -local dapp/api dapp/core pgp-server spp lib
	make -C ui fmt

test:
	./build/test.sh

clean:
	cd artifacts && rm -rf `ls . | grep -v 'cache'`

all: spp pgp dapp


	ln -s /core/central /go/src/git.proxeus.com/core/central

.PHONY: init pgp spp main all all-debug generate test clean fmt validate link-repo
.PHONY: dapp dapp-all-platforms dapp-all-platforms-go-only dapp-ui
.PHONY: main-hosted main-hosted-go-only main-hosted-ui
