package ethereum

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ProxeusApp/storage-app/dapp/core/ethglue"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ProxeusApp/storage-app/dapp/core/embdb"

	cache "github.com/ProxeusApp/memcache"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

type (
	baseClient struct {
		//to prevent from duplication with polling and pushing
		ensureUniquenessOfEventCache *cache.Cache
		gasCache                     *cache.Cache
		listenerLock                 *sync.RWMutex
		currentAddress               string
		txListener                   func(tx *PendingTx, txHash, status string) error
		pendingTxPoller              func()

		filterAddresses []common.Address

		ethClientURL    string
		ethWebSocketURL string
		ethwsconn       *ethclient.Client
		ethconn         *ethclient.Client
		connStatus      EthConnectionStatus
		nonceManager    ethglue.NonceManager

		sub  ethereum.Subscription
		logs chan types.Log

		eventsDB    *embdb.DB
		pendingTxDB *embdb.DB

		pendingTxListener         func(txHash string)
		eventsHandler             func(lg *types.Log, recent bool) error
		updateEthInterfaceHandler func() error

		workersRunning bool
		stopping       bool
		stopChan       chan bool
		stopWg         sync.WaitGroup
		workersLock    sync.Mutex
	}

	gasCacheEntry struct {
		GasPrice *big.Int
		GasLimit uint64
	}
)

const (
	pendingStorageName = "tx"
	eventsStorageName  = "events"
	lastBlockKey       = "last-block"
)

//this will be used only if it fails to estimate
var defaultGas = &gasCacheEntry{
	GasPrice: big.NewInt(10000000000), // 10 Gwei
	GasLimit: 3000000,
}

func NewBaseClient(listenerLock *sync.RWMutex, ethWebSocketURL, ethClientURL string,
	filterAddresses []common.Address) *baseClient {

	me := new(baseClient)
	me.listenerLock = listenerLock
	me.ethWebSocketURL = ethWebSocketURL
	me.ethClientURL = ethClientURL
	me.filterAddresses = filterAddresses
	return me
}

func (me *baseClient) initLocalStorage(storageDir string) (err error) {

	if me.eventsDB, err = embdb.Open(storageDir, eventsStorageName); err != nil {
		return
	}

	if me.pendingTxDB, err = embdb.Open(storageDir, pendingStorageName); err != nil {
		return err
	}

	me.gasCache = cache.New(60 * time.Minute)

	me.ensureUniquenessOfEventCache = cache.NewExtendExpiryOnGet(10*time.Minute, true)
	return nil
}

func (me *baseClient) initListeners(ethAddr string, xtl func(tx *PendingTx, txHash, status string) error,
	pendingTxListener func(txHash string), pendingTxPoller func(), eventsHandler func(lg *types.Log, recent bool) error,
	updateEthInterfaceHandler func() error) {

	me.listenerLock.Lock()
	me.txListener = xtl
	me.currentAddress = strings.ToLower(ethAddr)

	me.listenerLock.Unlock()

	me.ensureUniquenessOfEventCache.Clean()

	me.pendingTxListener = pendingTxListener
	me.pendingTxPoller = pendingTxPoller
	me.eventsHandler = eventsHandler
	me.updateEthInterfaceHandler = updateEthInterfaceHandler
}

func (me *baseClient) startWorkers() {
	me.workersLock.Lock()
	defer me.workersLock.Unlock()
	if me.workersRunning {
		return
	}
	me.stopping = false
	me.workersRunning = true
	me.logs = make(chan types.Log, 200)
	me.stopChan = make(chan bool)
	go me.listen()
	go me.poll()
}

func (me *baseClient) stopWorkers() {
	me.workersLock.Lock()
	defer me.workersLock.Unlock()
	if !me.workersRunning {
		return
	}
	log.Println("[baseClient][doStopWorkers] sending to channel me.stopChan <- true")
	me.workersRunning = false
	me.stopChan <- true
	close(me.stopChan)
	me.stopWg.Wait()
	if me.ethwsconn != nil {
		me.ethwsconn.Close()
	}
}

