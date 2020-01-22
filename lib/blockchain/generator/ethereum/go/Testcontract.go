// Code generated - DO NOT EDIT.
// This file is a generated binding and any manual changes will be lost.

package contract

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

// Reference imports to suppress errors if they are not otherwise used.
var (
	_ = big.NewInt
	_ = strings.NewReader
	_ = ethereum.NotFound
	_ = abi.U256
	_ = bind.Bind
	_ = common.Big1
	_ = types.BloomLookup
	_ = event.NewSubscription
)

// TestcontractABI is the input ABI used to generate the binding from.
const TestcontractABI = "[{\"constant\":false,\"inputs\":[{\"name\":\"input\",\"type\":\"bytes32\"}],\"name\":\"Testfunction\",\"outputs\":[],\"payable\":false,\"stateMutability\":\"nonpayable\",\"type\":\"function\"},{\"anonymous\":false,\"inputs\":[{\"indexed\":false,\"name\":\"message\",\"type\":\"bytes32\"}],\"name\":\"TestEvent\",\"type\":\"event\"}]"

// TestcontractBin is the compiled bytecode used for deploying new contracts.
const TestcontractBin = `0x6080604052348015600f57600080fd5b5060bc8061001e6000396000f300608060405260043610603e5763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166305183fea81146043575b600080fd5b348015604e57600080fd5b506058600435605a565b005b6040805182815290517f2d87364d1542bf89b684ede9ddff45aed45971c6f05deaca687bd3d1b6caf1c39181900360200190a1505600a165627a7a72305820182128bab2170ac22138590783e964cbb4da47804776efafcc36854feaeac6060029`

// DeployTestcontract deploys a new Ethereum contract, binding an instance of Testcontract to it.
func DeployTestcontract(auth *bind.TransactOpts, backend bind.ContractBackend) (common.Address, *types.Transaction, *Testcontract, error) {
	parsed, err := abi.JSON(strings.NewReader(TestcontractABI))
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	address, tx, contract, err := bind.DeployContract(auth, parsed, common.FromHex(TestcontractBin), backend)
	if err != nil {
		return common.Address{}, nil, nil, err
	}
	return address, tx, &Testcontract{TestcontractCaller: TestcontractCaller{contract: contract}, TestcontractTransactor: TestcontractTransactor{contract: contract}, TestcontractFilterer: TestcontractFilterer{contract: contract}}, nil
}

// Testcontract is an auto generated Go binding around an Ethereum contract.
type Testcontract struct {
	TestcontractCaller     // Read-only binding to the contract
	TestcontractTransactor // Write-only binding to the contract
	TestcontractFilterer   // Log filterer for contract events
}

// TestcontractCaller is an auto generated read-only Go binding around an Ethereum contract.
type TestcontractCaller struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestcontractTransactor is an auto generated write-only Go binding around an Ethereum contract.
type TestcontractTransactor struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestcontractFilterer is an auto generated log filtering Go binding around an Ethereum contract events.
type TestcontractFilterer struct {
	contract *bind.BoundContract // Generic contract wrapper for the low level calls
}

// TestcontractSession is an auto generated Go binding around an Ethereum contract,
// with pre-set call and transact options.
type TestcontractSession struct {
	Contract     *Testcontract     // Generic contract binding to set the session for
	CallOpts     bind.CallOpts     // Call options to use throughout this session
	TransactOpts bind.TransactOpts // Transaction auth options to use throughout this session
}

// TestcontractCallerSession is an auto generated read-only Go binding around an Ethereum contract,
// with pre-set call options.
type TestcontractCallerSession struct {
	Contract *TestcontractCaller // Generic contract caller binding to set the session for
	CallOpts bind.CallOpts       // Call options to use throughout this session
}

// TestcontractTransactorSession is an auto generated write-only Go binding around an Ethereum contract,
// with pre-set transact options.
type TestcontractTransactorSession struct {
	Contract     *TestcontractTransactor // Generic contract transactor binding to set the session for
	TransactOpts bind.TransactOpts       // Transaction auth options to use throughout this session
}

