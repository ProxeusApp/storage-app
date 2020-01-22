package ethereum

import (
	"encoding/json"
	"math/big"
	"os"
	"sync"
	"time"

	"git.proxeus.com/core/central/dapp/core/util"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"git.proxeus.com/core/central/dapp/core/embdb"
	"git.proxeus.com/core/central/dapp/core/ethglue"
	"git.proxeus.com/core/central/spp/eth"
)

const (
	uniqueStorageName = "unique"
)

type (
	fsTransactor struct {
		baseClient   *baseClient
		pfsAddress   common.Address
		proxeusFSABI abi.ABI

		proxeusFSContractTransactor *eth.ProxeusFSContractTransactor
		proxeusFSTransactorMutex    sync.Mutex

		nonceManager *ethglue.NonceManager

		uniqueTxDB *embdb.DB
	}
)

func NewFsTransactor(baseClient *baseClient, pfsAddress common.Address, proxeusFSABI abi.ABI, nonceManager *ethglue.NonceManager) *fsTransactor {
	fsTransactor := new(fsTransactor)
	fsTransactor.baseClient = baseClient
	fsTransactor.pfsAddress = pfsAddress
	fsTransactor.proxeusFSABI = proxeusFSABI
	fsTransactor.nonceManager = nonceManager
	return fsTransactor
}

func (me *fsTransactor) initLocalStorage(storageDir string) (err error) {
	me.uniqueTxDB, err = embdb.Open(storageDir, uniqueStorageName)
	if err != nil {
		return err
	}
	return
}

func (me *fsTransactor) createFileDefinedSignersEstimateGas(ethPrivKeyFrom string, fileHash [32]byte, fileName string, definedSigners []common.Address,
	expiry *big.Int, replacesFile [32]byte, storageProviders []common.Address, xesAmount *big.Int) (*bind.TransactOpts, error) {

	opts, err := me.baseClient.getAuth(ethPrivKeyFrom)
	input, err := me.proxeusFSABI.Pack("createFileDefinedSigners", fileHash, definedSigners, expiry, replacesFile,
		storageProviders, xesAmount) //bytes32 hash, address[] definedSigners, uint expiry, bytes32 replacesFile, address[] prvs
	if err != nil {
		return nil, err
	}

	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}

	gas, err := me.baseClient.estimateGas(ethereum.CallMsg{From: opts.From, To: &me.pfsAddress, Value: value, Data: input})
	if err != nil {
		return nil, err
	}

	opts.Value = value
	opts.GasPrice = gas.GasPrice
	opts.GasLimit = gas.GasLimit

	return opts, err
}

func (me *fsTransactor) createFileDefinedSigners(ethPrivKeyFrom string, fileHash [32]byte, fileName string, definedSigners []common.Address,
	expiry *big.Int, replacesFile [32]byte, storageProviders []common.Address, xesAmount *big.Int) (*types.Transaction, error) {

	err := me.alreadyExists(fileHash[:])
	if err != nil {
		return nil, err
	}

	opts, err := me.createFileDefinedSignersEstimateGas(ethPrivKeyFrom, fileHash, fileName, definedSigners, expiry, replacesFile, storageProviders, xesAmount)
	if err != nil {
		return nil, err
	}

	me.proxeusFSTransactorMutex.Lock()
	opts.Nonce = me.nonceManager.NextNonce()

	tx, err := me.proxeusFSContractTransactor.CreateFileDefinedSigners(opts, fileHash, definedSigners, expiry, replacesFile, storageProviders, xesAmount)
	me.nonceManager.OnError(err)
	me.proxeusFSTransactorMutex.Unlock()
	if err != nil {
		return nil, err
	}

	me.safeWhileTxPending(fileHash[:])
	me.baseClient.putFileTx(PendingTypeRegister, util.Bytes32ToHexStr(fileHash), fileName, opts.From, tx, &definedSigners)
	return tx, nil
}

