// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package ProxeusFSContract

import (
	"math/big"
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// ProxeusFSContractABI is the input ABI used to generate the binding from.
const ProxeusFSContractABI = "[{\"constant\":true,\"inputs\":[],\"name\":\"spList\",\"outputs\":[{\"name\":\"\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"strPrv\",\"type\":\"address\"}],\"name\":\"spInfo\",\"outputs\":[{\"name\":\"urlPrefix\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"}],\"name\":\"fileSign\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"}],\"name\":\"fileVerify\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"},{\"name\":\"\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"sendr\",\"type\":\"address\"}],\"name\":\"XESAllowence\",\"outputs\":[{\"name\":\"sum\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"}],\"name\":\"fileSigners\",\"outputs\":[{\"name\":\"\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"},{\"name\":\"addr\",\"type\":\"address\"}],\"name\":\"fileHasSP\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"version\",\"type\":\"bytes32\"}],\"name\":\"setDappVersion\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"strProv\",\"type\":\"address\"},{\"name\":\"urlPrefix\",\"type\":\"bytes32\"}],\"name\":\"spAdd\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"prvs\",\"type\":\"address[]\"}],\"name\":\"XESAmountPerFile\",\"outputs\":[{\"name\":\"sum\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"},{\"name\":\"addr\",\"type\":\"address\"},{\"name\":\"write\",\"type\":\"bool\"}],\"name\":\"fileGetPerm\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"},{\"name\":\"addr\",\"type\":\"address[]\"}],\"name\":\"fileSetPerm\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"}],\"name\":\"fileRemove\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"},{\"name\":\"signer\",\"type\":\"address[]\"}],\"name\":\"fileRequestSign\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"dappVersion\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"},{\"name\":\"mandatorySigners\",\"type\":\"uint256\"},{\"name\":\"expiry\",\"type\":\"uint256\"},{\"name\":\"replacesFile\",\"type\":\"bytes32\"},{\"name\":\"prvs\",\"type\":\"address[]\"}],\"name\":\"createFileUndefinedSigners\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"}],\"name\":\"fileRequestAccess\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"}],\"name\":\"fileInfo\",\"outputs\":[{\"name\":\"id\",\"type\":\"bytes32\"},{\"name\":\"ownr\",\"type\":\"address\"},{\"name\":\"fileType\",\"type\":\"uint256\"},{\"name\":\"removed\",\"type\":\"bool\"},{\"name\":\"expiry\",\"type\":\"uint256\"},{\"name\":\"isPublic\",\"type\":\"bool\"},{\"name\":\"thumbnailHash\",\"type\":\"bytes32\"},{\"name\":\"fparent\",\"type\":\"bytes32\"},{\"name\":\"replacesFile\",\"type\":\"bytes32\"},{\"name\":\"readAccess\",\"type\":\"address[]\"},{\"name\":\"definedSigners\",\"type\":\"address[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"},{\"name\":\"pParent\",\"type\":\"bytes32\"},{\"name\":\"pPublic\",\"type\":\"bool\"}],\"name\":\"createFileThumbnail\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"_eternalstorage\",\"type\":\"address\"}],\"name\":\"setEternalStorage\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"},{\"name\":\"definedSigners\",\"type\":\"address[]\"},{\"name\":\"expiry\",\"type\":\"uint256\"},{\"name\":\"replacesFile\",\"type\":\"bytes32\"},{\"name\":\"prvs\",\"type\":\"address[]\"}],\"name\":\"createFileDefinedSigners\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"}],\"name\":\"fileExpiry\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"},{\"name\":\"strPrv\",\"type\":\"address\"}],\"name\":\"fileAddSP\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"}],\"name\":\"isFileRemoved\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"}],\"name\":\"fileSignersCount\",\"outputs\":[{\"name\":\"\",\"type\":\"uint256\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"hash\",\"type\":\"bytes32\"},{\"name\":\"addr\",\"type\":\"address[]\"}],\"name\":\"fileRevokePerm\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":false,\"inputs\":[{\"name\":\"strPrv\",\"type\":\"address\"},{\"name\":\"urlPrefix\",\"type\":\"bytes32\"}],\"name\":\"spUpdate\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"fileList\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[{\"name\":\"ownr\",\"type\":\"address\"},{\"name\":\"tokenAddr\",\"type\":\"address\"}],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"hash\",\"type\":\"bytes32\"}],\"name\":\"Deleted\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"oldHash\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"newHash\",\"type\":\"bytes32\"}],\"name\":\"UpdatedEvent\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"hash\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"to\",\"type\":\"address\"}],\"name\":\"RequestSign\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"hash\",\"type\":\"bytes32\"},{\"indexed\":true,\"name\":\"who\",\"type\":\"address\"}],\"name\":\"NotifySign\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"hash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"oldOwner\",\"type\":\"address\"},{\"indexed\":false,\"name\":\"newOwner\",\"type\":\"address\"}],\"name\":\"OwnerChanged\",\"type\":\"event\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"hash\",\"type\":\"bytes32\"},{\"indexed\":false,\"name\":\"who\",\"type\":\"address\"}],\"name\":\"RequestAccess\",\"type\":\"event\"}]"

// ProxeusFSContract is an auto generated Go binding around an Ethereum contract.
type ProxeusFSContract struct {
	ProxeusFSContractCaller     // Read-only binding to the contract
	ProxeusFSContractTransactor // Write-only binding to the contract
	ProxeusFSContractFilterer   // Log filterer for contract events
}

// ProxeusFSContractCaller is an auto generated read-only Go binding around an Ethereum contract.
type ProxeusFSContractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProxeusFSContractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type ProxeusFSContractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProxeusFSContractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type ProxeusFSContractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// ProxeusFSContractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type ProxeusFSContractSession struct {
	Contract     *ProxeusFSContract // Generic contract binding to set the session for
	CallOpts     bind.CallOpts      // Call options to use throughout this session
	TransactOpts bind.TransactOpts  // Transaction auth options to use throughout this session
}

// ProxeusFSContractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type ProxeusFSContractCallerSession struct {
	Contract *ProxeusFSContractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts            // Call options to use throughout this session
}

// ProxeusFSContractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type ProxeusFSContractTransactorSession struct {
	Contract     *ProxeusFSContractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts            // Transaction auth options to use throughout this session
}

// ProxeusFSContractRaw is an auto generated low-level Go binding around an Ethereum contract.
type ProxeusFSContractRaw struct {
	Contract *ProxeusFSContract // Generic contract binding to access the raw methods on
}

// ProxeusFSContractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type ProxeusFSContractCallerRaw struct {
	Contract *ProxeusFSContractCaller // Generic read-only contract binding to access the raw methods on
}

// ProxeusFSContractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type ProxeusFSContractTransactorRaw struct {
	Contract *ProxeusFSContractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewProxeusFSContract creates a new instance of ProxeusFSContract, bound to a specific deployed contract.
func NewProxeusFSContract(address common.Address, backend bind.ContractBackend) (*ProxeusFSContract, error) {
	contract, err := bindProxeusFSContract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &ProxeusFSContract{ProxeusFSContractCaller: ProxeusFSContractCaller{contract: contract}, ProxeusFSContractTransactor: ProxeusFSContractTransactor{contract: contract}, ProxeusFSContractFilterer: ProxeusFSContractFilterer{contract: contract}}, nil
}

// NewProxeusFSContractCaller creates a new read-only instance of ProxeusFSContract, bound to a specific deployed contract.
func NewProxeusFSContractCaller(address common.Address, caller bind.ContractCaller) (*ProxeusFSContractCaller, error) {
	contract, err := bindProxeusFSContract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &ProxeusFSContractCaller{contract: contract}, nil
}

