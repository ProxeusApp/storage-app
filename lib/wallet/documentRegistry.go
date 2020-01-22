// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package wallet

import (
	"strings"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/event"
)

// DocumentRegistryABI is the input ABI used to generate the binding from.
const DocumentRegistryABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"docHash\",\"type\":\"bytes32\"}],\"name\":\"notarizeDocument\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[],\"name\":\"getNotarizedDocuments\",\"outputs\":[{\"name\":\"\",\"type\":\"bytes32[]\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"constant\":true,\"inputs\":[{\"name\":\"docHash\",\"type\":\"bytes32\"}],\"name\":\"isNotarized\",\"outputs\":[{\"name\":\"\",\"type\":\"bool\"}],\"payable\":false,\"stateMutability\":\"view\",\"type\":\"function\"},{\"inputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"constructor\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":true,\"name\":\"from\",\"type\":\"address\"},{\"indexed\":true,\"name\":\"docHash\",\"type\":\"bytes32\"}],\"name\":\"ENotarizeDocument\",\"type\":\"event\"}]"

// DocumentRegistry is an auto generated Go binding around an Ethereum contract.
type DocumentRegistry struct {
	DocumentRegistryCaller     // Read-only binding to the contract
	DocumentRegistryTransactor // Write-only binding to the contract
	DocumentRegistryFilterer   // Log filterer for contract events
}

// DocumentRegistryCaller is an auto generated read-only Go binding around an Ethereum contract.
type DocumentRegistryCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DocumentRegistryTransactor is an auto generated write-only Go binding around an Ethereum contract.
type DocumentRegistryTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DocumentRegistryFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type DocumentRegistryFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// DocumentRegistrySession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type DocumentRegistrySession struct {
	Contract     *DocumentRegistry // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// DocumentRegistryCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type DocumentRegistryCallerSession struct {
	Contract *DocumentRegistryCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts           // Call options to use throughout this session
}

// DocumentRegistryTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type DocumentRegistryTransactorSession struct {
	Contract     *DocumentRegistryTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts           // Transaction auth options to use throughout this session
}

// DocumentRegistryRaw is an auto generated low-level Go binding around an Ethereum contract.
type DocumentRegistryRaw struct {
	Contract *DocumentRegistry // Generic contract binding to access the raw methods on
}

// DocumentRegistryCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type DocumentRegistryCallerRaw struct {
	Contract *DocumentRegistryCaller // Generic read-only contract binding to access the raw methods on
}

// DocumentRegistryTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type DocumentRegistryTransactorRaw struct {
	Contract *DocumentRegistryTransactor // Generic write-only contract binding to access the raw methods on
}

