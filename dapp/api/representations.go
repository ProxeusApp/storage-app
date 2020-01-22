package api

type AddressBookRepresentation struct {
	ETHAddress   string `json:"address"`
	Name         string `json:"name"`
	PGPPublicKey string `json:"pgpPublicKey"`
}
