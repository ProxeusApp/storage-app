package account

import "time"

type (
	AddressBookEntry struct {
		Name                    string `json:"name"`
		ETHAddress              string `json:"address"`
		PGPPublicKey            string `json:"pgpPublicKey"`
		ValidatedWithPGPService bool   `json:"validatedWithPGPService"`
		lastPGPServiceCheck     time.Time
		Hidden                  bool `json:"hidden"` // we hide it to be able to use it for older files (rights, etc.)
	}
)

func NewAddressBookEntry(name, ethAddress, pgpPublicKey string) *AddressBookEntry {
	return &AddressBookEntry{
		Name:                    name,
		ETHAddress:              ethAddress,
		PGPPublicKey:            pgpPublicKey,
		ValidatedWithPGPService: false,
		lastPGPServiceCheck:     time.Time{},
		Hidden:                  false,
	}
}