func (me *fsTransactor) createFileShared(ethPrivKeyFrom string, fileHash [32]byte, fileName string, mandatorySigners *big.Int,
	expiry *big.Int, replacesFile [32]byte, storageProviders []common.Address, readers []common.Address,
	xesAmount *big.Int) (*types.Transaction, error) {

	err := me.alreadyExists(fileHash[:])
	if err != nil {
		return nil, err
	}

	opts, err := me.baseClient.getAuth(ethPrivKeyFrom)
	//bytes32 hash, address[] definedSigners, uint expiry, bytes32 replacesFile, address[] prvs, uint xesAmount
	input, err := me.proxeusFSABI.Pack("createFileShared", fileHash, mandatorySigners, expiry, replacesFile,
		storageProviders, readers, xesAmount)

	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	gas, err := me.baseClient.estimateGas((ethereum.CallMsg{From: opts.From, To: &me.pfsAddress, Value: value, Data: input}))
	if err != nil {
		return nil, err
	}
	opts.Value = value
	opts.GasPrice = gas.GasPrice
	opts.GasLimit = gas.GasLimit

	me.proxeusFSTransactorMutex.Lock()
	opts.Nonce = me.nonceManager.NextNonce()
	tx, err := me.proxeusFSContractTransactor.CreateFileShared(opts, fileHash, mandatorySigners, expiry, replacesFile, storageProviders, readers, xesAmount)
	me.nonceManager.OnError(err)
	me.proxeusFSTransactorMutex.Unlock()
	if err != nil {
		return nil, err
	}
	me.safeWhileTxPending(fileHash[:])
	me.baseClient.putFileTx(PendingTypeRegister, util.Bytes32ToHexStr(fileHash), fileName, opts.From, tx, nil)
	return tx, nil
}

func (me *fsTransactor) createFileUndefinedSignersEstimateGas(ethPrivKeyFrom string, fileHash [32]byte, fileName string, mandatorySigners *big.Int,
	expiry *big.Int, replacesFile [32]byte, storageProviders []common.Address, xesAmount *big.Int) (*bind.TransactOpts, error) {

	opts, err := me.baseClient.getAuth(ethPrivKeyFrom)

	input, err := me.proxeusFSABI.Pack("createFileUndefinedSigners", fileHash, mandatorySigners, expiry,
		replacesFile, storageProviders, xesAmount) //bytes32 hash, address[] definedSigners, uint expiry, bytes32 replacesFile, address[] prvs
	if err != nil {
		return nil, err
	}

	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}

	gas, err := me.baseClient.estimateGas(ethereum.CallMsg{From: opts.From, To: &me.pfsAddress, Value: value, Data: input})
	if err != nil {
		return nil, err
	}

	opts.Value = value
	opts.GasPrice = gas.GasPrice
	opts.GasLimit = gas.GasLimit

	return opts, err
}

func (me *fsTransactor) createFileUndefinedSigners(ethPrivKeyFrom string, fileHash [32]byte, fileName string, mandatorySigners *big.Int,
	expiry *big.Int, replacesFile [32]byte, storageProviders []common.Address, xesAmount *big.Int) (*types.Transaction, error) {

	err := me.alreadyExists(fileHash[:])
	if err != nil {
		return nil, err
	}

	opts, err := me.createFileUndefinedSignersEstimateGas(ethPrivKeyFrom, fileHash, fileName, mandatorySigners, expiry, replacesFile, storageProviders, xesAmount)
	if err != nil {
		return nil, err
	}

	me.proxeusFSTransactorMutex.Lock()
	opts.Nonce = me.nonceManager.NextNonce()

	tx, err := me.proxeusFSContractTransactor.CreateFileUndefinedSigners(opts, fileHash, mandatorySigners, expiry, replacesFile, storageProviders, xesAmount)
	me.nonceManager.OnError(err)
	me.proxeusFSTransactorMutex.Unlock()
	if err != nil {
		return nil, err
	}
	me.safeWhileTxPending(fileHash[:])
	me.baseClient.putFileTx(PendingTypeRegister, util.Bytes32ToHexStr(fileHash), fileName, opts.From, tx, nil)
	return tx, nil
}

func (me *fsTransactor) fileRemoveEstimateGas(ethPrivKeyFrom string, fileHash [32]byte) (*bind.TransactOpts, error) {
	opts, err := me.baseClient.getAuth(ethPrivKeyFrom)
	input, err := me.proxeusFSABI.Pack("fileRemove", fileHash)

	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}

	gas, err := me.baseClient.estimateGas(ethereum.CallMsg{From: opts.From, To: &me.pfsAddress, Value: value, Data: input})
	if err != nil {
		return nil, err
	}

	opts.Value = value
	opts.GasPrice = gas.GasPrice
	opts.GasLimit = gas.GasLimit

	return opts, err
}