// TestcontractRaw is an auto generated low-level Go binding around an Ethereum contract.
type TestcontractRaw struct {
	Contract *Testcontract // Generic contract binding to access the raw methods on
}

// TestcontractCallerRaw is an auto generated low-level read-only Go binding around an Ethereum contract.
type TestcontractCallerRaw struct {
	Contract *TestcontractCaller // Generic read-only contract binding to access the raw methods on
}

// TestcontractTransactorRaw is an auto generated low-level write-only Go binding around an Ethereum contract.
type TestcontractTransactorRaw struct {
	Contract *TestcontractTransactor // Generic write-only contract binding to access the raw methods on
}

// NewTestcontract creates a new instance of Testcontract, bound to a specific deployed contract.
func NewTestcontract(address common.Address, backend bind.ContractBackend) (*Testcontract, error) {
	contract, err := bindTestcontract(address, backend, backend, backend)
	if err != nil {
		return nil, err
	}
	return &Testcontract{TestcontractCaller: TestcontractCaller{contract: contract}, TestcontractTransactor: TestcontractTransactor{contract: contract}, TestcontractFilterer: TestcontractFilterer{contract: contract}}, nil
}

// NewTestcontractCaller creates a new read-only instance of Testcontract, bound to a specific deployed contract.
func NewTestcontractCaller(address common.Address, caller bind.ContractCaller) (*TestcontractCaller, error) {
	contract, err := bindTestcontract(address, caller, nil, nil)
	if err != nil {
		return nil, err
	}
	return &TestcontractCaller{contract: contract}, nil
}

// NewTestcontractTransactor creates a new write-only instance of Testcontract, bound to a specific deployed contract.
func NewTestcontractTransactor(address common.Address, transactor bind.ContractTransactor) (*TestcontractTransactor, error) {
	contract, err := bindTestcontract(address, nil, transactor, nil)
	if err != nil {
		return nil, err
	}
	return &TestcontractTransactor{contract: contract}, nil
}

// NewTestcontractFilterer creates a new log filterer instance of Testcontract, bound to a specific deployed contract.
func NewTestcontractFilterer(address common.Address, filterer bind.ContractFilterer) (*TestcontractFilterer, error) {
	contract, err := bindTestcontract(address, nil, nil, filterer)
	if err != nil {
		return nil, err
	}
	return &TestcontractFilterer{contract: contract}, nil
}

// bindTestcontract binds a generic wrapper to an already deployed contract.
func bindTestcontract(address common.Address, caller bind.ContractCaller, transactor bind.ContractTransactor, filterer bind.ContractFilterer) (*bind.BoundContract, error) {
	parsed, err := abi.JSON(strings.NewReader(TestcontractABI))
	if err != nil {
		return nil, err
	}
	return bind.NewBoundContract(address, parsed, caller, transactor, filterer), nil
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Testcontract *TestcontractRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Testcontract.Contract.TestcontractCaller.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Testcontract *TestcontractRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Testcontract.Contract.TestcontractTransactor.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Testcontract *TestcontractRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Testcontract.Contract.TestcontractTransactor.contract.Transact(opts, method, params...)
}

// Call invokes the (constant) contract method with params as input values and
// sets the output to result. The result type might be a single field for simple
// returns, a slice of interfaces for anonymous returns and a struct for named
// returns.
func (_Testcontract *TestcontractCallerRaw) Call(opts *bind.CallOpts, result interface{}, method string, params ...interface{}) error {
	return _Testcontract.Contract.contract.Call(opts, result, method, params...)
}

// Transfer initiates a plain transaction to move funds to the contract, calling
// its default method if one is available.
func (_Testcontract *TestcontractTransactorRaw) Transfer(opts *bind.TransactOpts) (*types.Transaction, error) {
	return _Testcontract.Contract.contract.Transfer(opts)
}

