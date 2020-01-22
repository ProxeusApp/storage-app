spp:
	./deplo.sh spp

pgp:
	./deplo.sh pgp

main-hosted:
	./deplo.sh main-hosted

all: spp pgp main-hosted

# prepares self-contained folder for rsync transfer
prepare:
	rm -rf				../artifacts/deploy
	mkdir -p			../artifacts/deploy/current-services
	cp -r -L ../deploy	../artifacts/deploy/current-services/deploy

	mkdir				../artifacts/deploy/current-services/artifacts
	cp -r main-hosted	../artifacts/deploy/current-services/artifacts
	cp -r pgp			../artifacts/deploy/current-services/artifacts
	cp -r spp			../artifacts/deploy/current-services/artifacts

	cp ../artifacts/main-hosted/server	../artifacts/deploy/current-services/artifacts/main-hosted/server
	mkdir ../artifacts/deploy/current-services/artifacts/main-hosted/connectors
	cp ../artifacts/main-hosted/connectors/http-plug	../artifacts/deploy/current-services/artifacts/main-hosted/connectors/http-plug

	cp ../artifacts/pgp/server			../artifacts/deploy/current-services/artifacts/pgp/server
	cp ../artifacts/spp/server			../artifacts/deploy/current-services/artifacts/spp/server

.PHONY: all spp pgp main-hosted setup