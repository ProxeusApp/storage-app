package bctypes

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/core/types"
)

type Transaction struct {
	Hash     common.Hash
	Ethtx    *types.Transaction
	Function string
	Params   []interface{}

	Mined  bool
	Result *TransactionResult
}
type TransactionResult struct {
	Receipt *types.Receipt
}

func NewTransaction(contract Contract, function string) (*Transaction, error) {
	transaction := Transaction{Function: function}
	return &transaction, nil
}

func (tra Transaction) GetResult() (*TransactionResult, error) {
	if !tra.Mined {
		return nil, errors.New("Not mined yet!")
	}
	return tra.Result, nil
}
