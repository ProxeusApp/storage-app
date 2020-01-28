package contract

import (
	"encoding/hex"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ProxeusApp/storage-app/lib/blockchain/bctypes"
	"github.com/ProxeusApp/storage-app/lib/blockchain/ethereum"
)

type TestcontractContract struct {
	bctypes.Contract
	address           common.Address
	binding           interface{}
	receipt           *types.Receipt
	deployTransaction *bctypes.Transaction

	Ethtx             *types.Transaction
	DeploymentReceipt *types.Receipt
}

//Functions

func (con *TestcontractContract) Testfunction(ethclient *ethereum.ETHClient, account bctypes.Account, parameter1 string) (*bctypes.Transaction, error) {
	transaction, err := con.createTransaction("Testfunction")
	if err != nil {
		return nil, err
	}

	var parameters []interface{}
	hexstring := hex.EncodeToString([]byte(parameter1))
	ethparameter1 := common.HexToHash(string(hexstring))
	parameters = append(parameters, ethparameter1)

	transaction.Params = parameters

	err = ethclient.Call(*con.asGeneric(), account, transaction)
	return transaction, err
}

//Glue Part (dynamic: Contract name)
func (con *TestcontractContract) BindFromAddress(address common.Address, backend bind.ContractBackend) error {
	binding, err := NewTestcontract(address, backend)
	con.SetBinding(binding)
	con.SetAddress(address)
	return err
}

func (con *TestcontractContract) ExecuteDeployment(auth *bind.TransactOpts, backend bind.ContractBackend) (*types.Transaction, error) {
	address, tx, binding, err := DeployTestcontract(auth, backend)
	if err != nil {
		return nil, err
	}
	con.SetBinding(binding)
	con.SetAddress(address)

	transaction := bctypes.Transaction{Hash: tx.Hash(), Ethtx: tx}
	con.SetDeployTransaction(&transaction)
	return tx, err
}

func (con *TestcontractContract) Transact(opts *bind.TransactOpts, transaction *bctypes.Transaction) (*types.Transaction, error) {
	binding := con.binding.(*Testcontract)
	rawTransactor := TestcontractRaw{binding}

	tx, err := rawTransactor.Transact(opts, transaction.Function, transaction.Params...)
	return tx, err
}

func (con *TestcontractContract) UnpackTestEvent(event bctypes.Event) string {
	ethevent := new(TestcontractTestEvent)
	err := con.binding.(*Testcontract).TestcontractFilterer.contract.UnpackLog(ethevent, "TestEvent", event.Log)
	if err != nil {
		fmt.Println(err)
		return "Invalid"
	}
	messageString := string(ethevent.Message[:])
	return messageString
}

// Generic Part
func (con *TestcontractContract) Address() common.Address {
	return con.address
}
func (con *TestcontractContract) SetAddress(address common.Address) {
	con.address = address
}
func (con *TestcontractContract) SetBinding(binding interface{}) {
	con.binding = binding
}
func (con *TestcontractContract) DeployReceipt() *types.Receipt {
	return con.receipt
}
func (con *TestcontractContract) SetDeployReceipt(receipt *types.Receipt) {
	con.receipt = receipt
}
func (con *TestcontractContract) DeployTransaction() *bctypes.Transaction {
	return con.deployTransaction
}
func (con *TestcontractContract) SetDeployTransaction(deployTransaction *bctypes.Transaction) {
	con.deployTransaction = deployTransaction
}

func (con *TestcontractContract) Deploy(ethclient *ethereum.ETHClient, account bctypes.Account) (*bctypes.Transaction, error) {
	var contract bctypes.Contract
	contract = con
	transaction, err := ethclient.Deploy(contract, account)
	return transaction, err
}

func (con *TestcontractContract) BindContract(address string, client *ethereum.ETHClient) error {
	return client.Bind(*con.asGeneric(), common.HexToAddress(address))
}

func (con *TestcontractContract) FetchEvents(ethclient *ethereum.ETHClient, fromBlock uint64, channel chan<- bctypes.Event) error {
	err := ethclient.FetchEventsForContract(*con.asGeneric(), fromBlock, channel)
	return err
}

func (con *TestcontractContract) SetupEventListener(ethclient *ethereum.ETHClient, fromBlock uint64, channel chan<- bctypes.Event, stop chan bool) error {
	err := ethclient.SetupListenerForContract(*con.asGeneric(), fromBlock, channel, stop)
	return err
}

func (con *TestcontractContract) createTransaction(function string) (*bctypes.Transaction, error) {
	transaction, err := bctypes.NewTransaction(*con.asGeneric(), function)
	if err != nil {
		return nil, err
	}
	return transaction, err
}

func (con *TestcontractContract) asGeneric() *bctypes.Contract {
	var contract bctypes.Contract
	contract = con
	return &contract
}