func (me *baseClient) ethConnectWebSocketsAsync(ctx context.Context) <-chan struct{} {
	readyCh := make(chan struct{})
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				var err error
				me.ethwsconn, err = ethglue.DialContext(ctx, me.ethWebSocketURL)
				if err != nil {
					log.Printf("failed to dial for eth events, will retry (%s)\n", err)
					continue
				}
				query := ethereum.FilterQuery{
					Addresses: me.filterAddresses,
				}
				ctx, cancel := context.WithTimeout(ctx, time.Duration(10*time.Second))
				me.sub, err = me.ethwsconn.SubscribeFilterLogs(ctx, query, me.logs)
				cancel()
				if err != nil {
					log.Printf("failed to subscribe for eth events, will retry (%s)\n", err)
					time.Sleep(time.Second * 4)
					continue
				}
				// success!
				readyCh <- struct{}{}
				return
			}
		}
	}()
	return readyCh
}

func (me *baseClient) listen() {
	me.stopWg.Add(1)
	ctx, cancel := context.WithCancel(context.TODO())
	var readyCh <-chan struct{}

	defer func() {
		me.stopping = true
		cancel()
		me.stopWg.Done()
		log.Println("listen() ended...")
	}()

	for {
		readyCh = me.ethConnectWebSocketsAsync(ctx)
		select {
		case <-readyCh:
			if me.stopping {
				return
			}
			log.Println("listen() started...")
			reconnect := me.listenLoop()
			if !reconnect {
				return
			}
		case <-me.stopChan:
			return
		}
	}
}

func (me *baseClient) listenLoop() (shouldReconnect bool) {
	for {
		select {
		case err, ok := <-me.sub.Err():
			if !ok {
				return true
			}
			log.Println("ERROR sub", err)
			return true
		case vLog, ok := <-me.logs:
			if !ok {
				return true
			}
			if me.stopping {
				return false
			}
			if me.currentAddress == "" {
				continue // not logged in
			}
			me.eventsHandler(&vLog, true)
			if me.pendingTxListener != nil {
				me.pendingTxListener(vLog.TxHash.Hex())
			}
		case <-me.stopChan:
			return false
		}
	}
}

func (me *baseClient) poll() {
	me.stopWg.Add(1)
	defer func() {
		me.stopping = true
		me.stopWg.Done()
		log.Println("poll() ended...")
	}()
	ticker := time.NewTicker(time.Second * 30)
	if me.connStatus == ConnOnline {
		me.connStatus = ConnSyncing
		me.pushConnStatus()
	}
	log.Println("poll() started...")

	tickerFunc := func() {
		if me.stopping {
			return
		}
		if me.currentAddress == "" {
			return // not logged in
		}
		//poll transactions because failed tx are not pushed and to ensure we catch up correctly
		if me.pendingTxPoller != nil {
			me.pendingTxPoller()
		}
		me.pollEvents()
	}

	tickerFunc()
	for {
		select {
		case <-ticker.C:
			tickerFunc()
		case <-me.stopChan:
			return
		}
	}
}

func (me *baseClient) ethReconnect() {
	if me.ethconn != nil {
		me.ethconn.Close()
	}
	me.connStatus = ConnOffline
	me.pushConnStatus()
	c, err := ethglue.Dial(me.ethClientURL)
	if err != nil {
		log.Printf("ethconn dial error (%s) \n", err)
		return
	}
	me.ethconn = c
	me.connStatus = ConnSyncing
	if me.updateEthInterfaceHandler != nil {
		me.updateEthInterfaceHandler()
	}
	me.pushConnStatus()
	me.nonceManager.OnDial(me.ethconn)
}