// NewProxeusFSContractTransactor creates a new write-only instance of ProxeusFSContract, bound to a specific deployed contract.
func NewProxeusFSContractTransactor(address common.Address, transactor bind.ContractTransactor) (*ProxeusFSContractTransactor, error) {
	contract, err := bindProxeusFSContract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &ProxeusFSContractTransactor{contract: contract}, nil
}

// NewProxeusFSContractFilterer creates a new log filterer instance of ProxeusFSContract, bound to a specific deployed contract.
func NewProxeusFSContractFilterer(address common.Address, filterer bind.ContractFilterer) (*ProxeusFSContractFilterer, error) {
	contract, err := bindProxeusFSContract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &ProxeusFSContractFilterer{contract: contract}, nil
}

// bindProxeusFSContract binds a generic wrapper to an already deployed contract.
func bindProxeusFSContract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(ProxeusFSContractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ProxeusFSContract *ProxeusFSContractRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ProxeusFSContract.Contract.ProxeusFSContractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ProxeusFSContract *ProxeusFSContractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.ProxeusFSContractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ProxeusFSContract *ProxeusFSContractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.ProxeusFSContractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_ProxeusFSContract *ProxeusFSContractCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _ProxeusFSContract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_ProxeusFSContract *ProxeusFSContractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_ProxeusFSContract *ProxeusFSContractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.contract.Transact(opts, method, params...)
}

// XESAllowence is a free data retrieval call binding the contract method 0x30bbc643.
//
// Solidity: function XESAllowence(sendr address) constant returns(sum uint256)
func (_ProxeusFSContract *ProxeusFSContractCaller) XESAllowence(opts *bind.CallOpts, sendr common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ProxeusFSContract.contract.Call(opts, out, "XESAllowence", sendr)
	return *ret0, err
}

// XESAllowence is a free data retrieval call binding the contract method 0x30bbc643.
//
// Solidity: function XESAllowence(sendr address) constant returns(sum uint256)
func (_ProxeusFSContract *ProxeusFSContractSession) XESAllowence(sendr common.Address) (*big.Int, error) {
	return _ProxeusFSContract.Contract.XESAllowence(&_ProxeusFSContract.CallOpts, sendr)
}

// XESAllowence is a free data retrieval call binding the contract method 0x30bbc643.
//
// Solidity: function XESAllowence(sendr address) constant returns(sum uint256)
func (_ProxeusFSContract *ProxeusFSContractCallerSession) XESAllowence(sendr common.Address) (*big.Int, error) {
	return _ProxeusFSContract.Contract.XESAllowence(&_ProxeusFSContract.CallOpts, sendr)
}

// XESAmountPerFile is a free data retrieval call binding the contract method 0x61286a3f.
//
// Solidity: function XESAmountPerFile(prvs address[]) constant returns(sum uint256)
func (_ProxeusFSContract *ProxeusFSContractCaller) XESAmountPerFile(opts *bind.CallOpts, prvs []common.Address) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ProxeusFSContract.contract.Call(opts, out, "XESAmountPerFile", prvs)
	return *ret0, err
}

// XESAmountPerFile is a free data retrieval call binding the contract method 0x61286a3f.
//
// Solidity: function XESAmountPerFile(prvs address[]) constant returns(sum uint256)
func (_ProxeusFSContract *ProxeusFSContractSession) XESAmountPerFile(prvs []common.Address) (*big.Int, error) {
	return _ProxeusFSContract.Contract.XESAmountPerFile(&_ProxeusFSContract.CallOpts, prvs)
}

// XESAmountPerFile is a free data retrieval call binding the contract method 0x61286a3f.
//
// Solidity: function XESAmountPerFile(prvs address[]) constant returns(sum uint256)
func (_ProxeusFSContract *ProxeusFSContractCallerSession) XESAmountPerFile(prvs []common.Address) (*big.Int, error) {
	return _ProxeusFSContract.Contract.XESAmountPerFile(&_ProxeusFSContract.CallOpts, prvs)
}

// DappVersion is a free data retrieval call binding the contract method 0x73a3c6cc.
//
// Solidity: function dappVersion() constant returns(bytes32)
func (_ProxeusFSContract *ProxeusFSContractCaller) DappVersion(opts *bind.CallOpts) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _ProxeusFSContract.contract.Call(opts, out, "dappVersion")
	return *ret0, err
}

// DappVersion is a free data retrieval call binding the contract method 0x73a3c6cc.
//
// Solidity: function dappVersion() constant returns(bytes32)
func (_ProxeusFSContract *ProxeusFSContractSession) DappVersion() ([32]byte, error) {
	return _ProxeusFSContract.Contract.DappVersion(&_ProxeusFSContract.CallOpts)
}

// DappVersion is a free data retrieval call binding the contract method 0x73a3c6cc.
//
// Solidity: function dappVersion() constant returns(bytes32)
func (_ProxeusFSContract *ProxeusFSContractCallerSession) DappVersion() ([32]byte, error) {
	return _ProxeusFSContract.Contract.DappVersion(&_ProxeusFSContract.CallOpts)
}

// FileExpiry is a free data retrieval call binding the contract method 0xbb90b9fb.
//
// Solidity: function fileExpiry(hash bytes32) constant returns(uint256)
func (_ProxeusFSContract *ProxeusFSContractCaller) FileExpiry(opts *bind.CallOpts, hash [32]byte) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ProxeusFSContract.contract.Call(opts, out, "fileExpiry", hash)
	return *ret0, err
}

// FileExpiry is a free data retrieval call binding the contract method 0xbb90b9fb.
//
// Solidity: function fileExpiry(hash bytes32) constant returns(uint256)
func (_ProxeusFSContract *ProxeusFSContractSession) FileExpiry(hash [32]byte) (*big.Int, error) {
	return _ProxeusFSContract.Contract.FileExpiry(&_ProxeusFSContract.CallOpts, hash)
}

// FileExpiry is a free data retrieval call binding the contract method 0xbb90b9fb.
//
// Solidity: function fileExpiry(hash bytes32) constant returns(uint256)
func (_ProxeusFSContract *ProxeusFSContractCallerSession) FileExpiry(hash [32]byte) (*big.Int, error) {
	return _ProxeusFSContract.Contract.FileExpiry(&_ProxeusFSContract.CallOpts, hash)
}

// FileGetPerm is a free data retrieval call binding the contract method 0x613ec716.
//
// Solidity: function fileGetPerm(hash bytes32, addr address, write bool) constant returns(bool)
func (_ProxeusFSContract *ProxeusFSContractCaller) FileGetPerm(opts *bind.CallOpts, hash [32]byte, addr common.Address, write bool) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _ProxeusFSContract.contract.Call(opts, out, "fileGetPerm", hash, addr, write)
	return *ret0, err
}

// FileGetPerm is a free data retrieval call binding the contract method 0x613ec716.
//
// Solidity: function fileGetPerm(hash bytes32, addr address, write bool) constant returns(bool)
func (_ProxeusFSContract *ProxeusFSContractSession) FileGetPerm(hash [32]byte, addr common.Address, write bool) (bool, error) {
	return _ProxeusFSContract.Contract.FileGetPerm(&_ProxeusFSContract.CallOpts, hash, addr, write)
}

// FileGetPerm is a free data retrieval call binding the contract method 0x613ec716.
//
// Solidity: function fileGetPerm(hash bytes32, addr address, write bool) constant returns(bool)
func (_ProxeusFSContract *ProxeusFSContractCallerSession) FileGetPerm(hash [32]byte, addr common.Address, write bool) (bool, error) {
	return _ProxeusFSContract.Contract.FileGetPerm(&_ProxeusFSContract.CallOpts, hash, addr, write)
}

// FileHasSP is a free data retrieval call binding the contract method 0x444ed7b1.
//
// Solidity: function fileHasSP(hash bytes32, addr address) constant returns(bool)
func (_ProxeusFSContract *ProxeusFSContractCaller) FileHasSP(opts *bind.CallOpts, hash [32]byte, addr common.Address) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _ProxeusFSContract.contract.Call(opts, out, "fileHasSP", hash, addr)
	return *ret0, err
}

