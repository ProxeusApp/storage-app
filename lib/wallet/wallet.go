package wallet

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ProxeusApp/storage-app/dapp/core/ethglue"
)

// Keystore is a fixed keystore bases on the ropsten server wallet TODO have different keystores for different environments
const Keystore = `{
	"version": 3,
	"id": "83e1b2ca-e9c0-4556-88f5-78636306ab0c",
	"address": "51ec4a3608614623c102d4136152b4ed0d522c98",
	"crypto": {
		"ciphertext": "efad565aacb40cc458d5ae08e04ccca8d53ae4285fcd949fa3f0fd0c9d67e76d",
		"cipherparams": {
			"iv": "06d0bcc01f261ed72943e31f146e6602"
		},
		"cipher": "aes-128-ctr",
		"kdf": "scrypt",
		"kdfparams": {
			"dklen": 32,
			"salt": "104558d7c106da5a9a069811955130c8a7f19e48ff7b5dcda586f614d0667c44",
			"n": 8192,
			"r": 8,
			"p": 1
		},
		"mac": "74cd4b34f82ac118247342a114c8f4aac9a4ac786b3f4e23e76568768b325012"
	}
}`

// ErrInvalidSignature is special error returned when signature verification fails
var ErrInvalidSignature = errors.New("login.error.invalidSignature")

// CreateSignInChallenge returns an hex string representation of a message to be used for login challenge.
// The challenge is prefixed by a human-readable message so that it can be displayed on Metamask.
// The challenge itself is an hex string of 32 random bytes.
func CreateSignInChallenge(i18nMessage string) string {
	// generate array from random 32 bytes
	challenge := make([]byte, 32)
	rand.Read(challenge)
	challengeHex := "0x" + hex.EncodeToString(challenge)

	result := append([]byte(i18nMessage), []byte(challengeHex)...)

	return "0x" + hex.EncodeToString(result)
}

// VerifySignInChallenge verifies if the given signature matches the provided challenge and returns the address
// of the wallet that made the signature.
func VerifySignInChallenge(challengeHex, signatureHex string) (addressHex string, err error) {
	minimumChallengeSize := 2
	if len(challengeHex) < minimumChallengeSize {
		// challenge is stored in memory, so it can happed that it will be empty after server restart
		err = fmt.Errorf("Wrong challenge size: Expected more than %d characters, but %d found", minimumChallengeSize, len(challengeHex))
		return
	}
	if len(signatureHex) != 132 {
		err = fmt.Errorf("Wrong signature size: Expected 132 characters, but %d found", len(signatureHex))
		return
	}

	// for some reason hex decode wants an hex string with no hex prefix("0x"), hence [2:]
	signature, err := hex.DecodeString(signatureHex[2:])
	if err != nil {
		return
	}

	// need to subtract 27 from V, reason here: https://github.com/ethereum/wiki/wiki/JavaScript-API#returns-45
	signature[64] -= 27

	// get the hash of the challenge
	challenge, err := hex.DecodeString(challengeHex[2:])
	if err != nil {
		return
	}

	// pre-pend the eth_sign RPC prefix https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_sign
	challenge = append([]byte("\x19Ethereum Signed Message:\n"+strconv.Itoa(len(challenge))), challenge...)
	challengeHash := crypto.Keccak256(challenge)

	// get the public key from the signature and challenge hash
	pubKeyECDSA, err := crypto.SigToPub(challengeHash, signature)

	// if any of the decodes raised errors, return it
	if err != nil {
		return
	}
	pubKey := crypto.FromECDSAPub(pubKeyECDSA)
	addressHex = crypto.PubkeyToAddress(*pubKeyECDSA).String()

	// build the byte array with R and S
	signatureRS := signature[0:64]

	// verify the signature and return result.
	// TODO check if this verify signature is really needed. Since we are obtaining the public key from the signature, it will always verify to true.
	if !crypto.VerifySignature(pubKey, challengeHash, signatureRS) {
		err = ErrInvalidSignature
	}
	return
}

// RegisterDocument executes a transaction on the blockchain to call the "notarize" method on a DocumentRegistry
// smart contract at the address provided. An approval of XES Token is done at the provided token address to the
// address of the provided document registry.
// The transaction hash is returned.
func RegisterDocument(ethClientURL, xesTokenAddress, documentRegistryAddress, documentHashHex string) (txHash string, err error) {
	if len(xesTokenAddress) != 42 {
		err = fmt.Errorf("wrong XES token address size: Expected 42 characters, but %d found", len(xesTokenAddress))
		return
	}
	if len(documentRegistryAddress) != 42 {
		err = fmt.Errorf("wrong document registry address size: Expected 42 characters, but %d found", len(documentRegistryAddress))
		return
	}
	if len(documentHashHex) != 66 {
		err = fmt.Errorf("wrong document hash size: Expected 66 characters, but %d found", len(documentHashHex))
		return
	}

	// Connect to infura TODO have the network come from some definition
	conn, err := ethglue.Dial(ethClientURL)
	if err != nil {
		err = fmt.Errorf("failed to connect to the Ethereum client: %v", err)
		return
	}

	// Instantiate the contracts
	documentRegistry, err := NewDocumentRegistry(common.HexToAddress(documentRegistryAddress), conn)
	if err != nil {
		err = fmt.Errorf("failed to instantiate a DocumentRegistry contract: %v", err)
		return
	}

	xesMainToken, err := NewXesMainToken(common.HexToAddress(xesTokenAddress), conn)
	if err != nil {
		err = fmt.Errorf("failed to instantiate a XesMainToken contract: %v", err)
		return
	}

	// Create an authorized transactor TODO have the password come from config
	auth, err := bind.NewTransactor(strings.NewReader(Keystore), "Password")
	if err != nil {
		err = fmt.Errorf("failed to create authorized transactor: %v", err)
		return
	}

	// TODO have the gas limit and price come from some config file, find an appropriate gas price
	auth.GasLimit = 4500000
	auth.GasPrice = big.NewInt(50000000000) // 50 Gwei

	documentHashBytes, err := hex.DecodeString(documentHashHex[2:])
	if err != nil {
		err = fmt.Errorf("failed to decode document hash to bytes: %v", err)
		return
	}

	// convert the documentHash from hex to fixed size byte[]
	var documentHashBytesFixed [32]byte
	copy(documentHashBytesFixed[:], documentHashBytes[:32])

	// convert the document registry address to byte array so it can be converted again to address
	documentRegistryAddressBytes, err := hex.DecodeString(documentRegistryAddress[2:])
	if err != nil {
		err = fmt.Errorf("failed to decode document registry address to bytes: %v", err)
		return
	}

	// no issues up to this point, approve the XES payment
	approvalTx, err := xesMainToken.Approve(auth, common.BytesToAddress(documentRegistryAddressBytes), big.NewInt(1))
	if err != nil {
		err = fmt.Errorf("failed to request token approval: %v", err)
		return
	}

	// wait for the approval to be mined
	bind.WaitMined(context.Background(), conn, approvalTx)

	registrationTx, err := documentRegistry.NotarizeDocument(auth, documentHashBytesFixed)
	if err != nil {
		err = fmt.Errorf("failed to request document registration: %v", err)
		return
	}

	// convert the returned transaction hash object into an hex string
	txHash = registrationTx.Hash().String()

	return
}
