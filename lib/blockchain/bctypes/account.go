package bctypes

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"io"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/crypto"
)

type Account struct {
	privKey ecdsa.PrivateKey
}
type Signature []byte

const EthereumSignatureLength int = 65

func New() (Account, error) {
	acc := Account{}
	key, err := crypto.GenerateKey()
	if err != nil {
		return acc, err
	}
	acc.privKey = *key
	return acc, nil
}

func LoadFromWallet(walletReader io.ReadCloser, password string) (Account, error) {
	acc := Account{}

	fcontent, err := ioutil.ReadAll(walletReader)
	if err != nil {
		return acc, err
	}

	err = acc.loadFromJSON(fcontent, password)
	if err != nil {
		return acc, err
	}

	return acc, nil
}

func (acc *Account) loadFromJSON(keystorejson []byte, password string) error {
	unlockedKey, err := keystore.DecryptKey(keystorejson, password)
	if err != nil {
		return err
	}
	acc.privKey = *unlockedKey.PrivateKey

	return nil
}

func (acc *Account) GetAddress() string {
	return strings.ToLower(crypto.PubkeyToAddress(acc.GetPubKey()).Hex())
}

func (acc *Account) GetPubKey() ecdsa.PublicKey {
	return acc.privKey.PublicKey
}

func (acc *Account) GetTransactor() *bind.TransactOpts {
	return bind.NewKeyedTransactor(&acc.privKey)
}

func (acc *Account) SignMessage(message []byte) ([]byte, error) {
	msg, err := decodeHex(message)
	if err != nil {
		return nil, err
	}

	msg = hashMessage(msg)

	sig, err := crypto.Sign(msg, &acc.privKey)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	sig[EthereumSignatureLength-1] = sig[EthereumSignatureLength-1] + 27 // Yes add 27, weird Ethereum quirk
	signature := Signature(buf.Bytes())
	return signature.format(), err
}

func VerifyMessage(message []byte, signature []byte, key ecdsa.PublicKey) (bool, error) {
	if len(signature) != EthereumSignatureLength {
		return false, os.ErrInvalid
	}
	sig, err := decodeHex(signature)
	if err != nil {
		return false, err
	}
	// need to subtract 27 from V, reason here: https://github.com/ethereum/wiki/wiki/JavaScript-API#returns-45
	sig[EthereumSignatureLength-1] -= 27

	msg, err := decodeHex(message)
	if err != nil {
		return false, err
	}

	msghash := hashMessage(msg)
	// get the public key from the signature and message hash
	pubKeyECDSA, err := crypto.SigToPub(msghash, sig)

	// if any of the decodes raised errors, return it
	if err != nil {
		return false, err
	}
	return bytes.Equal([]byte(crypto.PubkeyToAddress(*pubKeyECDSA).Hex()), []byte(crypto.PubkeyToAddress(key).Hex())), err
}

func hashMessage(message []byte) []byte {
	message = append([]byte("\x19Ethereum Signed Message:\n"+strconv.Itoa(len(message))), message...)
	return crypto.Keccak256(message)
}

func decodeHex(msg []byte) (res []byte, err error) {
	if bytes.HasPrefix(msg, []byte("0x")) {
		msg = msg[2:]
	}
	res, err = hex.DecodeString(string(msg))
	return
}

func (sig Signature) format() []byte {
	return []byte("0x" + hex.EncodeToString([]byte(sig)))
}
