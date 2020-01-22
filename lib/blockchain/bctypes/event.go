package bctypes

import (
	"github.com/ethereum/go-ethereum/core/types"
)

type Event struct {
	UniqueID string
	Log      types.Log
}

func FromLog(log types.Log) *Event {
	uniqueID := log.TxHash.String() + string(log.Index)
	event := Event{UniqueID: uniqueID, Log: log}
	return &event
}