func (me *fsTransactor) fileRemove(ethPrivKeyFrom string, fileHash [32]byte, filename string) (*types.Transaction, error) {
	opts, err := me.fileRemoveEstimateGas(ethPrivKeyFrom, fileHash)
	if err != nil {
		return nil, err
	}

	me.proxeusFSTransactorMutex.Lock()
	opts.Nonce = me.nonceManager.NextNonce()
	tx, err := me.proxeusFSContractTransactor.FileRemove(opts, fileHash)
	me.nonceManager.OnError(err)
	me.proxeusFSTransactorMutex.Unlock()
	if err != nil {
		return nil, err
	}
	me.baseClient.putFileTx(PendingTypeRemove, util.Bytes32ToHexStr(fileHash), filename, opts.From, tx, nil)
	return tx, nil
}

func (me *fsTransactor) fileRequestAccess(ethPrivKeyFrom string, fileHash [32]byte) (*types.Transaction, error) {
	opts, err := me.baseClient.getAuth(ethPrivKeyFrom)
	input, err := me.proxeusFSABI.Pack("fileRequestAccess", fileHash)

	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	gas, err := me.baseClient.estimateGas((ethereum.CallMsg{From: opts.From, To: &me.pfsAddress, Value: value, Data: input}))
	if err != nil {
		return nil, err
	}
	opts.Value = value
	opts.GasPrice = gas.GasPrice
	opts.GasLimit = gas.GasLimit

	me.proxeusFSTransactorMutex.Lock()
	opts.Nonce = me.nonceManager.NextNonce()
	tx, err := me.proxeusFSContractTransactor.FileRequestAccess(opts, fileHash)
	me.nonceManager.OnError(err)
	me.proxeusFSTransactorMutex.Unlock()
	if err != nil {
		return nil, err
	}
	me.baseClient.putFileHashTx("requestAccess", util.Bytes32ToHexStr(fileHash), opts.From, tx)
	return tx, nil
}

func (me *fsTransactor) fileRequestSignEstimateGas(ethPrivKeyFrom string, fileHash [32]byte, signer []common.Address) (*bind.TransactOpts, error) {
	opts, err := me.baseClient.getAuth(ethPrivKeyFrom)
	input, err := me.proxeusFSABI.Pack("fileRequestSign", fileHash, signer)

	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}

	gas, err := me.baseClient.estimateGas((ethereum.CallMsg{From: opts.From, To: &me.pfsAddress, Value: value, Data: input}))
	if err != nil {
		return nil, err
	}

	opts.Value = value
	opts.GasPrice = gas.GasPrice
	opts.GasLimit = gas.GasLimit

	return opts, err
}

func (me *fsTransactor) fileRequestSign(ethPrivKeyFrom string, fileHash [32]byte, filename string, signer []common.Address) (*types.Transaction, error) {
	opts, err := me.fileRequestSignEstimateGas(ethPrivKeyFrom, fileHash, signer)
	if err != nil {
		return nil, err
	}

	me.proxeusFSTransactorMutex.Lock()
	opts.Nonce = me.nonceManager.NextNonce()
	tx, err := me.proxeusFSContractTransactor.FileRequestSign(opts, fileHash, signer)
	me.nonceManager.OnError(err)
	me.proxeusFSTransactorMutex.Unlock()
	if err != nil {
		return nil, err
	}
	me.baseClient.putFileSignTx("requestSign", util.Bytes32ToHexStr(fileHash), filename, opts.From, tx, &signer)
	return tx, nil
}

func (me *fsTransactor) fileRevokePermEstimateGas(ethPrivKeyFrom string, fileHash [32]byte, addr []common.Address) (*bind.TransactOpts, error) {
	opts, err := me.baseClient.getAuth(ethPrivKeyFrom)
	input, err := me.proxeusFSABI.Pack("fileRevokePerm", fileHash, addr)

	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}

	gas, err := me.baseClient.estimateGas((ethereum.CallMsg{From: opts.From, To: &me.pfsAddress, Value: value, Data: input}))
	if err != nil {
		return nil, err
	}

	opts.Value = value
	opts.GasPrice = gas.GasPrice
	opts.GasLimit = gas.GasLimit

	return opts, err
}