// FileHasSP is a free data retrieval call binding the contract method 0x444ed7b1.
//
// Solidity: function fileHasSP(hash bytes32, addr address) constant returns(bool)
func (_ProxeusFSContract *ProxeusFSContractSession) FileHasSP(hash [32]byte, addr common.Address) (bool, error) {
	return _ProxeusFSContract.Contract.FileHasSP(&_ProxeusFSContract.CallOpts, hash, addr)
}

// FileHasSP is a free data retrieval call binding the contract method 0x444ed7b1.
//
// Solidity: function fileHasSP(hash bytes32, addr address) constant returns(bool)
func (_ProxeusFSContract *ProxeusFSContractCallerSession) FileHasSP(hash [32]byte, addr common.Address) (bool, error) {
	return _ProxeusFSContract.Contract.FileHasSP(&_ProxeusFSContract.CallOpts, hash, addr)
}

// FileInfo is a free data retrieval call binding the contract method 0x9754c8fd.
//
// Solidity: function fileInfo(hash bytes32) constant returns(id bytes32, ownr address, fileType uint256, removed bool, expiry uint256, isPublic bool, thumbnailHash bytes32, fparent bytes32, replacesFile bytes32, readAccess address[], definedSigners address[])
func (_ProxeusFSContract *ProxeusFSContractCaller) FileInfo(opts *bind.CallOpts, hash [32]byte) (struct {
	Id             [32]byte
	Ownr           common.Address
	FileType       *big.Int
	Removed        bool
	Expiry         *big.Int
	IsPublic       bool
	ThumbnailHash  [32]byte
	Fparent        [32]byte
	ReplacesFile   [32]byte
	ReadAccess     []common.Address
	DefinedSigners []common.Address
}, error) {
	ret := new(struct {
		Id             [32]byte
		Ownr           common.Address
		FileType       *big.Int
		Removed        bool
		Expiry         *big.Int
		IsPublic       bool
		ThumbnailHash  [32]byte
		Fparent        [32]byte
		ReplacesFile   [32]byte
		ReadAccess     []common.Address
		DefinedSigners []common.Address
	})
	out := ret
	err := _ProxeusFSContract.contract.Call(opts, out, "fileInfo", hash)
	return *ret, err
}

// FileInfo is a free data retrieval call binding the contract method 0x9754c8fd.
//
// Solidity: function fileInfo(hash bytes32) constant returns(id bytes32, ownr address, fileType uint256, removed bool, expiry uint256, isPublic bool, thumbnailHash bytes32, fparent bytes32, replacesFile bytes32, readAccess address[], definedSigners address[])
func (_ProxeusFSContract *ProxeusFSContractSession) FileInfo(hash [32]byte) (struct {
	Id             [32]byte
	Ownr           common.Address
	FileType       *big.Int
	Removed        bool
	Expiry         *big.Int
	IsPublic       bool
	ThumbnailHash  [32]byte
	Fparent        [32]byte
	ReplacesFile   [32]byte
	ReadAccess     []common.Address
	DefinedSigners []common.Address
}, error) {
	return _ProxeusFSContract.Contract.FileInfo(&_ProxeusFSContract.CallOpts, hash)
}

// FileInfo is a free data retrieval call binding the contract method 0x9754c8fd.
//
// Solidity: function fileInfo(hash bytes32) constant returns(id bytes32, ownr address, fileType uint256, removed bool, expiry uint256, isPublic bool, thumbnailHash bytes32, fparent bytes32, replacesFile bytes32, readAccess address[], definedSigners address[])
func (_ProxeusFSContract *ProxeusFSContractCallerSession) FileInfo(hash [32]byte) (struct {
	Id             [32]byte
	Ownr           common.Address
	FileType       *big.Int
	Removed        bool
	Expiry         *big.Int
	IsPublic       bool
	ThumbnailHash  [32]byte
	Fparent        [32]byte
	ReplacesFile   [32]byte
	ReadAccess     []common.Address
	DefinedSigners []common.Address
}, error) {
	return _ProxeusFSContract.Contract.FileInfo(&_ProxeusFSContract.CallOpts, hash)
}

// FileList is a free data retrieval call binding the contract method 0xff66135e.
//
// Solidity: function fileList() constant returns(bytes32[])
func (_ProxeusFSContract *ProxeusFSContractCaller) FileList(opts *bind.CallOpts) ([][32]byte, error) {
	var (
		ret0 = new([][32]byte)
	)
	out := ret0
	err := _ProxeusFSContract.contract.Call(opts, out, "fileList")
	return *ret0, err
}

// FileList is a free data retrieval call binding the contract method 0xff66135e.
//
// Solidity: function fileList() constant returns(bytes32[])
func (_ProxeusFSContract *ProxeusFSContractSession) FileList() ([][32]byte, error) {
	return _ProxeusFSContract.Contract.FileList(&_ProxeusFSContract.CallOpts)
}

// FileList is a free data retrieval call binding the contract method 0xff66135e.
//
// Solidity: function fileList() constant returns(bytes32[])
func (_ProxeusFSContract *ProxeusFSContractCallerSession) FileList() ([][32]byte, error) {
	return _ProxeusFSContract.Contract.FileList(&_ProxeusFSContract.CallOpts)
}

// FileSigners is a free data retrieval call binding the contract method 0x411c5f61.
//
// Solidity: function fileSigners(hash bytes32) constant returns(address[])
func (_ProxeusFSContract *ProxeusFSContractCaller) FileSigners(opts *bind.CallOpts, hash [32]byte) ([]common.Address, error) {
	var (
		ret0 = new([]common.Address)
	)
	out := ret0
	err := _ProxeusFSContract.contract.Call(opts, out, "fileSigners", hash)
	return *ret0, err
}

// FileSigners is a free data retrieval call binding the contract method 0x411c5f61.
//
// Solidity: function fileSigners(hash bytes32) constant returns(address[])
func (_ProxeusFSContract *ProxeusFSContractSession) FileSigners(hash [32]byte) ([]common.Address, error) {
	return _ProxeusFSContract.Contract.FileSigners(&_ProxeusFSContract.CallOpts, hash)
}

// FileSigners is a free data retrieval call binding the contract method 0x411c5f61.
//
// Solidity: function fileSigners(hash bytes32) constant returns(address[])
func (_ProxeusFSContract *ProxeusFSContractCallerSession) FileSigners(hash [32]byte) ([]common.Address, error) {
	return _ProxeusFSContract.Contract.FileSigners(&_ProxeusFSContract.CallOpts, hash)
}

// FileSignersCount is a free data retrieval call binding the contract method 0xc5ace443.
//
// Solidity: function fileSignersCount(hash bytes32) constant returns(uint256)
func (_ProxeusFSContract *ProxeusFSContractCaller) FileSignersCount(opts *bind.CallOpts, hash [32]byte) (*big.Int, error) {
	var (
		ret0 = new(*big.Int)
	)
	out := ret0
	err := _ProxeusFSContract.contract.Call(opts, out, "fileSignersCount", hash)
	return *ret0, err
}

// FileSignersCount is a free data retrieval call binding the contract method 0xc5ace443.
//
// Solidity: function fileSignersCount(hash bytes32) constant returns(uint256)
func (_ProxeusFSContract *ProxeusFSContractSession) FileSignersCount(hash [32]byte) (*big.Int, error) {
	return _ProxeusFSContract.Contract.FileSignersCount(&_ProxeusFSContract.CallOpts, hash)
}

