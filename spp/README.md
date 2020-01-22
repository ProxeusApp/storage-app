# Storage Provider Proxy

## **Prerequisite**
>```sh
>make init
>```

## **API**
- **GET /challenge**: Auth over an Ethereum account
- **POST /:fileHash/:token/:signature**: Upload a specific file. Signature of the challenge needs to be provided. Access is granted if the address has write permission on the Smart-Contract provided docHash
- **GET /:fileHash/:token/:signature**: Download a specific file. Signature of the challenge needs to be provided. Access is granted if the address has read permission on the Smart-Contract with the provided docHash.
- **GET /info**: Returns Storage Provider's info
- **GET /ping**: Returns "pong" if service running
- **GET /health**: Returns a list of the dependencies' statuses (for ex. Ethereum Node connection)


## **GO Build**
>```sh
>make spp
>```

## **Create Smart Contract source**
>```sh
>go get -u -v github.com/ethereum/go-ethereum/cmd/abigen
>abigen --abi ./eth/solidity/ProxeusFS.abi --pkg eth --type ProxeusFSContract --out ./eth/proxeusFSContract.go
>```

## Info

A file named `settings.json` should be placed on the same directory than the executable having the following structure:

```bash
{
  "name": "Storage Name",
  "description": "A description",
  "jurisdictionCountry": "Switzerland",
  "dataCenter": "Cham",
  "logoUrl": "https://proxeus.com/wp-content/uploads/2018/07/logo.svg",
  "termsAndConditionsUrl": "",
  "privacyPolicyUrl": "",
  "maxFileSizeByte": 10000,             
  "maxStorageDays": 1000,       
  "graceSeconds": 1000,                // How long is the file kept after expiration. UNUSED!
  "priceByte": "0.001",               // In XESWei
  "priceDay": "0.000005"              // In XESWei
}
```
