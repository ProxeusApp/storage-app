package models

import (
	"log"
	"math/big"

	"errors"
)

type StorageProviderInfo struct {
	// TODO: First block used by dApp layer only. Struct conversion to different layer should be discussed
	Address string `json:"address,omitempty"`
	URL     string `json:"url,omitempty"`
	Online  bool   `json:"isOnline,omitempty"`

	Name                  string `json:"name,omitempty"`
	Description           string `json:"description,omitempty"`
	LogoURL               string `json:"logoUrl,omitempty"`
	JurisdictionCountry   string `json:"jurisdictionCountry,omitempty"`
	DataCenter            string `json:"dataCenter,omitempty"`
	MaxFileSizeByte       int64  `json:"maxFileSizeByte,omitempty"`
	MaxStorageDays        int    `json:"maxStorageDays,omitempty"`
	GraceSeconds          int    `json:"graceSeconds,omitempty"` // How long is the file kept after expiration
	TermsAndConditionsURL string `json:"termsAndConditionsUrl,omitempty"`
	PrivacyPolicyURL      string `json:"privacyPolicyUrl,omitempty"`
	PriceByte             string `json:"priceByte,omitempty"` // XES per byte
	PriceDay              string `json:"priceDay,omitempty"`  // XES per day
}

func (me *StorageProviderInfo) MaxFileSizeMB() float32 {
	return float32(me.MaxFileSizeByte) / 1024 / 1024
}

func (me *StorageProviderInfo) priceXesWei(priceString string) (*big.Int, error) {
	xes, err := me.parseStringToXES(priceString) // Ignore parsing
	if err != nil {
		return nil, err
	}

	xesWei, err := me.xesToWei(xes)
	if err != nil {
		return nil, err
	}
	return xesWei, err
}

func (me *StorageProviderInfo) PriceByteXESWei() (*big.Int, error) {
	return me.priceXesWei(me.PriceByte)
}

func (me *StorageProviderInfo) PriceDayXESWei() (*big.Int, error) {
	return me.priceXesWei(me.PriceDay)
}

func (me *StorageProviderInfo) TotalPriceForFile(duration int, fileSizeByte *big.Int) (*big.Int, error) {
	totalPriceSize, err := me.PriceForSizeInXesWei(fileSizeByte)
	if err != nil {
		return totalPriceSize, err
	}
	totalPriceDuration, err := me.PriceForDurationInXesWei(duration)
	if err != nil {
		return totalPriceDuration, err
	}

	return new(big.Int).Add(totalPriceSize, totalPriceDuration), nil
}

func (me *StorageProviderInfo) PriceForDurationInXesWei(duration int) (*big.Int, error) {
	pricePerDay, err := me.PriceDayXESWei()
	if err != nil {
		return pricePerDay, err
	}

	return new(big.Int).Mul(big.NewInt(int64(duration)), pricePerDay), nil
}

func (me *StorageProviderInfo) PriceForSizeInXesWei(fileSizeByte *big.Int) (*big.Int, error) {
	priceByteXesWei, err := me.PriceByteXESWei()
	if err != nil {
		log.Println("storage price error on getting fileSizeByte: ", err)
		return priceByteXesWei, err
	}
	totalPriceByte := new(big.Int).Mul(fileSizeByte, priceByteXesWei)
	return totalPriceByte, nil
}

var ErrParseXesString = errors.New("error parsing xes string")

func (me *StorageProviderInfo) parseStringToXES(xesString string) (*big.Rat, error) {
	x := new(big.Rat)
	x, ok := x.SetString(xesString)
	if !ok {
		return x, ErrParseXesString
	}
	return x, nil
}

var ErrParseStringToXES = errors.New("can't convert xes to xes-wei. " +
	"spp price settings may contain too many decimals")

// 1 XES = 1000000000000000000 XESWei
func (me *StorageProviderInfo) xesToWei(xes *big.Rat) (*big.Int, error) {
	xesWei := xes.Mul(xes, big.NewRat(1000000000000000000, 1))

	if !xesWei.IsInt() {
		return xesWei.Num(), ErrParseStringToXES
	}

	return xesWei.Num(), nil
}