func (me *baseClient) pollEvents() {
	if me.stopping {
		return
	}

	b, err := me.eventsDB.Get(me.getLastBlockKey())
	var lastBlock uint64
	if err != nil || len(b) < 8 {
		// for mainnet should be more like 6m, but 4m potentially works too
		lastBlock = 6000000
	} else {
		lastBlock = binary.LittleEndian.Uint64(b) + 1
	}

	ctx, cancel := me.ctxWithTimeout()
	header, err := me.ethconn.HeaderByNumber(ctx, nil)
	cancel()
	if err != nil {
		log.Printf("[baseClient][pollEvents] Ethereum connection broken. Can't get header by number (%s). Reconnecting...\n", err)
		me.ethReconnect()
		return
	}
	latestBlockNumber := header.Number

	// Batch blocks
	startBlocks, toBlocks := me.getBlockChunks(big.NewInt(int64(lastBlock)), latestBlockNumber, 2000)

	var lastProcessedBlock uint64
	for i, block := range startBlocks {
		if me.stopping {
			return
		}

		log.Println("[baseClient][pollEvents] Fetching data starting from block", block, "to", toBlocks[i])
		var logs []types.Log
		ctx, cancel := me.ctxWithTimeout()
		logs, err = me.ethconn.FilterLogs(ctx, ethereum.FilterQuery{
			FromBlock: block,
			ToBlock:   toBlocks[i], // batched
			Addresses: me.filterAddresses,
		})
		cancel()
		if err != nil {
			log.Printf("[baseClient][pollEvents] Ethereum connection broken (%s). Reconnecting...\n", err)
			me.ethReconnect()
			return
		}

		for _, lg := range logs {
			// determine recent state and handle events
			if err = me.eventsHandler(&lg, lg.BlockNumber >= latestBlockNumber.Uint64()-3); err != nil {
				log.Println("[baseClient][pollEvents] Can't process Log", err.Error())
				return
			}
			lastProcessedBlock = lg.BlockNumber
		}

		if err = me.putEventBlock(toBlocks[i].Uint64()); err != nil {
			log.Println("[baseclient][pollEvents] can't update last processed block", err)
			return
		}
	}

	// No logs processed or latest block was processed
	if lastProcessedBlock == 0 || lastProcessedBlock == latestBlockNumber.Uint64() {
		if me.connStatus != ConnOnline {
			me.connStatus = ConnOnline
			me.pushConnStatus()
		}
	}
}

// Splits blocks in chunks of `size`. For example, given startBlock 10, toBlock 30 and size 3
// should return two arrays, startBlocks[10, 13, 16, 19] and toBlocks[12, 15, 18, 21]
func (me *baseClient) getBlockChunks(startBlock *big.Int, toBlock *big.Int, size int) ([]*big.Int, []*big.Int) {
	if startBlock == nil || toBlock == nil {
		log.Fatal("[baseClient][getBlockChunks] startBlock and toBlock parameters can't be nil")
	}
	var startBlocks []*big.Int
	var toBlocks []*big.Int
	counter := 0
	for i := startBlock; i.CmpAbs(toBlock) < 0; i.Add(i, big.NewInt(int64(size))) {
		startBlockTemp := new(big.Int).Set(i) // we want it by value not reference
		toBlockTemp := new(big.Int).Set(i)
		toBlockTemp.Add(toBlockTemp, big.NewInt(int64(size)))
		if toBlockTemp.Cmp(toBlock) > 0 {
			toBlockTemp.Set(toBlock)
		}
		if counter == 0 {
			startBlocks = append(startBlocks, startBlockTemp)
		} else {
			startBlockTemp.Add(startBlockTemp, big.NewInt(1))
			startBlocks = append(startBlocks, startBlockTemp)
		}
		toBlocks = append(toBlocks, toBlockTemp)
		counter++
	}
	return startBlocks, toBlocks
}

func (me *baseClient) pushConnStatus() {
	me.notify(&PendingTx{Type: ConnStatusNotification}, "", string(me.connStatus))
}

func (me *baseClient) putEventBlock(maxBlock uint64) error {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, maxBlock)
	return me.eventsDB.Put(me.getLastBlockKey(), b)

}

func (me *baseClient) getLastBlockKey() []byte {
	// reverse key concat to prevent from issues on FilterKeySuffix
	// because we use the same db for to store the last synced block and the signing events
	return []byte(me.currentAddress + "_" + lastBlockKey)
}

func (me *baseClient) currentEthAddress() common.Address {
	return common.HexToAddress(me.currentAddress)
}