// FileSignersCount is a free data retrieval call binding the contract method 0xc5ace443.
//
// Solidity: function fileSignersCount(hash bytes32) constant returns(uint256)
func (_ProxeusFSContract *ProxeusFSContractCallerSession) FileSignersCount(hash [32]byte) (*big.Int, error) {
	return _ProxeusFSContract.Contract.FileSignersCount(&_ProxeusFSContract.CallOpts, hash)
}

// FileVerify is a free data retrieval call binding the contract method 0x293c8e2e.
//
// Solidity: function fileVerify(hash bytes32) constant returns(bool, address[])
func (_ProxeusFSContract *ProxeusFSContractCaller) FileVerify(opts *bind.CallOpts, hash [32]byte) (bool, []common.Address, error) {
	var (
		ret0 = new(bool)
		ret1 = new([]common.Address)
	)
	out := &[]interface{}{
		ret0,
		ret1,
	}
	err := _ProxeusFSContract.contract.Call(opts, out, "fileVerify", hash)
	return *ret0, *ret1, err
}

// FileVerify is a free data retrieval call binding the contract method 0x293c8e2e.
//
// Solidity: function fileVerify(hash bytes32) constant returns(bool, address[])
func (_ProxeusFSContract *ProxeusFSContractSession) FileVerify(hash [32]byte) (bool, []common.Address, error) {
	return _ProxeusFSContract.Contract.FileVerify(&_ProxeusFSContract.CallOpts, hash)
}

// FileVerify is a free data retrieval call binding the contract method 0x293c8e2e.
//
// Solidity: function fileVerify(hash bytes32) constant returns(bool, address[])
func (_ProxeusFSContract *ProxeusFSContractCallerSession) FileVerify(hash [32]byte) (bool, []common.Address, error) {
	return _ProxeusFSContract.Contract.FileVerify(&_ProxeusFSContract.CallOpts, hash)
}

// IsFileRemoved is a free data retrieval call binding the contract method 0xc559d6cc.
//
// Solidity: function isFileRemoved(hash bytes32) constant returns(bool)
func (_ProxeusFSContract *ProxeusFSContractCaller) IsFileRemoved(opts *bind.CallOpts, hash [32]byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _ProxeusFSContract.contract.Call(opts, out, "isFileRemoved", hash)
	return *ret0, err
}

// IsFileRemoved is a free data retrieval call binding the contract method 0xc559d6cc.
//
// Solidity: function isFileRemoved(hash bytes32) constant returns(bool)
func (_ProxeusFSContract *ProxeusFSContractSession) IsFileRemoved(hash [32]byte) (bool, error) {
	return _ProxeusFSContract.Contract.IsFileRemoved(&_ProxeusFSContract.CallOpts, hash)
}

// IsFileRemoved is a free data retrieval call binding the contract method 0xc559d6cc.
//
// Solidity: function isFileRemoved(hash bytes32) constant returns(bool)
func (_ProxeusFSContract *ProxeusFSContractCallerSession) IsFileRemoved(hash [32]byte) (bool, error) {
	return _ProxeusFSContract.Contract.IsFileRemoved(&_ProxeusFSContract.CallOpts, hash)
}

// SpInfo is a free data retrieval call binding the contract method 0x10b6b12e.
//
// Solidity: function spInfo(strPrv address) constant returns(urlPrefix bytes32)
func (_ProxeusFSContract *ProxeusFSContractCaller) SpInfo(opts *bind.CallOpts, strPrv common.Address) ([32]byte, error) {
	var (
		ret0 = new([32]byte)
	)
	out := ret0
	err := _ProxeusFSContract.contract.Call(opts, out, "spInfo", strPrv)
	return *ret0, err
}

// SpInfo is a free data retrieval call binding the contract method 0x10b6b12e.
//
// Solidity: function spInfo(strPrv address) constant returns(urlPrefix bytes32)
func (_ProxeusFSContract *ProxeusFSContractSession) SpInfo(strPrv common.Address) ([32]byte, error) {
	return _ProxeusFSContract.Contract.SpInfo(&_ProxeusFSContract.CallOpts, strPrv)
}

// SpInfo is a free data retrieval call binding the contract method 0x10b6b12e.
//
// Solidity: function spInfo(strPrv address) constant returns(urlPrefix bytes32)
func (_ProxeusFSContract *ProxeusFSContractCallerSession) SpInfo(strPrv common.Address) ([32]byte, error) {
	return _ProxeusFSContract.Contract.SpInfo(&_ProxeusFSContract.CallOpts, strPrv)
}

// SpList is a free data retrieval call binding the contract method 0x0361b949.
//
// Solidity: function spList() constant returns(address[])
func (_ProxeusFSContract *ProxeusFSContractCaller) SpList(opts *bind.CallOpts) ([]common.Address, error) {
	var (
		ret0 = new([]common.Address)
	)
	out := ret0
	err := _ProxeusFSContract.contract.Call(opts, out, "spList")
	return *ret0, err
}

// SpList is a free data retrieval call binding the contract method 0x0361b949.
//
// Solidity: function spList() constant returns(address[])
func (_ProxeusFSContract *ProxeusFSContractSession) SpList() ([]common.Address, error) {
	return _ProxeusFSContract.Contract.SpList(&_ProxeusFSContract.CallOpts)
}

// SpList is a free data retrieval call binding the contract method 0x0361b949.
//
// Solidity: function spList() constant returns(address[])
func (_ProxeusFSContract *ProxeusFSContractCallerSession) SpList() ([]common.Address, error) {
	return _ProxeusFSContract.Contract.SpList(&_ProxeusFSContract.CallOpts)
}

// CreateFileDefinedSigners is a paid mutator transaction binding the contract method 0xba890dc3.
//
// Solidity: function createFileDefinedSigners(hash bytes32, definedSigners address[], expiry uint256, replacesFile bytes32, prvs address[]) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactor) CreateFileDefinedSigners(opts *bind.TransactOpts, hash [32]byte, definedSigners []common.Address, expiry *big.Int, replacesFile [32]byte, prvs []common.Address) (*types.Transaction, error) {
	return _ProxeusFSContract.contract.Transact(opts, "createFileDefinedSigners", hash, definedSigners, expiry, replacesFile, prvs)
}

// CreateFileDefinedSigners is a paid mutator transaction binding the contract method 0xba890dc3.
//
// Solidity: function createFileDefinedSigners(hash bytes32, definedSigners address[], expiry uint256, replacesFile bytes32, prvs address[]) returns()
func (_ProxeusFSContract *ProxeusFSContractSession) CreateFileDefinedSigners(hash [32]byte, definedSigners []common.Address, expiry *big.Int, replacesFile [32]byte, prvs []common.Address) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.CreateFileDefinedSigners(&_ProxeusFSContract.TransactOpts, hash, definedSigners, expiry, replacesFile, prvs)
}

// CreateFileDefinedSigners is a paid mutator transaction binding the contract method 0xba890dc3.
//
// Solidity: function createFileDefinedSigners(hash bytes32, definedSigners address[], expiry uint256, replacesFile bytes32, prvs address[]) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactorSession) CreateFileDefinedSigners(hash [32]byte, definedSigners []common.Address, expiry *big.Int, replacesFile [32]byte, prvs []common.Address) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.CreateFileDefinedSigners(&_ProxeusFSContract.TransactOpts, hash, definedSigners, expiry, replacesFile, prvs)
}

// CreateFileThumbnail is a paid mutator transaction binding the contract method 0xa8efa269.
//
// Solidity: function createFileThumbnail(hash bytes32, pParent bytes32, pPublic bool) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactor) CreateFileThumbnail(opts *bind.TransactOpts, hash [32]byte, pParent [32]byte, pPublic bool) (*types.Transaction, error) {
	return _ProxeusFSContract.contract.Transact(opts, "createFileThumbnail", hash, pParent, pPublic)
}

