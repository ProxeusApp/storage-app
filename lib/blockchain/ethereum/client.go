package ethereum

import (
	"context"
	"fmt"
	"math/big"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/accounts/abi/bind/backends"
	"github.com/ethereum/go-ethereum/core"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	cache "github.com/patrickmn/go-cache"

	"git.proxeus.com/core/central/lib/blockchain/bctypes"
)

type ETHClient struct {
	nodeURL           string
	nodeWSURL         string
	connection        *ethclient.Client
	wsconnconnection  *ethclient.Client
	eventSubscription ethereum.Subscription

	nonceManager NonceManager

	pendingTxCache *cache.Cache
	eventsCache    *cache.Cache
	gasCache       *cache.Cache
	contractCache  *cache.Cache
}

const gasCacheExpiration = 60 * time.Minute
const gasCacheCleanup = 10 * time.Minute

const DefaultContextTimeout = time.Duration(20 * time.Second)

//var instance *ETHClient
//var initialized uint32
//var once sync.Once

//Singleton Pattern
//func GetETHClientInstance() *ETHClient {
//	once.Do(func() {
//		instance = new(ETHClient)
//	})
//	return instance
//}

func NewETHClient(nodeURL string, nodeWSURL string) *ETHClient {
	instance := new(ETHClient)
	instance.Connect(nodeURL, nodeWSURL)
	return instance
}

func (eth *ETHClient) Connect(nodeURL string, nodeWSURL string) (*ETHClient, error) {
	var err error

	eth.pendingTxCache = cache.New(cache.NoExpiration, cache.NoExpiration)
	eth.eventsCache = cache.New(cache.NoExpiration, cache.NoExpiration)
	eth.gasCache = cache.New(gasCacheExpiration, gasCacheCleanup)

	eth.nodeURL = nodeURL
	eth.nodeWSURL = nodeWSURL

	c, err := Dial(eth.nodeURL)
	if err != nil {
		return nil, err
	}
	eth.connection = c
	eth.nonceManager.OnDial(eth.connection)

	return eth, nil
}

func (eth *ETHClient) Bind(contract bctypes.Contract, address common.Address) error {
	var conBack bind.ContractBackend
	conBack = eth.connection

	err := contract.BindFromAddress(address, conBack)
	return err

}

func (eth *ETHClient) DeploySim(contract bctypes.Contract, account bctypes.Account) (*bctypes.Transaction, error) {
	return eth.deploy(contract, account, true)
}
func (eth *ETHClient) Deploy(contract bctypes.Contract, account bctypes.Account) (*bctypes.Transaction, error) {
	return eth.deploy(contract, account, false)
}

func (eth *ETHClient) deploy(contract bctypes.Contract, account bctypes.Account, sim bool) (*bctypes.Transaction, error) {
	auth := account.GetTransactor()
	eth.nonceManager.OnAccountChange(account.GetAddress())
	auth.Nonce = eth.nonceManager.NextNonce()

	var conBack bind.ContractBackend
	var depBack bind.DeployBackend
	conBack = eth.connection
	depBack = eth.connection
	if sim == true {
		alloc := make(core.GenesisAlloc)
		alloc[auth.From] = core.GenesisAccount{Balance: big.NewInt(133700000)}
		simBak := backends.NewSimulatedBackend(alloc, 8000000)
		conBack = simBak
		depBack = simBak
	}

	tx, err := contract.ExecuteDeployment(auth, conBack)
	if err != nil {
		return nil, err
	}

	go func() {
		receipt, err := bind.WaitMined(context.Background(), depBack, tx)
		if err != nil {
			fmt.Println("Error getting tx receipt: ")
			fmt.Println(err)
		} else {
			contract.DeployTransaction().Mined = true
			contract.SetDeployReceipt(receipt)
		}
	}()
	return contract.DeployTransaction(), nil
}

