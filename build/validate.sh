#!/bin/bash
set -Eeuo pipefail

# is go mod in sync
go mod verify

installed () {
    which $1
}
require () {
    if ! installed $1; then echo "Please install $1"; exit 1; fi
}
require goimports
require gofmt

if [[ "$(goimports -l -local dapp/api dapp/core pgp-server spp \
 | grep -v bindata.go \
 | wc -l)" -ne "0" ]]
then
    echo "code not formatted, run make fmt to fix"; exit 1;
fi

# formatting

#gofiles=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$')
#[ -z "$gofiles" ] && exit 0

go vet github.com/ProxeusApp/storage-app/dapp/... github.com/ProxeusApp/storage-app/spp/... github.com/ProxeusApp/storage-app/pgp-server/... github.com/ProxeusApp/storage-app/lib/... github.com/ProxeusApp/storage-app/web/...

gofiles=$(find . -regex "^.*\.go$" \
    | grep -v "^./artifacts/" \
    | grep -v "^./dapp/bind_" \
    | grep -v "/bindata.go$" \
)

unformatted=$(gofmt -l ${gofiles})
[[ -z "${unformatted}" ]] && exit 0

echo >&2 "Go files must be formatted with gofmt. Please run:"
for fn in ${unformatted}; do
    echo >&2 "  gofmt -w $PWD/$fn"
done

exit 1