// CreateFileThumbnail is a paid mutator transaction binding the contract method 0xa8efa269.
//
// Solidity: function createFileThumbnail(hash bytes32, pParent bytes32, pPublic bool) returns()
func (_ProxeusFSContract *ProxeusFSContractSession) CreateFileThumbnail(hash [32]byte, pParent [32]byte, pPublic bool) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.CreateFileThumbnail(&_ProxeusFSContract.TransactOpts, hash, pParent, pPublic)
}

// CreateFileThumbnail is a paid mutator transaction binding the contract method 0xa8efa269.
//
// Solidity: function createFileThumbnail(hash bytes32, pParent bytes32, pPublic bool) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactorSession) CreateFileThumbnail(hash [32]byte, pParent [32]byte, pPublic bool) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.CreateFileThumbnail(&_ProxeusFSContract.TransactOpts, hash, pParent, pPublic)
}

// CreateFileUndefinedSigners is a paid mutator transaction binding the contract method 0x76189601.
//
// Solidity: function createFileUndefinedSigners(hash bytes32, mandatorySigners uint256, expiry uint256, replacesFile bytes32, prvs address[]) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactor) CreateFileUndefinedSigners(opts *bind.TransactOpts, hash [32]byte, mandatorySigners *big.Int, expiry *big.Int, replacesFile [32]byte, prvs []common.Address) (*types.Transaction, error) {
	return _ProxeusFSContract.contract.Transact(opts, "createFileUndefinedSigners", hash, mandatorySigners, expiry, replacesFile, prvs)
}

// CreateFileUndefinedSigners is a paid mutator transaction binding the contract method 0x76189601.
//
// Solidity: function createFileUndefinedSigners(hash bytes32, mandatorySigners uint256, expiry uint256, replacesFile bytes32, prvs address[]) returns()
func (_ProxeusFSContract *ProxeusFSContractSession) CreateFileUndefinedSigners(hash [32]byte, mandatorySigners *big.Int, expiry *big.Int, replacesFile [32]byte, prvs []common.Address) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.CreateFileUndefinedSigners(&_ProxeusFSContract.TransactOpts, hash, mandatorySigners, expiry, replacesFile, prvs)
}

// CreateFileUndefinedSigners is a paid mutator transaction binding the contract method 0x76189601.
//
// Solidity: function createFileUndefinedSigners(hash bytes32, mandatorySigners uint256, expiry uint256, replacesFile bytes32, prvs address[]) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactorSession) CreateFileUndefinedSigners(hash [32]byte, mandatorySigners *big.Int, expiry *big.Int, replacesFile [32]byte, prvs []common.Address) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.CreateFileUndefinedSigners(&_ProxeusFSContract.TransactOpts, hash, mandatorySigners, expiry, replacesFile, prvs)
}

// FileAddSP is a paid mutator transaction binding the contract method 0xc00aa3f6.
//
// Solidity: function fileAddSP(hash bytes32, strPrv address) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactor) FileAddSP(opts *bind.TransactOpts, hash [32]byte, strPrv common.Address) (*types.Transaction, error) {
	return _ProxeusFSContract.contract.Transact(opts, "fileAddSP", hash, strPrv)
}

// FileAddSP is a paid mutator transaction binding the contract method 0xc00aa3f6.
//
// Solidity: function fileAddSP(hash bytes32, strPrv address) returns()
func (_ProxeusFSContract *ProxeusFSContractSession) FileAddSP(hash [32]byte, strPrv common.Address) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.FileAddSP(&_ProxeusFSContract.TransactOpts, hash, strPrv)
}

// FileAddSP is a paid mutator transaction binding the contract method 0xc00aa3f6.
//
// Solidity: function fileAddSP(hash bytes32, strPrv address) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactorSession) FileAddSP(hash [32]byte, strPrv common.Address) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.FileAddSP(&_ProxeusFSContract.TransactOpts, hash, strPrv)
}

// FileRemove is a paid mutator transaction binding the contract method 0x6ca3b7b6.
//
// Solidity: function fileRemove(hash bytes32) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactor) FileRemove(opts *bind.TransactOpts, hash [32]byte) (*types.Transaction, error) {
	return _ProxeusFSContract.contract.Transact(opts, "fileRemove", hash)
}

// FileRemove is a paid mutator transaction binding the contract method 0x6ca3b7b6.
//
// Solidity: function fileRemove(hash bytes32) returns()
func (_ProxeusFSContract *ProxeusFSContractSession) FileRemove(hash [32]byte) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.FileRemove(&_ProxeusFSContract.TransactOpts, hash)
}

// FileRemove is a paid mutator transaction binding the contract method 0x6ca3b7b6.
//
// Solidity: function fileRemove(hash bytes32) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactorSession) FileRemove(hash [32]byte) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.FileRemove(&_ProxeusFSContract.TransactOpts, hash)
}

// FileRequestAccess is a paid mutator transaction binding the contract method 0x82d9f1c2.
//
// Solidity: function fileRequestAccess(hash bytes32) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactor) FileRequestAccess(opts *bind.TransactOpts, hash [32]byte) (*types.Transaction, error) {
	return _ProxeusFSContract.contract.Transact(opts, "fileRequestAccess", hash)
}

// FileRequestAccess is a paid mutator transaction binding the contract method 0x82d9f1c2.
//
// Solidity: function fileRequestAccess(hash bytes32) returns()
func (_ProxeusFSContract *ProxeusFSContractSession) FileRequestAccess(hash [32]byte) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.FileRequestAccess(&_ProxeusFSContract.TransactOpts, hash)
}

// FileRequestAccess is a paid mutator transaction binding the contract method 0x82d9f1c2.
//
// Solidity: function fileRequestAccess(hash bytes32) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactorSession) FileRequestAccess(hash [32]byte) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.FileRequestAccess(&_ProxeusFSContract.TransactOpts, hash)
}

// FileRequestSign is a paid mutator transaction binding the contract method 0x6fafeb3b.
//
// Solidity: function fileRequestSign(hash bytes32, signer address[]) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactor) FileRequestSign(opts *bind.TransactOpts, hash [32]byte, signer []common.Address) (*types.Transaction, error) {
	return _ProxeusFSContract.contract.Transact(opts, "fileRequestSign", hash, signer)
}

// FileRequestSign is a paid mutator transaction binding the contract method 0x6fafeb3b.
//
// Solidity: function fileRequestSign(hash bytes32, signer address[]) returns()
func (_ProxeusFSContract *ProxeusFSContractSession) FileRequestSign(hash [32]byte, signer []common.Address) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.FileRequestSign(&_ProxeusFSContract.TransactOpts, hash, signer)
}

// FileRequestSign is a paid mutator transaction binding the contract method 0x6fafeb3b.
//
// Solidity: function fileRequestSign(hash bytes32, signer address[]) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactorSession) FileRequestSign(hash [32]byte, signer []common.Address) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.FileRequestSign(&_ProxeusFSContract.TransactOpts, hash, signer)
}

// FileRevokePerm is a paid mutator transaction binding the contract method 0xdd4877c5.
//
// Solidity: function fileRevokePerm(hash bytes32, addr address[]) returns(bool)
func (_ProxeusFSContract *ProxeusFSContractTransactor) FileRevokePerm(opts *bind.TransactOpts, hash [32]byte, addr []common.Address) (*types.Transaction, error) {
	return _ProxeusFSContract.contract.Transact(opts, "fileRevokePerm", hash, addr)
}

// FileRevokePerm is a paid mutator transaction binding the contract method 0xdd4877c5.
//
// Solidity: function fileRevokePerm(hash bytes32, addr address[]) returns(bool)
func (_ProxeusFSContract *ProxeusFSContractSession) FileRevokePerm(hash [32]byte, addr []common.Address) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.FileRevokePerm(&_ProxeusFSContract.TransactOpts, hash, addr)
}