func (me *fsTransactor) fileRevokePerm(ethPrivKeyFrom string, fileHash [32]byte, filename string, addr []common.Address) (*types.Transaction, error) {
	opts, err := me.fileRevokePermEstimateGas(ethPrivKeyFrom, fileHash, addr)
	if err != nil {
		return nil, err
	}

	me.proxeusFSTransactorMutex.Lock()
	opts.Nonce = me.nonceManager.NextNonce()
	tx, err := me.proxeusFSContractTransactor.FileRevokePerm(opts, fileHash, addr)
	me.nonceManager.OnError(err)
	me.proxeusFSTransactorMutex.Unlock()
	if err != nil {
		return nil, err
	}
	me.baseClient.putFileTx(PendingTypeRevoke, util.Bytes32ToHexStr(fileHash), filename, opts.From, tx, &addr)
	return tx, nil
}

func (me *fsTransactor) FileSetPermEstimateGas(ethPrivKeyFrom string, fileHash [32]byte, addr []common.Address) (*bind.TransactOpts, error) {
	opts, err := me.baseClient.getAuth(ethPrivKeyFrom)
	input, err := me.proxeusFSABI.Pack("fileSetPerm", fileHash, addr)

	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}

	gas, err := me.baseClient.estimateGas((ethereum.CallMsg{From: opts.From, To: &me.pfsAddress, Value: value, Data: input}))
	if err != nil {
		return nil, err
	}

	opts.Value = value
	opts.GasPrice = gas.GasPrice
	opts.GasLimit = gas.GasLimit

	return opts, err
}

func (me *fsTransactor) fileSetPerm(ethPrivKeyFrom string, fileHash [32]byte, filename string, addr []common.Address) (*types.Transaction, error) {
	opts, err := me.FileSetPermEstimateGas(ethPrivKeyFrom, fileHash, addr)
	if err != nil {
		return nil, err
	}

	me.proxeusFSTransactorMutex.Lock()
	opts.Nonce = me.nonceManager.NextNonce()
	tx, err := me.proxeusFSContractTransactor.FileSetPerm(opts, fileHash, addr)
	me.nonceManager.OnError(err)
	me.proxeusFSTransactorMutex.Unlock()
	if err != nil {
		return nil, err
	}
	me.baseClient.putFileTx(PendingTypeShare, util.Bytes32ToHexStr(fileHash), filename, opts.From, tx, &addr)
	return tx, nil
}

func (me *fsTransactor) fileSignEstimateGas(ethPrivKeyFrom string, fileHash [32]byte) (*bind.TransactOpts, error) {
	opts, err := me.baseClient.getAuth(ethPrivKeyFrom)
	input, err := me.proxeusFSABI.Pack("fileSign", fileHash)

	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}

	gas, err := me.baseClient.estimateGas((ethereum.CallMsg{From: opts.From, To: &me.pfsAddress, Value: value, Data: input}))
	if err != nil {
		return nil, err
	}

	opts.Value = value
	opts.GasPrice = gas.GasPrice
	opts.GasLimit = gas.GasLimit

	return opts, err
}

func (me *fsTransactor) fileSign(ethPrivKeyFrom, filename string, fileHash [32]byte) (*types.Transaction, error) {
	opts, err := me.fileSignEstimateGas(ethPrivKeyFrom, fileHash)
	if err != nil {
		return nil, err
	}

	me.proxeusFSTransactorMutex.Lock()
	opts.Nonce = me.nonceManager.NextNonce()
	tx, err := me.proxeusFSContractTransactor.FileSign(opts, fileHash)
	me.nonceManager.OnError(err)
	me.proxeusFSTransactorMutex.Unlock()
	if err != nil {
		return nil, err
	}
	me.baseClient.putFileTx(PendingTypeSign, util.Bytes32ToHexStr(fileHash), filename, opts.From, tx, nil)
	return tx, nil
}

