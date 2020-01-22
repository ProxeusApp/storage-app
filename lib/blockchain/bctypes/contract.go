package bctypes

import (
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

type Contract interface {
	Address() common.Address
	SetAddress(common.Address)
	SetBinding(interface{})

	DeployReceipt() *types.Receipt
	SetDeployReceipt(receipt *types.Receipt)
	DeployTransaction() *Transaction
	SetDeployTransaction(*Transaction)

	BindFromAddress(address common.Address, backend bind.ContractBackend) error
	ExecuteDeployment(auth *bind.TransactOpts, backend bind.ContractBackend) (*types.Transaction, error)
	Transact(opts *bind.TransactOpts, transaction *Transaction) (*types.Transaction, error)
}
