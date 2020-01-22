package ethereum

import (
	"math/big"
	"sync"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"git.proxeus.com/core/central/dapp/core/embdb"
	"git.proxeus.com/core/central/dapp/core/ethglue"
	"git.proxeus.com/core/central/spp/eth"
)

type (
	xesTransactor struct {
		baseClient *baseClient
		xesAddress common.Address
		pfsAddress common.Address
		xesABI     abi.ABI

		xesTokenContractTransactor *eth.XESTokenContractTransactor
		tokenTransactorMutex       *sync.Mutex

		nonceManager *ethglue.NonceManager

		uniqueTxDB *embdb.DB
	}
)

func NewXesTransactor(baseClient *baseClient, xesAddress, pfsAddress common.Address, xesABI abi.ABI,
	nonceManager *ethglue.NonceManager, tokenTransactorMutex *sync.Mutex) *xesTransactor {
	xesTransactor := new(xesTransactor)
	xesTransactor.baseClient = baseClient
	xesTransactor.xesAddress = xesAddress
	xesTransactor.pfsAddress = pfsAddress
	xesTransactor.xesABI = xesABI
	xesTransactor.nonceManager = nonceManager
	xesTransactor.tokenTransactorMutex = tokenTransactorMutex
	return xesTransactor
}

func (me *xesTransactor) xesTransferEstimateGas(ethPrivKeyFrom string, ethAddressTo string, xesAmount *big.Int) (*bind.TransactOpts, error) {
	to := common.HexToAddress(ethAddressTo)
	opts, err := me.baseClient.getAuth(ethPrivKeyFrom)
	input, err := me.xesABI.Pack("transfer", to, xesAmount)
	if err != nil {
		return nil, err
	}

	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	msg := ethereum.CallMsg{From: opts.From, To: &me.xesAddress, Value: value, Data: input}

	gas, err := me.baseClient.estimateGas(msg)
	if err != nil {
		return nil, err
	}
	opts.Value = value
	opts.GasPrice = gas.GasPrice
	opts.GasLimit = gas.GasLimit

	return opts, err
}

func (me *xesTransactor) xesTransfer(ethPrivKeyFrom string, ethAddressTo string, xesAmount *big.Int) (*types.Transaction, error) {
	to := common.HexToAddress(ethAddressTo)
	opts, err := me.baseClient.getAuth(ethPrivKeyFrom)
	input, err := me.xesABI.Pack("transfer", to, xesAmount)
	if err != nil {
		return nil, err
	}
	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	msg := ethereum.CallMsg{From: opts.From, To: &me.xesAddress, Value: value, Data: input}
	gas, err := me.baseClient.estimateGas(msg)
	if err != nil {
		return nil, err
	}
	opts.Value = value
	opts.GasPrice = gas.GasPrice
	opts.GasLimit = gas.GasLimit

	me.tokenTransactorMutex.Lock()
	opts.Nonce = me.nonceManager.NextNonce()
	tx, err := me.xesTokenContractTransactor.Transfer(opts, common.HexToAddress(ethAddressTo), xesAmount)
	me.nonceManager.OnError(err)
	me.tokenTransactorMutex.Unlock()
	if err != nil {
		return nil, err
	}
	me.baseClient.putXesTx("xes-transfer", opts.From, tx, xesAmount)
	return tx, nil
}

func (me *xesTransactor) xesApprove(ethPrivKeyFrom string, ethAddressTo string, xesAmount *big.Int) (*types.Transaction, error) {
	to := common.HexToAddress(ethAddressTo)
	opts, err := me.baseClient.getAuth(ethPrivKeyFrom)
	input, err := me.xesABI.Pack("approve", to, xesAmount)
	//var err error

	// Ensure a valid value field and resolve the account nonce
	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	msg := ethereum.CallMsg{From: opts.From, To: &me.xesAddress, Value: value, Data: input}
	gas, err := me.baseClient.estimateGas(msg)
	if err != nil {
		return nil, err
	}
	opts.Value = value
	opts.GasPrice = gas.GasPrice
	opts.GasLimit = gas.GasLimit

	me.tokenTransactorMutex.Lock()
	opts.Nonce = me.nonceManager.NextNonce()
	tx, err := me.xesTokenContractTransactor.Approve(opts, to, xesAmount)
	me.nonceManager.OnError(err)
	me.tokenTransactorMutex.Unlock()
	if err != nil {
		return nil, err
	}
	me.baseClient.putXesTx("xes-approve", opts.From, tx, xesAmount)
	return tx, nil
}