// Transact invokes the (paid) contract method with params as input values.
func (_Testcontract *TestcontractTransactorRaw) Transact(opts *bind.TransactOpts, method string, params ...interface{}) (*types.Transaction, error) {
	return _Testcontract.Contract.contract.Transact(opts, method, params...)
}

// Testfunction is a paid mutator transaction binding the contract method 0x05183fea.
//
// Solidity: function Testfunction(input bytes32) returns()
func (_Testcontract *TestcontractTransactor) Testfunction(opts *bind.TransactOpts, input [32]byte) (*types.Transaction, error) {
	return _Testcontract.contract.Transact(opts, "Testfunction", input)
}

// Testfunction is a paid mutator transaction binding the contract method 0x05183fea.
//
// Solidity: function Testfunction(input bytes32) returns()
func (_Testcontract *TestcontractSession) Testfunction(input [32]byte) (*types.Transaction, error) {
	return _Testcontract.Contract.Testfunction(&_Testcontract.TransactOpts, input)
}

// Testfunction is a paid mutator transaction binding the contract method 0x05183fea.
//
// Solidity: function Testfunction(input bytes32) returns()
func (_Testcontract *TestcontractTransactorSession) Testfunction(input [32]byte) (*types.Transaction, error) {
	return _Testcontract.Contract.Testfunction(&_Testcontract.TransactOpts, input)
}

// TestcontractTestEventIterator is returned from FilterTestEvent and is used to iterate over the raw logs and unpacked data for TestEvent events raised by the Testcontract contract.
type TestcontractTestEventIterator struct {
	Event *TestcontractTestEvent // Event containing the contract specifics and raw log

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
func (it *TestcontractTestEventIterator) Next() bool {
	// If the iterator failed, stop iterating
	if it.fail != nil {
		return false
	}
	// If the iterator completed, deliver directly whatever's available
	if it.done {
		select {
		case log := <-it.logs:
			it.Event = new(TestcontractTestEvent)
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
		it.Event = new(TestcontractTestEvent)
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
func (it *TestcontractTestEventIterator) Error() error {
	return it.fail
}

// Close terminates the iteration process, releasing any pending underlying
// resources.
func (it *TestcontractTestEventIterator) Close() error {
	it.sub.Unsubscribe()
	return nil
}

// TestcontractTestEvent represents a TestEvent event raised by the Testcontract contract.
type TestcontractTestEvent struct {
	Message [32]byte
	Raw     types.Log // Blockchain specific contextual infos
}

// FilterTestEvent is a free log retrieval operation binding the contract event 0x2d87364d1542bf89b684ede9ddff45aed45971c6f05deaca687bd3d1b6caf1c3.
//
// Solidity: e TestEvent(message bytes32)
func (_Testcontract *TestcontractFilterer) FilterTestEvent(opts *bind.FilterOpts) (*TestcontractTestEventIterator, error) {

	logs, sub, err := _Testcontract.contract.FilterLogs(opts, "TestEvent")
	if err != nil {
		return nil, err
	}
	return &TestcontractTestEventIterator{contract: _Testcontract.contract, event: "TestEvent", logs: logs, sub: sub}, nil
}

// WatchTestEvent is a free log subscription operation binding the contract event 0x2d87364d1542bf89b684ede9ddff45aed45971c6f05deaca687bd3d1b6caf1c3.
//
// Solidity: e TestEvent(message bytes32)
func (_Testcontract *TestcontractFilterer) WatchTestEvent(opts *bind.WatchOpts, sink chan<- *TestcontractTestEvent) (event.Subscription, error) {

	logs, sub, err := _Testcontract.contract.WatchLogs(opts, "TestEvent")
	if err != nil {
		return nil, err
	}
	return event.NewSubscription(func(quit <-chan struct{}) error {
		defer sub.Unsubscribe()
		for {
			select {
			case log := <-logs:
				// New log arrived, parse the event and forward to the user
				event := new(TestcontractTestEvent)
				if err := _Testcontract.contract.UnpackLog(event, "TestEvent", log); err != nil {
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