// FileRevokePerm is a paid mutator transaction binding the contract method 0xdd4877c5.
//
// Solidity: function fileRevokePerm(hash bytes32, addr address[]) returns(bool)
func (_ProxeusFSContract *ProxeusFSContractTransactorSession) FileRevokePerm(hash [32]byte, addr []common.Address) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.FileRevokePerm(&_ProxeusFSContract.TransactOpts, hash, addr)
}

// FileSetPerm is a paid mutator transaction binding the contract method 0x6a732ffe.
//
// Solidity: function fileSetPerm(hash bytes32, addr address[]) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactor) FileSetPerm(opts *bind.TransactOpts, hash [32]byte, addr []common.Address) (*types.Transaction, error) {
	return _ProxeusFSContract.contract.Transact(opts, "fileSetPerm", hash, addr)
}

// FileSetPerm is a paid mutator transaction binding the contract method 0x6a732ffe.
//
// Solidity: function fileSetPerm(hash bytes32, addr address[]) returns()
func (_ProxeusFSContract *ProxeusFSContractSession) FileSetPerm(hash [32]byte, addr []common.Address) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.FileSetPerm(&_ProxeusFSContract.TransactOpts, hash, addr)
}

// FileSetPerm is a paid mutator transaction binding the contract method 0x6a732ffe.
//
// Solidity: function fileSetPerm(hash bytes32, addr address[]) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactorSession) FileSetPerm(hash [32]byte, addr []common.Address) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.FileSetPerm(&_ProxeusFSContract.TransactOpts, hash, addr)
}

// FileSign is a paid mutator transaction binding the contract method 0x274b15a6.
//
// Solidity: function fileSign(hash bytes32) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactor) FileSign(opts *bind.TransactOpts, hash [32]byte) (*types.Transaction, error) {
	return _ProxeusFSContract.contract.Transact(opts, "fileSign", hash)
}

// FileSign is a paid mutator transaction binding the contract method 0x274b15a6.
//
// Solidity: function fileSign(hash bytes32) returns()
func (_ProxeusFSContract *ProxeusFSContractSession) FileSign(hash [32]byte) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.FileSign(&_ProxeusFSContract.TransactOpts, hash)
}

// FileSign is a paid mutator transaction binding the contract method 0x274b15a6.
//
// Solidity: function fileSign(hash bytes32) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactorSession) FileSign(hash [32]byte) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.FileSign(&_ProxeusFSContract.TransactOpts, hash)
}

// SetDappVersion is a paid mutator transaction binding the contract method 0x4f61b021.
//
// Solidity: function setDappVersion(version bytes32) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactor) SetDappVersion(opts *bind.TransactOpts, version [32]byte) (*types.Transaction, error) {
	return _ProxeusFSContract.contract.Transact(opts, "setDappVersion", version)
}

// SetDappVersion is a paid mutator transaction binding the contract method 0x4f61b021.
//
// Solidity: function setDappVersion(version bytes32) returns()
func (_ProxeusFSContract *ProxeusFSContractSession) SetDappVersion(version [32]byte) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.SetDappVersion(&_ProxeusFSContract.TransactOpts, version)
}

// SetDappVersion is a paid mutator transaction binding the contract method 0x4f61b021.
//
// Solidity: function setDappVersion(version bytes32) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactorSession) SetDappVersion(version [32]byte) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.SetDappVersion(&_ProxeusFSContract.TransactOpts, version)
}

// SetEternalStorage is a paid mutator transaction binding the contract method 0xb5bb5619.
//
// Solidity: function setEternalStorage(_eternalstorage address) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactor) SetEternalStorage(opts *bind.TransactOpts, _eternalstorage common.Address) (*types.Transaction, error) {
	return _ProxeusFSContract.contract.Transact(opts, "setEternalStorage", _eternalstorage)
}

// SetEternalStorage is a paid mutator transaction binding the contract method 0xb5bb5619.
//
// Solidity: function setEternalStorage(_eternalstorage address) returns()
func (_ProxeusFSContract *ProxeusFSContractSession) SetEternalStorage(_eternalstorage common.Address) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.SetEternalStorage(&_ProxeusFSContract.TransactOpts, _eternalstorage)
}

// SetEternalStorage is a paid mutator transaction binding the contract method 0xb5bb5619.
//
// Solidity: function setEternalStorage(_eternalstorage address) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactorSession) SetEternalStorage(_eternalstorage common.Address) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.SetEternalStorage(&_ProxeusFSContract.TransactOpts, _eternalstorage)
}

// SpAdd is a paid mutator transaction binding the contract method 0x54eb5e67.
//
// Solidity: function spAdd(strProv address, urlPrefix bytes32) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactor) SpAdd(opts *bind.TransactOpts, strProv common.Address, urlPrefix [32]byte) (*types.Transaction, error) {
	return _ProxeusFSContract.contract.Transact(opts, "spAdd", strProv, urlPrefix)
}

// SpAdd is a paid mutator transaction binding the contract method 0x54eb5e67.
//
// Solidity: function spAdd(strProv address, urlPrefix bytes32) returns()
func (_ProxeusFSContract *ProxeusFSContractSession) SpAdd(strProv common.Address, urlPrefix [32]byte) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.SpAdd(&_ProxeusFSContract.TransactOpts, strProv, urlPrefix)
}

// SpAdd is a paid mutator transaction binding the contract method 0x54eb5e67.
//
// Solidity: function spAdd(strProv address, urlPrefix bytes32) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactorSession) SpAdd(strProv common.Address, urlPrefix [32]byte) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.SpAdd(&_ProxeusFSContract.TransactOpts, strProv, urlPrefix)
}

// SpUpdate is a paid mutator transaction binding the contract method 0xeef3798e.
//
// Solidity: function spUpdate(strPrv address, urlPrefix bytes32) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactor) SpUpdate(opts *bind.TransactOpts, strPrv common.Address, urlPrefix [32]byte) (*types.Transaction, error) {
	return _ProxeusFSContract.contract.Transact(opts, "spUpdate", strPrv, urlPrefix)
}

// SpUpdate is a paid mutator transaction binding the contract method 0xeef3798e.
//
// Solidity: function spUpdate(strPrv address, urlPrefix bytes32) returns()
func (_ProxeusFSContract *ProxeusFSContractSession) SpUpdate(strPrv common.Address, urlPrefix [32]byte) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.SpUpdate(&_ProxeusFSContract.TransactOpts, strPrv, urlPrefix)
}

// SpUpdate is a paid mutator transaction binding the contract method 0xeef3798e.
//
// Solidity: function spUpdate(strPrv address, urlPrefix bytes32) returns()
func (_ProxeusFSContract *ProxeusFSContractTransactorSession) SpUpdate(strPrv common.Address, urlPrefix [32]byte) (*types.Transaction, error) {
	return _ProxeusFSContract.Contract.SpUpdate(&_ProxeusFSContract.TransactOpts, strPrv, urlPrefix)
}