func (me *fsTransactor) setDappVersion(ethPrivKeyFrom string, version [32]byte) (*types.Transaction, error) {
	opts, err := me.baseClient.getAuth(ethPrivKeyFrom)
	input, err := me.proxeusFSABI.Pack("setDappVersion", version)

	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	gas, err := me.baseClient.estimateGas((ethereum.CallMsg{From: opts.From, To: &me.pfsAddress, Value: value, Data: input}))
	if err != nil {
		return nil, err
	}
	opts.Value = value
	opts.GasPrice = gas.GasPrice
	opts.GasLimit = gas.GasLimit

	me.proxeusFSTransactorMutex.Lock()
	opts.Nonce = me.nonceManager.NextNonce()
	tx, err := me.proxeusFSContractTransactor.SetDappVersion(opts, version)
	me.nonceManager.OnError(err)
	me.proxeusFSTransactorMutex.Unlock()
	if err != nil {
		return nil, err
	}
	me.baseClient.putTx("dappVersion", opts.From, tx)
	return tx, nil
}

func (me *fsTransactor) spAdd(ethPrivKeyFrom string, strProv common.Address, urlPrefix [32]byte) (*types.Transaction, error) {
	opts, err := me.baseClient.getAuth(ethPrivKeyFrom)
	input, err := me.proxeusFSABI.Pack("spAdd", strProv, urlPrefix)

	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	gas, err := me.baseClient.estimateGas((ethereum.CallMsg{From: opts.From, To: &me.pfsAddress, Value: value, Data: input}))
	if err != nil {
		return nil, err
	}
	opts.Value = value
	opts.GasPrice = gas.GasPrice
	opts.GasLimit = gas.GasLimit

	me.proxeusFSTransactorMutex.Lock()
	opts.Nonce = me.nonceManager.NextNonce()
	tx, err := me.proxeusFSContractTransactor.SpAdd(opts, strProv, urlPrefix)
	me.nonceManager.OnError(err)
	me.proxeusFSTransactorMutex.Unlock()
	if err != nil {
		return nil, err
	}
	me.baseClient.putTx("spAdd", opts.From, tx)
	return tx, nil
}

func (me *fsTransactor) spUpdate(ethPrivKeyFrom string, strProv common.Address, urlPrefix [32]byte) (*types.Transaction, error) {
	opts, err := me.baseClient.getAuth(ethPrivKeyFrom)
	input, err := me.proxeusFSABI.Pack("spUpdate", strProv, urlPrefix)

	value := opts.Value
	if value == nil {
		value = new(big.Int)
	}
	gas, err := me.baseClient.estimateGas((ethereum.CallMsg{From: opts.From, To: &me.pfsAddress, Value: value, Data: input}))
	if err != nil {
		return nil, err
	}
	opts.Value = value
	opts.GasPrice = gas.GasPrice
	opts.GasLimit = gas.GasLimit

	me.proxeusFSTransactorMutex.Lock()
	opts.Nonce = me.nonceManager.NextNonce()
	tx, err := me.proxeusFSContractTransactor.SpUpdate(opts, strProv, urlPrefix)
	me.nonceManager.OnError(err)
	me.proxeusFSTransactorMutex.Unlock()
	if err != nil {
		return nil, err
	}
	me.baseClient.putTx("spUpdate", opts.From, tx)
	return tx, nil
}

func (me *fsTransactor) alreadyExists(fileHash []byte) error {
	ufBts, _ := me.uniqueTxDB.Get(fileHash)
	if len(ufBts) > 0 {
		ufh := uniqueFileHash{}
		err := json.Unmarshal(ufBts, &ufh)
		// condition should be substituted to the next one when expiry is implemented
		//if err == nil && ufh.Expired > time.Now().Unix() {
		if err == nil {
			return os.ErrExist
		}
	}
	return nil
}

func (me *fsTransactor) safeWhileTxPending(fileHash []byte) {
	ufh := uniqueFileHash{Expired: time.Now().Add(time.Minute * 10).Unix()}
	ufhBts, _ := json.Marshal(ufh)
	me.uniqueTxDB.Put(fileHash, ufhBts)
}

func (me *fsTransactor) close() {
	if me.uniqueTxDB != nil {
		me.uniqueTxDB.Close()
	}
}
