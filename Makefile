# required tooling
init:
	./build/deps.sh

go-init-linux:
	./build/go.sh

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
	goimports -w -local dapp/api dapp/core pgp-server spp lib web
	make -C ui fmt

test:
	go test github.com/ProxeusApp/storage-app/dapp/... github.com/ProxeusApp/storage-app/pgp-server/... github.com/ProxeusApp/storage-app/spp/...

test-api-dapp:
	go build -o ./dapp-service dapp/api/main/service.go
	TESTMODE=true docker-compose up -d --build spp pgp
	TESTMODE=true ./dapp-service &
	STORAGE_APP_URL=http://localhost:8081 go test -count=1 -v ./test
	pkill -f dapp-service
	docker-compose down
	rm dapp-service
	rm -r ~/.proxeus-data-api-test

clean:
	cd artifacts && rm -rf `ls . | grep -v 'cache'`

all: spp pgp dapp

.PHONY: init pgp spp main all all-debug generate test clean fmt validate link-repo
.PHONY: dapp dapp-all-platforms dapp-all-platforms-go-only dapp-ui
.PHONY: main-hosted main-hosted-go-only main-hosted-ui
.PHONY: test-api-dapp