// ProxeusFSContractDeletedIterator is returned from FilterDeleted and is used to iterate over the raw logs and unpacked data for Deleted events raised by the ProxeusFSContract contract.
type ProxeusFSContractDeletedIterator struct {
	Event *ProxeusFSContractDeleted // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProxeusFSContractDeletedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProxeusFSContractDeleted)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProxeusFSContractDeleted)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProxeusFSContractDeletedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProxeusFSContractDeletedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProxeusFSContractDeleted represents a Deleted event raised by the ProxeusFSContract contract.
type ProxeusFSContractDeleted struct {
	Hash [32]byte
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterDeleted is a free log retrieval operation binding the contract event 0x5c098fc6091f34b53be138bd13af35c3082d14080fd720c188e0db7a50820473.
//
// Solidity: e Deleted(hash indexed bytes32)
func (_ProxeusFSContract *ProxeusFSContractFilterer) FilterDeleted(opts *bind.FilterOpts, hash [][32]byte) (*ProxeusFSContractDeletedIterator, error) {

	var hashRule []interface{}
	for _, hashItem := range hash {
		hashRule = append(hashRule, hashItem)
	}

	logs, sub, err := _ProxeusFSContract.contract.FilterLogs(opts, "Deleted", hashRule)
	if err != nil {
		return nil, err
	}
	return &ProxeusFSContractDeletedIterator{contract: _ProxeusFSContract.contract, event: "Deleted", logs: logs, sub: sub}, nil
}

// WatchDeleted is a free log subscription operation binding the contract event 0x5c098fc6091f34b53be138bd13af35c3082d14080fd720c188e0db7a50820473.
//
// Solidity: e Deleted(hash indexed bytes32)
func (_ProxeusFSContract *ProxeusFSContractFilterer) WatchDeleted(opts *bind.WatchOpts, sink chan<- *ProxeusFSContractDeleted, hash [][32]byte) (event.Subscription, error) {

	var hashRule []interface{}
	for _, hashItem := range hash {
		hashRule = append(hashRule, hashItem)
	}

	logs, sub, err := _ProxeusFSContract.contract.WatchLogs(opts, "Deleted", hashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProxeusFSContractDeleted)
				if err := _ProxeusFSContract.contract.UnpackLog(event, "Deleted", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ProxeusFSContractNotifySignIterator is returned from FilterNotifySign and is used to iterate over the raw logs and unpacked data for NotifySign events raised by the ProxeusFSContract contract.
type ProxeusFSContractNotifySignIterator struct {
	Event *ProxeusFSContractNotifySign // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProxeusFSContractNotifySignIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProxeusFSContractNotifySign)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProxeusFSContractNotifySign)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProxeusFSContractNotifySignIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProxeusFSContractNotifySignIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProxeusFSContractNotifySign represents a NotifySign event raised by the ProxeusFSContract contract.
type ProxeusFSContractNotifySign struct {
	Hash [32]byte
	Who  common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterNotifySign is a free log retrieval operation binding the contract event 0x09a81450c56f699ff59c6e12b672e0536f5f93fc92cc2c8cd3eb1df5a2899340.
//
// Solidity: e NotifySign(hash indexed bytes32, who indexed address)
func (_ProxeusFSContract *ProxeusFSContractFilterer) FilterNotifySign(opts *bind.FilterOpts, hash [][32]byte, who []common.Address) (*ProxeusFSContractNotifySignIterator, error) {

	var hashRule []interface{}
	for _, hashItem := range hash {
		hashRule = append(hashRule, hashItem)
	}
	var whoRule []interface{}
	for _, whoItem := range who {
		whoRule = append(whoRule, whoItem)
	}

	logs, sub, err := _ProxeusFSContract.contract.FilterLogs(opts, "NotifySign", hashRule, whoRule)
	if err != nil {
		return nil, err
	}
	return &ProxeusFSContractNotifySignIterator{contract: _ProxeusFSContract.contract, event: "NotifySign", logs: logs, sub: sub}, nil
}

// WatchNotifySign is a free log subscription operation binding the contract event 0x09a81450c56f699ff59c6e12b672e0536f5f93fc92cc2c8cd3eb1df5a2899340.
//
// Solidity: e NotifySign(hash indexed bytes32, who indexed address)
func (_ProxeusFSContract *ProxeusFSContractFilterer) WatchNotifySign(opts *bind.WatchOpts, sink chan<- *ProxeusFSContractNotifySign, hash [][32]byte, who []common.Address) (event.Subscription, error) {

	var hashRule []interface{}
	for _, hashItem := range hash {
		hashRule = append(hashRule, hashItem)
	}
	var whoRule []interface{}
	for _, whoItem := range who {
		whoRule = append(whoRule, whoItem)
	}

	logs, sub, err := _ProxeusFSContract.contract.WatchLogs(opts, "NotifySign", hashRule, whoRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProxeusFSContractNotifySign)
				if err := _ProxeusFSContract.contract.UnpackLog(event, "NotifySign", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ProxeusFSContractOwnerChangedIterator is returned from FilterOwnerChanged and is used to iterate over the raw logs and unpacked data for OwnerChanged events raised by the ProxeusFSContract contract.
type ProxeusFSContractOwnerChangedIterator struct {
	Event *ProxeusFSContractOwnerChanged // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProxeusFSContractOwnerChangedIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProxeusFSContractOwnerChanged)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProxeusFSContractOwnerChanged)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProxeusFSContractOwnerChangedIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProxeusFSContractOwnerChangedIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProxeusFSContractOwnerChanged represents a OwnerChanged event raised by the ProxeusFSContract contract.
type ProxeusFSContractOwnerChanged struct {
	Hash     [32]byte
	OldOwner common.Address
	NewOwner common.Address
	Raw      types.Log // Blockchain specific contextual infos
}

// FilterOwnerChanged is a free log retrieval operation binding the contract event 0x06e9c07310f63759634ddbb7257dbb19ca404f90bd6bdef1d3386fab033cebce.
//
// Solidity: e OwnerChanged(hash indexed bytes32, oldOwner address, newOwner address)
func (_ProxeusFSContract *ProxeusFSContractFilterer) FilterOwnerChanged(opts *bind.FilterOpts, hash [][32]byte) (*ProxeusFSContractOwnerChangedIterator, error) {

	var hashRule []interface{}
	for _, hashItem := range hash {
		hashRule = append(hashRule, hashItem)
	}

	logs, sub, err := _ProxeusFSContract.contract.FilterLogs(opts, "OwnerChanged", hashRule)
	if err != nil {
		return nil, err
	}
	return &ProxeusFSContractOwnerChangedIterator{contract: _ProxeusFSContract.contract, event: "OwnerChanged", logs: logs, sub: sub}, nil
}

// WatchOwnerChanged is a free log subscription operation binding the contract event 0x06e9c07310f63759634ddbb7257dbb19ca404f90bd6bdef1d3386fab033cebce.
//
// Solidity: e OwnerChanged(hash indexed bytes32, oldOwner address, newOwner address)
func (_ProxeusFSContract *ProxeusFSContractFilterer) WatchOwnerChanged(opts *bind.WatchOpts, sink chan<- *ProxeusFSContractOwnerChanged, hash [][32]byte) (event.Subscription, error) {

	var hashRule []interface{}
	for _, hashItem := range hash {
		hashRule = append(hashRule, hashItem)
	}

	logs, sub, err := _ProxeusFSContract.contract.WatchLogs(opts, "OwnerChanged", hashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProxeusFSContractOwnerChanged)
				if err := _ProxeusFSContract.contract.UnpackLog(event, "OwnerChanged", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ProxeusFSContractRequestAccessIterator is returned from FilterRequestAccess and is used to iterate over the raw logs and unpacked data for RequestAccess events raised by the ProxeusFSContract contract.
type ProxeusFSContractRequestAccessIterator struct {
	Event *ProxeusFSContractRequestAccess // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProxeusFSContractRequestAccessIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProxeusFSContractRequestAccess)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProxeusFSContractRequestAccess)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProxeusFSContractRequestAccessIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProxeusFSContractRequestAccessIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProxeusFSContractRequestAccess represents a RequestAccess event raised by the ProxeusFSContract contract.
type ProxeusFSContractRequestAccess struct {
	Hash [32]byte
	Who  common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterRequestAccess is a free log retrieval operation binding the contract event 0xf61ef224bc19aca79d0a8d921f0932f3e3e2bf3253c8d7740a62f8de52d5154a.
//
// Solidity: e RequestAccess(hash bytes32, who address)
func (_ProxeusFSContract *ProxeusFSContractFilterer) FilterRequestAccess(opts *bind.FilterOpts) (*ProxeusFSContractRequestAccessIterator, error) {

	logs, sub, err := _ProxeusFSContract.contract.FilterLogs(opts, "RequestAccess")
	if err != nil {
		return nil, err
	}
	return &ProxeusFSContractRequestAccessIterator{contract: _ProxeusFSContract.contract, event: "RequestAccess", logs: logs, sub: sub}, nil
}

// WatchRequestAccess is a free log subscription operation binding the contract event 0xf61ef224bc19aca79d0a8d921f0932f3e3e2bf3253c8d7740a62f8de52d5154a.
//
// Solidity: e RequestAccess(hash bytes32, who address)
func (_ProxeusFSContract *ProxeusFSContractFilterer) WatchRequestAccess(opts *bind.WatchOpts, sink chan<- *ProxeusFSContractRequestAccess) (event.Subscription, error) {

	logs, sub, err := _ProxeusFSContract.contract.WatchLogs(opts, "RequestAccess")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProxeusFSContractRequestAccess)
				if err := _ProxeusFSContract.contract.UnpackLog(event, "RequestAccess", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ProxeusFSContractRequestSignIterator is returned from FilterRequestSign and is used to iterate over the raw logs and unpacked data for RequestSign events raised by the ProxeusFSContract contract.
type ProxeusFSContractRequestSignIterator struct {
	Event *ProxeusFSContractRequestSign // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProxeusFSContractRequestSignIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProxeusFSContractRequestSign)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProxeusFSContractRequestSign)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProxeusFSContractRequestSignIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProxeusFSContractRequestSignIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProxeusFSContractRequestSign represents a RequestSign event raised by the ProxeusFSContract contract.
type ProxeusFSContractRequestSign struct {
	Hash [32]byte
	To   common.Address
	Raw  types.Log // Blockchain specific contextual infos
}

// FilterRequestSign is a free log retrieval operation binding the contract event 0x1363b0abc70ee6913b8afda36ec223301fc12067a0f6fa9bf042bb2722cf9fee.
//
// Solidity: e RequestSign(hash bytes32, to indexed address)
func (_ProxeusFSContract *ProxeusFSContractFilterer) FilterRequestSign(opts *bind.FilterOpts, to []common.Address) (*ProxeusFSContractRequestSignIterator, error) {

	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _ProxeusFSContract.contract.FilterLogs(opts, "RequestSign", toRule)
	if err != nil {
		return nil, err
	}
	return &ProxeusFSContractRequestSignIterator{contract: _ProxeusFSContract.contract, event: "RequestSign", logs: logs, sub: sub}, nil
}

// WatchRequestSign is a free log subscription operation binding the contract event 0x1363b0abc70ee6913b8afda36ec223301fc12067a0f6fa9bf042bb2722cf9fee.
//
// Solidity: e RequestSign(hash bytes32, to indexed address)
func (_ProxeusFSContract *ProxeusFSContractFilterer) WatchRequestSign(opts *bind.WatchOpts, sink chan<- *ProxeusFSContractRequestSign, to []common.Address) (event.Subscription, error) {

	var toRule []interface{}
	for _, toItem := range to {
		toRule = append(toRule, toItem)
	}

	logs, sub, err := _ProxeusFSContract.contract.WatchLogs(opts, "RequestSign", toRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProxeusFSContractRequestSign)
				if err := _ProxeusFSContract.contract.UnpackLog(event, "RequestSign", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}

// ProxeusFSContractUpdatedEventIterator is returned from FilterUpdatedEvent and is used to iterate over the raw logs and unpacked data for UpdatedEvent events raised by the ProxeusFSContract contract.
type ProxeusFSContractUpdatedEventIterator struct {
	Event *ProxeusFSContractUpdatedEvent // Event containing the contract specifics and raw log

	contract *bind.BoundContract // Generic contract to use for unpacking event data
	event    string              // Event name to use for unpacking event data

	logs chan types.Log        // Log channel receiving the found contract events
	sub  ethereum.Subscription // Subscription for errors, completion and termination
	done bool                  // Whether the subscription completed delivering logs
	fail error                 // Occurred error to stop iteration
}

// Next advances the iterator to the subsequent event, returning whether there
// are any more events found. In case of a retrieval or parsing error, false is
// returned and Error() can be queried for the exact failure.
func (it *ProxeusFSContractUpdatedEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(ProxeusFSContractUpdatedEvent)
			if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
				it.fail = err
				return false
			}
			it.Event.Raw = log
			return true

		default:
			return false
		}
	}
	// Iterator still in progress, wait for either a data or an error event
	select {
	case log := <-it.logs:
		it.Event = new(ProxeusFSContractUpdatedEvent)
		if err := it.contract.UnpackLog(it.Event, it.event, log); err != nil {
			it.fail = err
			return false
		}
		it.Event.Raw = log
		return true

	case err := <-it.sub.Err():
		it.done = true
		it.fail = err
		return it.Next()
	}
}

// Error returns any retrieval or parsing error occurred during filtering.
func (it *ProxeusFSContractUpdatedEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *ProxeusFSContractUpdatedEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// ProxeusFSContractUpdatedEvent represents a UpdatedEvent event raised by the ProxeusFSContract contract.
type ProxeusFSContractUpdatedEvent struct {
	OldHash [32]byte
	NewHash [32]byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterUpdatedEvent is a free log retrieval operation binding the contract event 0xd97d4fc6bd80b061984d0af2aaa1813b97eebabc71ed9307414249dcafb99167.
//
// Solidity: e UpdatedEvent(oldHash indexed bytes32, newHash indexed bytes32)
func (_ProxeusFSContract *ProxeusFSContractFilterer) FilterUpdatedEvent(opts *bind.FilterOpts, oldHash [][32]byte, newHash [][32]byte) (*ProxeusFSContractUpdatedEventIterator, error) {

	var oldHashRule []interface{}
	for _, oldHashItem := range oldHash {
		oldHashRule = append(oldHashRule, oldHashItem)
	}
	var newHashRule []interface{}
	for _, newHashItem := range newHash {
		newHashRule = append(newHashRule, newHashItem)
	}

	logs, sub, err := _ProxeusFSContract.contract.FilterLogs(opts, "UpdatedEvent", oldHashRule, newHashRule)
	if err != nil {
		return nil, err
	}
	return &ProxeusFSContractUpdatedEventIterator{contract: _ProxeusFSContract.contract, event: "UpdatedEvent", logs: logs, sub: sub}, nil
}

// WatchUpdatedEvent is a free log subscription operation binding the contract event 0xd97d4fc6bd80b061984d0af2aaa1813b97eebabc71ed9307414249dcafb99167.
//
// Solidity: e UpdatedEvent(oldHash indexed bytes32, newHash indexed bytes32)
func (_ProxeusFSContract *ProxeusFSContractFilterer) WatchUpdatedEvent(opts *bind.WatchOpts, sink chan<- *ProxeusFSContractUpdatedEvent, oldHash [][32]byte, newHash [][32]byte) (event.Subscription, error) {

	var oldHashRule []interface{}
	for _, oldHashItem := range oldHash {
		oldHashRule = append(oldHashRule, oldHashItem)
	}
	var newHashRule []interface{}
	for _, newHashItem := range newHash {
		newHashRule = append(newHashRule, newHashItem)
	}

	logs, sub, err := _ProxeusFSContract.contract.WatchLogs(opts, "UpdatedEvent", oldHashRule, newHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(ProxeusFSContractUpdatedEvent)
				if err := _ProxeusFSContract.contract.UnpackLog(event, "UpdatedEvent", log); err != nil {
					return err
				}
				event.Raw = log

				select {
				case sink <- event:
				case err := <-sub.Err():
					return err
				case <-quit:
					return nil
				}
			case err := <-sub.Err():
				return err
			case <-quit:
				return nil
			}
		}
	}), nil
}