func (me *baseClient) alreadyExecutedRecently(lg *types.Log) (string, bool) {
	var exists bool
	topicsLen := len(lg.Topics)
	eventNameHash := ""
	if topicsLen > 0 {
		eventNameHash = lg.Topics[0].String()
	}
	eventKey := fmt.Sprintf("%v-%v-%v-%v-%v", lg.BlockNumber, lg.Index, lg.TxHash.String(), topicsLen, eventNameHash)
	_ = me.ensureUniquenessOfEventCache.Get(eventKey, &exists)
	return eventKey, exists
}

func (me *baseClient) alreadyExecutedSuccessfully(eventKey string) {
	me.ensureUniquenessOfEventCache.Put(eventKey, true)
}

func (me *baseClient) notify(tx *PendingTx, txHash, status string) error {
	me.listenerLock.RLock()
	defer me.listenerLock.RUnlock()
	if me.txListener != nil {
		return me.txListener(tx, txHash, status)
	}
	return os.ErrClosed
}

func (me *baseClient) delEvent(fhash, currentAddr string) error {
	return me.eventsDB.Del([]byte(strings.ToLower(fhash) + "_" + currentAddr))
}

func (me *baseClient) ctxWithTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.TODO(), time.Duration(8*time.Second))
}

func (me *baseClient) deserializePendingTx(bts []byte) *PendingTx {
	if len(bts) > 0 {
		tx := PendingTx{}
		err := json.Unmarshal(bts, &tx)
		if err != nil {
			log.Println("[baseClient][deserializePendingTx] error: ", err, string(bts))
			return nil
		}
		log.Printf("[baseClient][deserializePendingTx] found pendingTx in db: filehash: %s, filename: %s, txhash: %s\n",
			tx.FileHash, tx.FileName, tx.TxHash)
		if len(tx.TxBts) > 0 {
			t := &types.Transaction{}
			err = t.UnmarshalJSON(tx.TxBts)
			if err != nil {
				log.Println("[baseClient][deserializePendingTx] error: ", err, string(tx.TxBts))
				return nil
			}
			tx.Tx = t
		}
		return &tx
	}
	return nil
}

func (me *baseClient) getAuth(ethPrivKey string) (*bind.TransactOpts, error) {
	ecdsaPriv, err := crypto.HexToECDSA(ethPrivKey)
	if err != nil {
		return nil, err
	}
	auth := bind.NewKeyedTransactor(ecdsaPriv)
	return auth, nil
}

func (me *baseClient) estimateGas(msg ethereum.CallMsg) (*gasCacheEntry, error) {
	gasKey, err := json.Marshal(msg)
	var gascEntry *gasCacheEntry
	k := string(gasKey)
	err = me.gasCache.Get(k, &gascEntry)
	if err == nil {
		return gascEntry, nil
	}
	ctx, cancel := me.ctxWithTimeout()
	gasLimit, err := me.ethconn.EstimateGas(ctx, msg)
	cancel()
	if err != nil {
		if strings.Contains(err.Error(), "always failing transaction") {
			return nil, err
		}
		log.Println("EstimateGas: failed to call estimateGas will use defaultGas definition", defaultGas)
		return defaultGas, nil
	}
	ctx, cancel = me.ctxWithTimeout()
	gasPrice, err := me.ethconn.SuggestGasPrice(ctx)
	cancel()
	if err != nil {
		log.Println("SuggestGasPrice: failed to call SuggestGasPrice will use defaultGas definition", defaultGas)
		return defaultGas, nil
	}
	newGasce := &gasCacheEntry{}
	// multiply price by 1.6
	newGasce.GasPrice = gasPrice.Mul(gasPrice, big.NewInt(16)).Div(gasPrice, big.NewInt(10))
	newGasce.GasLimit = uint64(float64(gasLimit) * float64(1.6))
	me.gasCache.Put(k, newGasce)
	log.Println("EstimateGas: estimated and cached new gas limit", newGasce)
	return newGasce, nil
}

func (me *baseClient) putXesTx(stype string, from common.Address, tx *types.Transaction, xesAmount *big.Int) {
	me.putTxWithOpts(stype, "", "", from, tx, xesAmount, nil)
}

