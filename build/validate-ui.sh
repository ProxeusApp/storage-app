#!/bin/bash

# validate dapp
cd dapp

yarn run lint:js
if [[ $? != "0" ]]
then
    echo "js code not formatted or has errors, run make fmt to show errors and fix autofixable errors"
    exit 1
fi

yarn run lint:css
if [[ $? != "0" ]]
then
    echo "css code not formatted or has errors, run make fmt to show errors and fix autofixable errors"
    exit 1
fi
# end: validate dapp

exit 0