// TODO(mmal): proper harness for tests
var TestingModeBreakGas bool

func (me *xesTransactor) xesApproveToProxeusFSEstimateGas(ethPrivKeyFrom string, xesAmount *big.Int) (*bind.TransactOpts, error) {
	to := me.pfsAddress
	opts, err := me.baseClient.getAuth(ethPrivKeyFrom)
	input, err := me.xesABI.Pack("approve", to, xesAmount)

	// Ensure a valid value field and resolve the account nonce
	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	msg := ethereum.CallMsg{From: opts.From, To: &me.xesAddress, Value: value, Data: input}

	gas, err := me.baseClient.estimateGas(msg)
	if err != nil {
		return nil, err
	}
	opts.Value = value
	opts.GasPrice = gas.GasPrice
	opts.GasLimit = gas.GasLimit
	if TestingModeBreakGas {
		opts.GasLimit = 1
	}

	return opts, nil
}

func (me *xesTransactor) xesApproveToProxeusFS(ethPrivKeyFrom string, xesAmount *big.Int) (*types.Transaction, error) {
	to := me.pfsAddress
	opts, err := me.baseClient.getAuth(ethPrivKeyFrom)
	input, err := me.xesABI.Pack("approve", to, xesAmount)
	//var err error

	// Ensure a valid value field and resolve the account nonce
	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	msg := ethereum.CallMsg{From: opts.From, To: &me.xesAddress, Value: value, Data: input}
	gas, err := me.baseClient.estimateGas(msg)
	if err != nil {
		return nil, err
	}
	opts.Value = value
	opts.GasPrice = gas.GasPrice
	opts.GasLimit = gas.GasLimit
	if TestingModeBreakGas {
		opts.GasLimit = 1
	}

	me.tokenTransactorMutex.Lock()
	opts.Nonce = me.nonceManager.NextNonce()
	tx, err := me.xesTokenContractTransactor.Approve(opts, to, xesAmount)
	me.nonceManager.OnError(err)
	me.tokenTransactorMutex.Unlock()
	if err != nil {
		return nil, err
	}
	me.baseClient.putXesTx("xes-approve", opts.From, tx, xesAmount)
	return tx, nil
}

func (me *xesTransactor) xesTransferFrom(ethPrivKeyFrom, ethAddressFrom, ethAddressTo string, xesAmount *big.Int) (*types.Transaction, error) {
	from := common.HexToAddress(ethAddressFrom)
	to := common.HexToAddress(ethAddressTo)
	opts, err := me.baseClient.getAuth(ethPrivKeyFrom)
	input, err := me.xesABI.Pack("transferFrom", from, to, xesAmount)
	//var err error

	// Ensure a valid value field and resolve the account nonce
	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	msg := ethereum.CallMsg{From: opts.From, To: &me.xesAddress, Value: value, Data: input}
	gas, err := me.baseClient.estimateGas(msg)
	if err != nil {
		return nil, err
	}
	opts.Value = value
	opts.GasPrice = gas.GasPrice
	opts.GasLimit = gas.GasLimit

	me.tokenTransactorMutex.Lock()
	opts.Nonce = me.nonceManager.NextNonce()
	tx, err := me.xesTokenContractTransactor.TransferFrom(opts, from, to, xesAmount)
	me.nonceManager.OnError(err)
	me.tokenTransactorMutex.Unlock()
	if err != nil {
		return nil, err
	}
	me.baseClient.putXesTx("xes-transferFrom", opts.From, tx, xesAmount)
	return tx, nil
}
