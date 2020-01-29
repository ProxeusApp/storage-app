#!/bin/bash
set -Eeuo pipefail

go test github.com/ProxeusApp/storage-app/dapp/... github.com/ProxeusApp/storage-app/pgp-server/... github.com/ProxeusApp/storage-app/spp/...
