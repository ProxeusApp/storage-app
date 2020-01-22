#!/bin/bash
set -Eeuxo pipefail

# command args
HOST=$1
TARGET=$2

eval $(ssh-agent -s)
set +x
ssh-add <(echo "${SSH_DEPLOY_KEY_DEV}")
set -x
ssh-add -l
make -C deploy prepare
rsync -rhzv --perms --stats ./artifacts/deploy/current-services ${HOST}:/data
ssh ${HOST} << EOF
    ls -lah /data/current-services/deploy
    make -C /data/current-services/deploy ${TARGET}
EOF
# cleanup
ssh-agent -k
