package util

import (
	"encoding/hex"

	"github.com/ethereum/go-ethereum/common"
)

func StrHexToBytes32(msg string) [32]byte {
	return common.HexToHash(msg)
}

func Bytes32ToHexStr(src [32]byte) string {
	return "0x" + hex.EncodeToString(src[:])
}

func Bytes32Empty(src [32]byte) bool {
	countZeros := 0
	for _, b := range src {
		if b == 0 {
			countZeros++
			continue
		}
		return false
	}
	return countZeros == 32
}