func (eth *ETHClient) Call(contract bctypes.Contract, account bctypes.Account, transaction *bctypes.Transaction) error {
	auth := account.GetTransactor()
	eth.nonceManager.OnAccountChange(account.GetAddress())
	auth.Nonce = eth.nonceManager.NextNonce()

	con := eth.connection

	tx, err := contract.Transact(auth, transaction)
	if err != nil {
		return err
	}
	go func() {
		receipt, err := bind.WaitMined(context.Background(), con, tx)
		if err != nil {
			fmt.Println("Error getting tx receipt: ")
			fmt.Println(err)
		} else {
			transaction.Mined = true
			trr := bctypes.TransactionResult{receipt}
			transaction.Result = &trr
		}
	}()
	transaction.Hash = tx.Hash()
	transaction.Ethtx = tx
	return nil
}

func (eth *ETHClient) SetupListenerForContract(contract bctypes.Contract, fromBlock uint64, channel chan<- bctypes.Event, stop chan bool) error {
	return eth.SetupListener(contract.Address(), fromBlock, channel, stop)
}
func (eth *ETHClient) SetupListener(address common.Address, fromBlock uint64, channel chan<- bctypes.Event, stop chan bool) error {
	contractAddress := address

	ctx := context.Background()
	var err error
	eth.wsconnconnection, err = DialContext(ctx, eth.nodeWSURL)
	if err != nil {
		return err
	}
	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}
	if fromBlock > 0 {
		query.FromBlock = big.NewInt(int64(fromBlock))
	}

	ctx, cancelFunc := context.WithTimeout(ctx, DefaultContextTimeout)

	logChan := make(chan types.Log)
	//Passing Logs/Events to outer Channel and closing when done
	go func() {
		for n := range logChan {
			eth.forwardOnce(&channel, *bctypes.FromLog(n))
		}
		//close(channel)
		//close(stop)
	}()
	//Closing Subscription & Log Channel when outer Channel closed
	go func() {
		stopped, closed := <-stop
		if stopped || closed {
			eth.eventSubscription.Unsubscribe()
			close(logChan)
		}
	}()
	eth.eventSubscription, err = eth.wsconnconnection.SubscribeFilterLogs(ctx, query, logChan)
	if err != nil {
		close(logChan)
		return err
	}
	cancelFunc()

	//Querying past Events from Block
	eth.FetchEvents(address, fromBlock, channel)

	return nil
}
func (eth *ETHClient) FetchEventsForContract(contract bctypes.Contract, fromBlock uint64, channel chan<- bctypes.Event) error {
	return eth.FetchEvents(contract.Address(), fromBlock, channel)
}
func (eth *ETHClient) FetchEvents(address common.Address, fromBlock uint64, channel chan<- bctypes.Event) error {
	contractAddress := address

	query := ethereum.FilterQuery{
		Addresses: []common.Address{contractAddress},
	}
	if fromBlock > 0 {
		query.FromBlock = big.NewInt(int64(fromBlock))
	}
	ctx := context.Background()
	ctx, cancelFunc := context.WithTimeout(ctx, DefaultContextTimeout)
	//Querying past Events from Block
	var logs []types.Log
	fromblockquery := ethereum.FilterQuery{
		FromBlock: big.NewInt(int64(fromBlock)),
		Addresses: []common.Address{contractAddress},
	}
	logs, err := eth.connection.FilterLogs(ctx, fromblockquery)
	if err != nil {
		return err
	}
	go func() {
		for _, n := range logs {
			eth.forwardOnce(&channel, *bctypes.FromLog(n))
		}
	}()
	cancelFunc()
	return nil
}

func (eth *ETHClient) forwardOnce(channel *chan<- bctypes.Event, event bctypes.Event) {
	_, exists := eth.eventsCache.Get(event.UniqueID)
	if !exists {
		eth.eventsCache.Add(event.UniqueID, event, cache.NoExpiration)
		*channel <- event
	}
}