func (me *baseClient) putTx(stype string, from common.Address, tx *types.Transaction) {
	me.putTxWithOpts(stype, "", "", from, tx, nil, nil)
}

func (me *baseClient) putFileHashTx(stype, fhash string, from common.Address, tx *types.Transaction) {
	me.putTxWithOpts(stype, fhash, "", from, tx, nil, nil)
}

func (me *baseClient) putFileTx(stype, fhash, fileName string, from common.Address, tx *types.Transaction, who *[]common.Address) {
	me.putTxWithOpts(stype, fhash, fileName, from, tx, nil, who)
}

func (me *baseClient) putFileSignTx(stype, fhash, fileName string, from common.Address, tx *types.Transaction, who *[]common.Address) {
	me.putTxWithOpts(stype, fhash, fileName, from, tx, nil, who)
}

func (me *baseClient) putTxWithOpts(stype, fhash, fileName string, from common.Address, tx *types.Transaction, xesAmount *big.Int, who *[]common.Address) {
	var Who []string = nil
	if who != nil {
		Who = make([]string, len(*who))
		for i, val := range *who {
			Who[i] = val.Hex()
		}
	}
	ptx := PendingTx{Type: stype, FileHash: fhash, FileName: fileName, CurrentAddress: strings.ToLower(from.Hex()), Tx: tx, XesAmount: xesAmount, Who: Who}
	log.Println("Tx broadcasted: ", stype, fhash, tx.Hash().Hex())
	me.notify(&ptx, ptx.Tx.Hash().Hex(), StatusPending)
	ptx.Tx = nil
	txbts, err := tx.MarshalJSON()
	if err != nil {
		log.Println("putTx: ", err)
		return
	}
	ptx.TxBts = txbts
	bts, err := json.Marshal(ptx)
	if err != nil {
		log.Println("putTx: ", err)
		return
	}
	if me.currentAddress != "" {
		me.pendingTxDB.Put([]byte(strings.ToLower(tx.Hash().Hex())+"_"+me.currentAddress), bts)
	}
}

func (me *baseClient) delTx(txHash string) {
	if me.currentAddress != "" {
		me.pendingTxDB.Del([]byte(strings.ToLower(txHash) + "_" + me.currentAddress))
	}
}

func (me *baseClient) getAllTx() []*PendingTx {
	if me.currentAddress == "" {
		return nil
	}
	keys, _ := me.pendingTxDB.FilterKeySuffix([]byte(me.currentAddress))
	res := make([]*PendingTx, 0)
	for _, btxHash := range keys {
		tx := me.getTxByKey(btxHash)
		if tx != nil {
			res = append(res, tx)
		}
	}
	return res
}

func (me *baseClient) getTxByKey(key []byte) *PendingTx {
	bts, _ := me.pendingTxDB.Get([]byte(key))
	return me.deserializePendingTx(bts)
}

func (me *baseClient) getTx(txHash string) *PendingTx {
	if me.currentAddress != "" {
		return me.getTxByKey([]byte(strings.ToLower(txHash) + "_" + me.currentAddress))
	}
	return nil
}

func (me *baseClient) HeaderByNumber(number *big.Int) (*types.Header, error) {
	ctx, cancel := me.ctxWithTimeout()
	defer cancel()
	return me.ethconn.HeaderByNumber(ctx, number)
}

func (me *baseClient) remListener() {
	me.listenerLock.Lock()
	me.txListener = nil
	me.currentAddress = ""
	me.listenerLock.Unlock()
}

func (me *baseClient) close() {
	if me.ethconn != nil {
		me.ethconn.Close()
	}
	me.stopWorkers()
	if me.ensureUniquenessOfEventCache != nil {
		me.ensureUniquenessOfEventCache.Close()
	}
	if me.gasCache != nil {
		me.gasCache.Close()
	}
	if me.pendingTxDB != nil {
		me.pendingTxDB.Close()
	}
	if me.eventsDB != nil {
		me.eventsDB.Close()
	}
	me.remListener()
}
