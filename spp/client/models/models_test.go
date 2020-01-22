package models

import (
	"math/big"
	"testing"
)

func TestServiceProviderInfo_MaxFileSizeMB(t *testing.T) {
	spi := StorageProviderInfo{
		MaxFileSizeByte: 389430,
	}
	if spi.MaxFileSizeMB() != 0.3713894 {
		t.Error(spi.MaxFileSizeMB())
	}
}

func TestServiceProviderInfo_MaxFileSizeMBZeroBytes(t *testing.T) {
	spi := StorageProviderInfo{}
	if spi.MaxFileSizeMB() != 0.0 {
		t.Error("MaxFileSizeMB should be 0 and shouldn't fail")
	}
}

func TestStorageProviderInfo_PriceByteXESWei(t *testing.T) {
	spi := StorageProviderInfo{PriceByte: "0.4"}
	expected := big.NewInt(400000000000000000)
	priceWei, err := spi.PriceByteXESWei()
	if err != nil {
		t.Errorf("Can't convert %s", spi.PriceByte)
	}
	if priceWei.Cmp(expected) != 0 {
		t.Errorf("Wei conversion wrong. Expected %d, got %d", expected, priceWei)
	}

	spi.PriceByte = "0.3"
	expectedDiffers := big.NewInt(400000000000000000)
	priceWei, err = spi.PriceByteXESWei()
	if err != nil {
		t.Errorf("Can't convert %s", spi.PriceByte)
	}
	if priceWei.Cmp(expectedDiffers) == 0 {
		t.Errorf("Wei conversion wrong. Expected %d, got %d", expected, priceWei)
	}
}