// NewDocumentRegistry creates a new instance of DocumentRegistry, bound to a specific deployed contract.
func NewDocumentRegistry(address common.Address, backend bind.ContractBackend) (*DocumentRegistry, error) {
	contract, err := bindDocumentRegistry(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &DocumentRegistry{DocumentRegistryCaller: DocumentRegistryCaller{contract: contract}, DocumentRegistryTransactor: DocumentRegistryTransactor{contract: contract}, DocumentRegistryFilterer: DocumentRegistryFilterer{contract: contract}}, nil
}

// NewDocumentRegistryCaller creates a new read-only instance of DocumentRegistry, bound to a specific deployed contract.
func NewDocumentRegistryCaller(address common.Address, caller bind.ContractCaller) (*DocumentRegistryCaller, error) {
	contract, err := bindDocumentRegistry(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &DocumentRegistryCaller{contract: contract}, nil
}

// NewDocumentRegistryTransactor creates a new write-only instance of DocumentRegistry, bound to a specific deployed contract.
func NewDocumentRegistryTransactor(address common.Address, transactor bind.ContractTransactor) (*DocumentRegistryTransactor, error) {
	contract, err := bindDocumentRegistry(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &DocumentRegistryTransactor{contract: contract}, nil
}

// NewDocumentRegistryFilterer creates a new log filterer instance of DocumentRegistry, bound to a specific deployed contract.
func NewDocumentRegistryFilterer(address common.Address, filterer bind.ContractFilterer) (*DocumentRegistryFilterer, error) {
	contract, err := bindDocumentRegistry(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &DocumentRegistryFilterer{contract: contract}, nil
}

// bindDocumentRegistry binds a generic wrapper to an already deployed contract.
func bindDocumentRegistry(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(DocumentRegistryABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DocumentRegistry *DocumentRegistryRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _DocumentRegistry.Contract.DocumentRegistryCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DocumentRegistry *DocumentRegistryRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DocumentRegistry.Contract.DocumentRegistryTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DocumentRegistry *DocumentRegistryRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DocumentRegistry.Contract.DocumentRegistryTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_DocumentRegistry *DocumentRegistryCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _DocumentRegistry.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_DocumentRegistry *DocumentRegistryTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _DocumentRegistry.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_DocumentRegistry *DocumentRegistryTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _DocumentRegistry.Contract.contract.Transact(opts, method, params...)
}

// GetNotarizedDocuments is a free data retrieval call binding the contract method 0x9637f2d6.
//
// Solidity: function getNotarizedDocuments() constant returns(bytes32[])
func (_DocumentRegistry *DocumentRegistryCaller) GetNotarizedDocuments(opts *bind.CallOpts) ([][32]byte, error) {
	var (
		ret0 = new([][32]byte)
	)
	out := ret0
	err := _DocumentRegistry.contract.Call(opts, out, "getNotarizedDocuments")
	return *ret0, err
}

// GetNotarizedDocuments is a free data retrieval call binding the contract method 0x9637f2d6.
//
// Solidity: function getNotarizedDocuments() constant returns(bytes32[])
func (_DocumentRegistry *DocumentRegistrySession) GetNotarizedDocuments() ([][32]byte, error) {
	return _DocumentRegistry.Contract.GetNotarizedDocuments(&_DocumentRegistry.CallOpts)
}

// GetNotarizedDocuments is a free data retrieval call binding the contract method 0x9637f2d6.
//
// Solidity: function getNotarizedDocuments() constant returns(bytes32[])
func (_DocumentRegistry *DocumentRegistryCallerSession) GetNotarizedDocuments() ([][32]byte, error) {
	return _DocumentRegistry.Contract.GetNotarizedDocuments(&_DocumentRegistry.CallOpts)
}

// IsNotarized is a free data retrieval call binding the contract method 0xfe6ad6c6.
//
// Solidity: function isNotarized(docHash bytes32) constant returns(bool)
func (_DocumentRegistry *DocumentRegistryCaller) IsNotarized(opts *bind.CallOpts, docHash [32]byte) (bool, error) {
	var (
		ret0 = new(bool)
	)
	out := ret0
	err := _DocumentRegistry.contract.Call(opts, out, "isNotarized", docHash)
	return *ret0, err
}

// IsNotarized is a free data retrieval call binding the contract method 0xfe6ad6c6.
//
// Solidity: function isNotarized(docHash bytes32) constant returns(bool)
func (_DocumentRegistry *DocumentRegistrySession) IsNotarized(docHash [32]byte) (bool, error) {
	return _DocumentRegistry.Contract.IsNotarized(&_DocumentRegistry.CallOpts, docHash)
}

// IsNotarized is a free data retrieval call binding the contract method 0xfe6ad6c6.
//
// Solidity: function isNotarized(docHash bytes32) constant returns(bool)
func (_DocumentRegistry *DocumentRegistryCallerSession) IsNotarized(docHash [32]byte) (bool, error) {
	return _DocumentRegistry.Contract.IsNotarized(&_DocumentRegistry.CallOpts, docHash)
}

// NotarizeDocument is a paid mutator transaction binding the contract method 0x4c565d5b.
//
// Solidity: function notarizeDocument(docHash bytes32) returns()
func (_DocumentRegistry *DocumentRegistryTransactor) NotarizeDocument(opts *bind.TransactOpts, docHash [32]byte) (*types.Transaction, error) {
	return _DocumentRegistry.contract.Transact(opts, "notarizeDocument", docHash)
}

// NotarizeDocument is a paid mutator transaction binding the contract method 0x4c565d5b.
//
// Solidity: function notarizeDocument(docHash bytes32) returns()
func (_DocumentRegistry *DocumentRegistrySession) NotarizeDocument(docHash [32]byte) (*types.Transaction, error) {
	return _DocumentRegistry.Contract.NotarizeDocument(&_DocumentRegistry.TransactOpts, docHash)
}

// NotarizeDocument is a paid mutator transaction binding the contract method 0x4c565d5b.
//
// Solidity: function notarizeDocument(docHash bytes32) returns()
func (_DocumentRegistry *DocumentRegistryTransactorSession) NotarizeDocument(docHash [32]byte) (*types.Transaction, error) {
	return _DocumentRegistry.Contract.NotarizeDocument(&_DocumentRegistry.TransactOpts, docHash)
}

// DocumentRegistryENotarizeDocumentIterator is returned from FilterENotarizeDocument and is used to iterate over the raw logs and unpacked data for ENotarizeDocument events raised by the DocumentRegistry contract.
type DocumentRegistryENotarizeDocumentIterator struct {
	Event *DocumentRegistryENotarizeDocument // Event containing the contract specifics and raw log

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
func (it *DocumentRegistryENotarizeDocumentIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(DocumentRegistryENotarizeDocument)
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
		it.Event = new(DocumentRegistryENotarizeDocument)
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
func (it *DocumentRegistryENotarizeDocumentIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *DocumentRegistryENotarizeDocumentIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// DocumentRegistryENotarizeDocument represents a ENotarizeDocument event raised by the DocumentRegistry contract.
type DocumentRegistryENotarizeDocument struct {
	From    common.Address
	DocHash [32]byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterENotarizeDocument is a free log retrieval operation binding the contract event 0x5d4f80fffb9c702b33662ae7ed9446384567c81e045ba99652ece723c2dec6dc.
//
// Solidity: e ENotarizeDocument(from indexed address, docHash indexed bytes32)
func (_DocumentRegistry *DocumentRegistryFilterer) FilterENotarizeDocument(opts *bind.FilterOpts, from []common.Address, docHash [][32]byte) (*DocumentRegistryENotarizeDocumentIterator, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var docHashRule []interface{}
	for _, docHashItem := range docHash {
		docHashRule = append(docHashRule, docHashItem)
	}

	logs, sub, err := _DocumentRegistry.contract.FilterLogs(opts, "ENotarizeDocument", fromRule, docHashRule)
	if err != nil {
		return nil, err
	}
	return &DocumentRegistryENotarizeDocumentIterator{contract: _DocumentRegistry.contract, event: "ENotarizeDocument", logs: logs, sub: sub}, nil
}

// WatchENotarizeDocument is a free log subscription operation binding the contract event 0x5d4f80fffb9c702b33662ae7ed9446384567c81e045ba99652ece723c2dec6dc.
//
// Solidity: e ENotarizeDocument(from indexed address, docHash indexed bytes32)
func (_DocumentRegistry *DocumentRegistryFilterer) WatchENotarizeDocument(opts *bind.WatchOpts, sink chan<- *DocumentRegistryENotarizeDocument, from []common.Address, docHash [][32]byte) (event.Subscription, error) {

	var fromRule []interface{}
	for _, fromItem := range from {
		fromRule = append(fromRule, fromItem)
	}
	var docHashRule []interface{}
	for _, docHashItem := range docHash {
		docHashRule = append(docHashRule, docHashItem)
	}

	logs, sub, err := _DocumentRegistry.contract.WatchLogs(opts, "ENotarizeDocument", fromRule, docHashRule)
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(DocumentRegistryENotarizeDocument)
				if err := _DocumentRegistry.contract.UnpackLog(event, "ENotarizeDocument", log); err != nil {
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
