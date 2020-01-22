package account

import (
	b64 "encoding/base64"
	"encoding/json"
	"strings"
)

/**
  Keystores represents the proxeus key and offers some utility functions
*/
type ProxeusKeystore struct {
	Keystores  []*Keystore            `json:"keystore"`
	Name       string                 `json:"name"`
	PrivateKey string                 `json:"privateKey"`
	PublicKey  string                 `json:"publicKey"`
	PgpPw      string                 `json:"pgpPw"`
	PgpKeys    map[string]interface{} `json:"pgpKeys"`
}

type Keystore struct {
	Address string `json:"address"`
	Crypto  struct {
		Cipher       string `json:"cipher"`
		Ciphertext   string `json:"ciphertext"`
		Cipherparams struct {
			Iv string `json:"iv"`
		} `json:"cipherparams"`
		Kdf       string `json:"kdf"`
		Kdfparams struct {
			Dklen int    `json:"dklen"`
			N     int    `json:"n"`
			P     int    `json:"p"`
			R     int    `json:"r"`
			Salt  string `json:"salt"`
		} `json:"kdfparams"`
		Mac string `json:"mac"`
	} `json:"crypto"`
	ID      string `json:"id"`
	Version int    `json:"version"`
}

// Parses the content into KeyStores
func (ks *ProxeusKeystore) Load(content []byte) error {
	sDec := make([]byte, b64.StdEncoding.DecodedLen(len(content)))
	var n int
	n, err := b64.StdEncoding.Decode(sDec, content)
	if err != nil {
		return err
	}
	sDec = sDec[0:n]
	err = json.Unmarshal(sDec, ks)
	return err
}

// Returns the first keystore in the array. We're only handling one for now, and that's Ethereum
func (ks *ProxeusKeystore) ETHKeystore() *Keystore {
	return ks.Keystores[0]
}

func (ks *ProxeusKeystore) SetETHKeystore(content []byte) (err error) {
	keystore := &Keystore{}
	err = json.Unmarshal(content, keystore)
	if err != nil {
		return
	}
	ks.Keystores = append(ks.Keystores, keystore)
	return
}

func (ks *ProxeusKeystore) AddressWithPrefix() string {
	address := ks.ETHKeystore().Address
	if !strings.HasPrefix(address, "0x") {
		address = "0x" + address
	}
	return address
}
