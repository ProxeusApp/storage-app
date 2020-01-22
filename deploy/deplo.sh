#!/bin/bash
set -Eeuxo pipefail

# command args
target=$1
extraCmd=${2:-}

DIR=$(pwd)
function bring-back-dir {
    cd ${DIR}
}
trap bring-back-dir EXIT
cd "$( dirname "${BASH_SOURCE[0]}" )"
pwd

echo "Deploying ${target}..."

cp ./${target}/Dockerfile ../artifacts/${target}/Dockerfile
cp ./${target}/docker-compose.yaml ../artifacts/${target}/docker-compose.yaml

if [[ ${target} = "main-hosted" ]]
then
    cp ./${target}/.env ../artifacts/${target}/.env
fi

if [[ ${target} = "spp" ]]
then
    cp ./${target}/settings.json ../artifacts/${target}/settings.json
fi

if [[ "${extraCmd}" = "-debug" ]]
then
    cat ${target}/docker-compose.yaml | grep -v "restart: always" \
      > ../artifacts/${target}/docker-compose.yaml
    sed -i 's/document-service:2115/10.131.0.161:2115/g' ../artifacts/${target}/.env
fi

cd ../artifacts/${target}
docker build -t ${target} .
docker-compose up -d

docker ps | grep ${target}
echo "Success!"