package ethereum

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"git.proxeus.com/core/central/dapp/core/util"
	"git.proxeus.com/core/central/spp/fs"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"git.proxeus.com/core/central/dapp/core/ethglue"
	"git.proxeus.com/core/central/spp/eth"
)

type (
	DappClient struct {
		baseClient    *baseClient
		xesClient     *xesClient
		fsClient      *fsClient
		fsTransactor  *fsTransactor
		xesTransactor *xesTransactor
		ethTransactor *ethTransactor
		listenerLock  *sync.RWMutex

		pendingTxTrigger chan struct{}

		//to prevent from duplication with polling and pushing
		//while notifying success or fail
		pendingSync sync.Mutex

		xesABI       abi.ABI
		proxeusFSABI abi.ABI

		xesAddress common.Address
		pfsAddress common.Address
	}
	PendingTx struct {
		CurrentAddress string
		FileName       string
		Type           string
		FileHash       string
		TxHash         string
		Tx             *types.Transaction
		TxBts          []byte
		XesAmount      *big.Int
		Who            []string
	}
	uniqueFileHash struct {
		Expired int64
	}
)

const (
	StatusPending = "pending"
	StatusSuccess = "success"
	StatusFail    = "fail"

	PendingTypeShare       = "share"
	PendingTypeRemove      = "remove"
	PendingTypeRegister    = "register"
	PendingTypeRevoke      = "revoke"
	PendingTypeSign        = "sign"
	PendingTypeSignRequest = "requestSign"
	EventSigningRequest    = "signingRequest"
	EventNotifySign        = "notifySign"
	Event                  = "someEvent"

	ConnStatusNotification = "connectionStatus"
)

type EthConnectionStatus string

const (
	ConnOffline EthConnectionStatus = "offline"
	ConnSyncing EthConnectionStatus = "syncing"
	ConnOnline  EthConnectionStatus = "online"
)

// File types
const (
	NONE             = int64(0)
	THUMBNAIL        = int64(1)
	UNDEFINEDSIGNERS = int64(2)
	DEFINEDSIGNERS   = int64(3)

	TransactionsDBName = "transactions"
)

var (
	ErrFileNotFound = errors.New("file not found")
)

func NewDappClient(ethClientURL, wsUrl, XESAddress, proxeusFSAddress string) (*DappClient, error) {
	var err error
	me := &DappClient{xesAddress: common.HexToAddress(XESAddress), pfsAddress: common.HexToAddress(proxeusFSAddress)}

	me.listenerLock = new(sync.RWMutex)
	me.baseClient = NewBaseClient(me.listenerLock, wsUrl, ethClientURL, []common.Address{me.pfsAddress, me.xesAddress})

	// to avoid panic when offline
	me.baseClient.ethconn, err = ethclient.Dial("http://localhost/")
	me.baseClient.nonceManager.OnDial(me.baseClient.ethconn)
	me.baseClient.connStatus = ConnOffline
	me.baseClient.pushConnStatus()

	go func() {
		for {
			me.baseClient.ethconn, err = ethglue.Dial(ethClientURL)
			if err != nil {
				fmt.Println("eth dial err", err)
				continue
			}
			me.baseClient.connStatus = ConnSyncing
			me.updateEthInterfaces()
			me.baseClient.pushConnStatus()
			me.baseClient.nonceManager.OnDial(me.baseClient.ethconn)
			break
		}
	}()

	me.xesABI, err = abi.JSON(strings.NewReader(eth.XESTokenContractABI))
	if err != nil {
		return nil, err
	}
	me.proxeusFSABI, err = abi.JSON(strings.NewReader(eth.ProxeusFSContractABI))
	if err != nil {
		return nil, err
	}

	tokenTransactorMutex := new(sync.Mutex)
	me.xesClient = NewXesClient(me.baseClient, me.xesAddress, me.xesABI)
	me.fsClient, err = NewFsClient(me.baseClient, me.pfsAddress)
	if err != nil {
		return nil, err
	}
	me.fsTransactor = NewFsTransactor(me.baseClient, me.pfsAddress, me.proxeusFSABI, &me.baseClient.nonceManager)
	me.xesTransactor = NewXesTransactor(me.baseClient, me.xesAddress, me.pfsAddress, me.xesABI, &me.baseClient.nonceManager, tokenTransactorMutex)
	me.ethTransactor = NewEthTransactor(me.baseClient, &me.baseClient.nonceManager, tokenTransactorMutex)

	err = me.updateEthInterfaces()
	if err != nil {
		return nil, err
	}
	return me, nil
}

func (me *DappClient) initLocalStorage(storageDir string) error {
	var err error

	if storageDir == "" {
		storageDir = "."
	}
	storageDir = filepath.Join(storageDir, TransactionsDBName)
	if err != nil {
		return err
	}

	if err = me.fsTransactor.initLocalStorage(storageDir); err != nil {
		return err
	}

	if err = me.fsClient.initLocalStorage(storageDir); err != nil {
		return err
	}

	if err = me.baseClient.initLocalStorage(storageDir); err != nil {
		return err
	}

	return nil
}

func (me *DappClient) updateEthInterfaces() error {
	var err error
	me.xesTransactor.xesTokenContractTransactor, err = eth.NewXESTokenContractTransactor(me.xesAddress, me.baseClient.ethconn)
	if err != nil {
		return err
	}
	me.fsTransactor.proxeusFSContractTransactor, err = eth.NewProxeusFSContractTransactor(me.pfsAddress, me.baseClient.ethconn)
	if err != nil {
		return err
	}
	me.xesClient.xesTokenContractCaller, err = eth.NewXESTokenContractCaller(me.xesAddress, me.baseClient.ethconn)
	if err != nil {
		return err
	}
	me.fsClient.proxeusFSContractCaller, err = eth.NewProxeusFSContractCaller(me.pfsAddress, me.baseClient.ethconn)
	if err != nil {
		return err
	}

	return nil
}

func (me *DappClient) startWorkers() {
	me.baseClient.startWorkers()
	me.pendingTxTrigger = make(chan struct{}, 50)
	go me.pendingTx()
}

func (me *DappClient) stopWorkers() {
	if me.pendingTxTrigger != nil {
		close(me.pendingTxTrigger)
	}
}

func (me *DappClient) pollPendingTxs() {
	me.pendingSync.Lock()
	defer me.pendingSync.Unlock()
	pendingTxs := me.baseClient.getAllTx()
	for _, tx := range pendingTxs {
		if me.baseClient.stopping {
			return
		}
		me.handlePendingTx(tx)
	}
}

func (me *DappClient) listenForPendingTx(txHash string) {
	me.pendingSync.Lock()
	defer me.pendingSync.Unlock()
	tx := me.baseClient.getTx(txHash)
	me.handlePendingTx(tx)
}

func (me *DappClient) handlePendingTx(tx *PendingTx) {
	if me.baseClient.stopping {
		return
	}
	if tx == nil {
		return
	}
	if tx.CurrentAddress == me.baseClient.currentAddress {
		ctx, cancel := me.baseClient.ctxWithTimeout()
		r, err := me.baseClient.ethconn.TransactionReceipt(ctx, tx.Tx.Hash())
		cancel()
		if me.baseClient.stopping {
			return
		}
		if err != nil || r == nil || me.baseClient.stopping {
			return
		}

		txHash := tx.Tx.Hash().Hex()
		status := StatusSuccess
		if types.ReceiptStatusSuccessful != r.Status {
			status = StatusFail
		}
		err = me.baseClient.notify(tx, txHash, status)
		if err == nil {
			log.Printf("[dappClient][handlePendingTx] handling %s with status %s", tx.Type, status)
			if status == StatusSuccess && (tx.Type == PendingTypeSign || tx.Type == PendingTypeRemove) {
				signingCacheKey := tx.FileHash + "_" + tx.CurrentAddress
				evData := me.fsClient.getSigningRequestFromCache([]byte(signingCacheKey))
				if evData == nil {
					log.Printf("[fsClient][getSigningRequestFromCache] Did not find pendingTx in eventsDB for key: %s", signingCacheKey)
				} else {
					log.Printf("[dappClient][handlePendingTx] got SigningRequestFromCache txHash: %s for fileHash: %s and will notify", tx.TxHash, tx.FileHash)
					me.baseClient.delEvent(tx.FileHash, tx.CurrentAddress)
					me.baseClient.notify(evData, evData.TxHash, status)
				}
			}
			//remove pending tx from the store if notification went well
			me.baseClient.delTx(txHash)
		}
	}
}

func (me *DappClient) pendingTx() {
	me.baseClient.stopWg.Add(1)
	defer func() {
		me.baseClient.stopping = true
		me.baseClient.stopWg.Done()
		log.Println("pending() ended...")
	}()
	log.Println("pending() started...")
	for {
		select {
		case _, ok := <-me.pendingTxTrigger:
			if !ok {
				return
			}
			if me.baseClient.stopping {
				return
			}
			me.notifyPendingTxStatus()
		case <-me.baseClient.stopChan:
			return
		}
	}
}

func (me *DappClient) notifyPendingTxStatus() {
	me.pendingSync.Lock()
	defer me.pendingSync.Unlock()
	ptxs := me.baseClient.getAllTx()
	for _, tx := range ptxs {
		if me.baseClient.stopping {
			return
		}
		me.baseClient.notify(tx, tx.Tx.Hash().Hex(), StatusPending)
	}
	//notify pending singing requests
	for _, ptx := range me.fsClient.listAllSigningRequests() {
		if me.baseClient.stopping {
			return
		}
		me.baseClient.notify(ptx, ptx.TxHash, StatusPending)
	}
}

// Processes a Log's events
// After it went through all the possible cases mark the log as processed. Exists the function whenever an error occurs
func (me *DappClient) DefaultEventsHandler(lg *types.Log, recent bool) error {
	xesTransferEvent, xesTransferEventErr := me.xesClient.logAsXesTransfer(lg, recent)
	if xesTransferEvent != nil && xesTransferEventErr == nil {
		log.Printf("Event[xesTransferEvent] %v %v incoming... tx %s from: %s, to: %s\n", lg.BlockNumber, lg.TxIndex, lg.TxHash.Hex(),
			xesTransferEvent.FromAddress.String(), xesTransferEvent.ToAddress.String())
		xesTransferEventErr = me.xesClient.handleXesTransferEvent(xesTransferEvent.FromAddress, xesTransferEvent.ToAddress, xesTransferEvent.Raw.TxHash, xesTransferEvent.Value)
	}

	signEvent, signEventErr := me.fsClient.LogAsRequestSign(lg, recent)
	if signEvent != nil && signEventErr == nil {
		me.eventNotify(&signEvent.Raw, signEvent.Hash, recent)
		toAddr := strings.ToLower("0x" + hex.EncodeToString(signEvent.To[:]))
		if toAddr == me.baseClient.currentAddress { // request directed to us
			log.Printf("Event[RequestSign] %v %v incoming... tx %s fileHash %s\n", lg.BlockNumber, lg.TxIndex, lg.TxHash.Hex(), common.Hash(signEvent.Hash).Hex())
			signEventErr = me.fsClient.handleRequestSignEvent(signEvent.Hash, signEvent.Raw.TxHash)
		}

		if signEventErr = me.handleRequestUndefinedSignEvent(signEvent); signEventErr != nil {
			log.Println("Event[RequestSign] error: ", signEventErr.Error())
		}
	}

	notifySign, notifySignErr := me.fsClient.LogAsNotifySign(lg, recent)
	if notifySign != nil && notifySignErr == nil {
		if recent {
			_, notifySignErr = me.fsClient.fileSigners(notifySign.Hash, false)
		}
		if notifySignErr == nil {
			log.Printf("Event[NotifySign] %v %v incoming... tx %s fileHash %s\n", lg.BlockNumber, lg.TxIndex, lg.TxHash.Hex(), common.Hash(notifySign.Hash).Hex())
			me.eventNotifySign(notifySign)
		}
	}

	del, delErr := me.fsClient.LogAsDeleted(lg, recent)
	if del != nil && delErr == nil {
		log.Printf("Event[Delete] %v %v incoming... tx %s fileHash %s\n", lg.BlockNumber, lg.TxIndex, lg.TxHash.Hex(), common.Hash(del.Hash).Hex())
		me.eventNotify(&del.Raw, del.Hash, recent)
		log.Println("[dappClient][DefaultEventsHandler] will notify pendingTx of type 'remove' with status 'success'")
		deletePendingTx := &PendingTx{TxHash: del.Raw.TxHash.Hex(), FileHash: strings.ToLower(common.Hash(del.Hash).Hex()), Type: PendingTypeRemove}
		if err := me.baseClient.notify(deletePendingTx, del.Raw.TxHash.Hex(), StatusSuccess); err != nil {
			log.Println("[dappClient][DefaultEventsHandler] error on notify, err: ", err.Error())
		}
	}

	upEv, upEvErr := me.fsClient.LogAsUpdatedEvent(lg, recent)
	if upEv != nil && upEvErr == nil {
		if recent {
			_, err := me.fsClient.fileSigners(upEv.NewHash, false)
			if err != nil {
				upEvErr = err
			}
			_, err = me.FileList(false)
			if err != nil {
				upEvErr = err
			}
		}
		if upEvErr == nil {
			log.Printf("Event[UpdatedEvent] %v %v incoming... tx %s fileHash %s\n", lg.BlockNumber, lg.TxIndex, lg.TxHash.Hex(), common.Hash(upEv.NewHash).Hex())
			me.eventNotify(&upEv.Raw, upEv.OldHash, recent)
			me.eventNotify(&upEv.Raw, upEv.NewHash, recent)
		}
	}

	oc, ocErr := me.fsClient.LogAsOwnerChanged(lg, recent)
	if oc != nil && ocErr == nil {
		log.Printf("Event[OwnerChanged] %v %v incoming... tx %s fileHash %s\n", lg.BlockNumber, lg.TxIndex, lg.TxHash.Hex(), common.Hash(oc.Hash).Hex())
		me.eventNotify(&oc.Raw, oc.Hash, recent)
	}

	if signEventErr != nil {
		return signEventErr
	}
	if notifySignErr != nil {
		return notifySignErr
	}
	if delErr != nil {
		return delErr
	}
	if upEvErr != nil {
		return upEvErr
	}
	if ocErr != nil {
		return ocErr
	}
	if xesTransferEventErr != nil {
		return xesTransferEventErr
	}
	return nil
}

func (me *DappClient) eventNotifySign(NotifySign *eth.ProxeusFSContractNotifySign) {
	r := NotifySign.Raw
	var fhash common.Hash = NotifySign.Hash
	if me.baseClient.currentAddress != "" {
		myAddr := common.HexToAddress(me.baseClient.currentAddress)
		//maybe not for current account because of wrong event handling
		if util.Bytes32Empty(fhash) {
			return
		}
		yes, err := me.fsClient.hasReadRights(fhash, myAddr, true)
		if !yes || err != nil {
			return
		}
		who := make([]string, 1)
		who[0] = strings.ToLower(NotifySign.Who.Hex())
		if myAddr.Hex() == me.baseClient.currentAddress {
			me.baseClient.notify(&PendingTx{TxHash: r.TxHash.Hex(), FileHash: strings.ToLower(fhash.Hex()),
				Who: who, Type: EventNotifySign}, r.TxHash.Hex(), StatusSuccess)
		}
		fi, err := me.FileInfo(NotifySign.Hash, true)
		if err != nil {
			return
		}
		if fi.FileType.Int64() == UNDEFINEDSIGNERS {
			me.handleSigningFileWithUndefinedSigners(NotifySign)
		}
	}
}

func (me *DappClient) handleSigningFileWithUndefinedSigners(notifySign *eth.ProxeusFSContractNotifySign) {
	// note we don't want to read from the cache at this point
	signers, err := me.fsClient.fileSigners(notifySign.Hash, false)
	allHaveSigned := true
	if err != nil {
		return
	}
	for _, r := range signers {
		if isEmptyAddress(r) {
			allHaveSigned = false
			break
		}
	}
	log.Printf("[fsClient][LogAsNotifySign] allHaveSigned: %t, file: %s", allHaveSigned, common.Hash(notifySign.Hash).Hex())
	if !allHaveSigned {
		return
	}
	log.Println("[fsClient][LogAsNotifySign] all undefined signatories have signed, will handlePendingTx and clean signingRequests for file: ")
	me.listenForPendingTx(notifySign.Raw.TxHash.Hex())
	me.fsClient.cleanSigningRequest(notifySign.Hash)
}

func (me *DappClient) eventNotify(eventLogEntry *types.Log, fhash common.Hash, recent bool) {
	if me.baseClient.currentAddress == "" {
		return
	}
	myAddr := common.HexToAddress(me.baseClient.currentAddress)
	//maybe not for current account because of wrong event handling
	if util.Bytes32Empty(fhash) {
		return
	}
	readableForCurrentAcc, err := me.fsClient.hasReadRights(fhash, myAddr, true)
	if !readableForCurrentAcc || err != nil {
		return
	}
	if recent {
		//update FileInfo cache
		fi, err := me.fsClient.FileInfo(fhash, false)

		//automatically add signing request if we are the defined signer for a specific file
		if err == nil && fi.FileType.Int64() == DEFINEDSIGNERS {
			for _, addr := range fi.DefinedSigners {
				if bytes.Equal(addr.Bytes(), myAddr.Bytes()) {
					me.fsClient.handleRequestSignEvent(fhash, eventLogEntry.TxHash)
				}
			}
		}
	}
	me.baseClient.notify(&PendingTx{TxHash: eventLogEntry.TxHash.Hex(), FileHash: strings.ToLower(fhash.Hex()), Type: Event}, eventLogEntry.TxHash.Hex(), StatusSuccess)
}

func (me *DappClient) handleRequestUndefinedSignEvent(signEvent *eth.ProxeusFSContractRequestSign) error {
	if me.baseClient.currentAddress == "" {
		return os.ErrInvalid
	}
	myAddr := common.HexToAddress(me.baseClient.currentAddress)

	fi, err := me.fsClient.FileInfo(signEvent.Hash, false)
	if err != nil {
		return err
	}

	if fi.FileType.Int64() == UNDEFINEDSIGNERS {
		writableForCurrentAcc, err := me.fsClient.hasWriteRights(signEvent.Hash, myAddr, true)
		if err == nil && writableForCurrentAcc == true {
			//assume we are owner of file
			toAddr := strings.ToLower("0x" + hex.EncodeToString(signEvent.To[:]))
			if err = me.fsClient.handleRequestUndefinedSignEvent(signEvent.Hash, signEvent.Raw.TxHash, toAddr); err != nil {
				return err
			}
		}
	}
	return nil
}

func (me *DappClient) GetSentSignRequestAddrForFileWithUndefinedSigners(fileHash string) []common.Address {
	if fileHash == "" {
		log.Println("[app][checkForAlreadySentSignRequests] fileHash not set.")
		return nil
	}
	pendingTx := me.getSigningRequestUndefinedSignersCache(fileHash)
	if pendingTx == nil {
		return nil
	}

	signRequestAddresses := pendingTx.Who
	var signReqAddresses []common.Address
	for _, addr := range signRequestAddresses {
		signReqAddresses = append(signReqAddresses, common.HexToAddress(addr))
	}
	return signReqAddresses
}

func (me *DappClient) getSigningRequestUndefinedSignersCache(fileHash string) *PendingTx {
	return me.fsClient.GetSigningRequestUndefinedSignersCache(fileHash)
}

func (me *DappClient) BalanceETHof(ethAddress string) (*big.Int, error) {
	ctx, cancel := me.baseClient.ctxWithTimeout()
	b, err := me.baseClient.ethconn.BalanceAt(ctx, common.HexToAddress(ethAddress), nil)
	cancel()
	return b, err
}

func (me *DappClient) InitListeners(storDir, ethAddr string, xtl func(tx *PendingTx, txHash, status string) error, eventHandler func(lg *types.Log, recent bool) error) {
	if ethAddr == "" {
		log.Println("[client][InitListeners] error: ethAddr parameter is empty")
		return
	}
	err := me.initLocalStorage(storDir)
	if err != nil {
		log.Println("[client][InitListeners] error: ", err.Error())
	}

	me.baseClient.initListeners(ethAddr, xtl, me.listenForPendingTx, me.pollPendingTxs, eventHandler, me.updateEthInterfaces)

	me.baseClient.listenerLock.Lock()
	me.baseClient.nonceManager.OnAccountChange(me.baseClient.currentAddress)
	me.baseClient.listenerLock.Unlock()

	me.startWorkers()
}

func (me *DappClient) NotifyLastState() {
	go func() {
		me.baseClient.pushConnStatus()
		defer func() {
			if r := recover(); r != nil { //can happen when closing
				log.Println("PendingTrigger client channel closed", r)
			}
		}()
		me.pendingTxTrigger <- struct{}{}
	}()
}

func (me *DappClient) BalanceXESof(ethAddress string) (*big.Int, error) {
	return me.xesClient.balanceXESof(ethAddress)
}

func (me *DappClient) ProxeusFSAllowance(ethAddr string) (*big.Int, error) {
	return me.xesClient.proxeusFSAllowance(ethAddr, me.pfsAddress)
}

func (me *DappClient) Tx(ethPrivKey string) (*bind.TransactOpts, error) {
	ecdsaPriv, err := crypto.HexToECDSA(ethPrivKey)
	if err != nil {
		return nil, err
	}
	auth := bind.NewKeyedTransactor(ecdsaPriv)

	return auth, nil
}

func (me *DappClient) ETHTransferEstimateGas() (gasPrice *big.Int, gasLimit uint64, err error) {
	return me.ethTransactor.ethTransferEstimateGas()
}

func (me *DappClient) ETHTransfer(ethPrivKeyFrom, ethAddressTo string, ethAmount *big.Int) (*types.Transaction, error) {
	return me.ethTransactor.ethTransfer(ethPrivKeyFrom, ethAddressTo, ethAmount)
}

func (me *DappClient) XESTransferEstimateGas(ethPrivKeyFrom string, ethAddressTo string, xesAmount *big.Int) (*bind.TransactOpts, error) {
	return me.xesTransactor.xesTransferEstimateGas(ethPrivKeyFrom, ethAddressTo, xesAmount)
}

func (me *DappClient) XESTransfer(ethPrivKeyFrom string, ethAddressTo string, xesAmount *big.Int) (*types.Transaction, error) {
	return me.xesTransactor.xesTransfer(ethPrivKeyFrom, ethAddressTo, xesAmount)
}

func (me *DappClient) XESApprove(ethPrivKeyFrom string, ethAddressTo string, xesAmount *big.Int) (*types.Transaction, error) {
	return me.xesTransactor.xesApprove(ethPrivKeyFrom, ethAddressTo, xesAmount)
}

func (me *DappClient) XESApproveToProxeusFSEstimateGas(ethPrivKeyFrom string, xesAmount *big.Int) (*bind.TransactOpts, error) {
	return me.xesTransactor.xesApproveToProxeusFSEstimateGas(ethPrivKeyFrom, xesAmount)
}

func (me *DappClient) XESApproveToProxeusFS(ethPrivKeyFrom string, xesAmount *big.Int) (*types.Transaction, error) {
	return me.xesTransactor.xesApproveToProxeusFS(ethPrivKeyFrom, xesAmount)
}

func (me *DappClient) XESTransferFrom(ethPrivKeyFrom, ethAddressFrom, ethAddressTo string, xesAmount *big.Int) (*types.Transaction, error) {
	return me.xesTransactor.xesTransferFrom(ethPrivKeyFrom, ethAddressFrom, ethAddressTo, xesAmount)
}

func (me *DappClient) CreateFileDefinedSignersEstimateGas(ethPrivKeyFrom string, fileHash [32]byte, fileName string, definedSigners []common.Address,
	expiry *big.Int, replacesFile [32]byte, storageProviders []common.Address, xesAmount *big.Int) (*bind.TransactOpts, error) {

	return me.fsTransactor.createFileDefinedSignersEstimateGas(ethPrivKeyFrom, fileHash, fileName, definedSigners, expiry,
		replacesFile, storageProviders, xesAmount)
}

func (me *DappClient) CreateFileDefinedSigners(ethPrivKeyFrom string, fileHash [32]byte, fileName string, definedSigners []common.Address,
	expiry *big.Int, replacesFile [32]byte, storageProviders []common.Address, xesAmount *big.Int) (*types.Transaction, error) {

	return me.fsTransactor.createFileDefinedSigners(ethPrivKeyFrom, fileHash, fileName, definedSigners, expiry,
		replacesFile, storageProviders, xesAmount)
}

func (me *DappClient) CreateFileShared(ethPrivKeyFrom string, fileHash [32]byte, fileName string, mandatorySigners *big.Int,
	expiry *big.Int, replacesFile [32]byte, storageProviders []common.Address, readers []common.Address,
	xesAmount *big.Int) (*types.Transaction, error) {

	return me.fsTransactor.createFileShared(ethPrivKeyFrom, fileHash, fileName, mandatorySigners, expiry, replacesFile,
		storageProviders, readers, xesAmount)
}

func (me *DappClient) CreateFileUndefinedSignersEstimateGas(ethPrivKeyFrom string, fileHash [32]byte, fileName string, mandatorySigners *big.Int,
	expiry *big.Int, replacesFile [32]byte, storageProviders []common.Address, xesAmount *big.Int) (*bind.TransactOpts, error) {

	return me.fsTransactor.createFileUndefinedSignersEstimateGas(ethPrivKeyFrom, fileHash, fileName, mandatorySigners,
		expiry, replacesFile, storageProviders, xesAmount)
}

func (me *DappClient) CreateFileUndefinedSigners(ethPrivKeyFrom string, fileHash [32]byte, fileName string, mandatorySigners *big.Int,
	expiry *big.Int, replacesFile [32]byte, storageProviders []common.Address, xesAmount *big.Int) (*types.Transaction, error) {

	return me.fsTransactor.createFileUndefinedSigners(ethPrivKeyFrom, fileHash, fileName, mandatorySigners,
		expiry, replacesFile, storageProviders, xesAmount)
}

func (me *DappClient) fileRemoveEstimateGas(ethPrivKeyFrom string, fileHash [32]byte) (*bind.TransactOpts, error) {
	return me.fsTransactor.fileRemoveEstimateGas(ethPrivKeyFrom, fileHash)
}

func (me *DappClient) FileInfo(fileHash [32]byte, readFromCache bool) (fi fs.FileInfo, err error) {
	return me.fsClient.FileInfo(fileHash, readFromCache)
}

//[0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0]
func (me *DappClient) XESAmountPerFile(prvs []common.Address) (*big.Int, error) {
	return me.fsClient.xesAmountPerFile(prvs)
}

func (me *DappClient) FileHashToFileHashOrder(fileHash common.Hash) (fho *FileHashOrder) {
	return me.fsClient.fileHashToFileHashOrder(fileHash)
}

func (me *DappClient) FileList(readFromCache bool) (fis []*FileHashOrder, err error) {
	return me.fsClient.fileList(readFromCache)
}

func (me *DappClient) FileRemoveEstimateGas(ethPrivKeyFrom string, fileHash [32]byte) (*bind.TransactOpts, error) {
	return me.fsTransactor.fileRemoveEstimateGas(ethPrivKeyFrom, fileHash)
}

func (me *DappClient) FileRemove(ethPrivKeyFrom string, fileHash [32]byte, filename string) (*types.Transaction, error) {
	return me.fsTransactor.fileRemove(ethPrivKeyFrom, fileHash, filename)
}

func (me *DappClient) FileRequestAccess(ethPrivKeyFrom string, fileHash [32]byte) (*types.Transaction, error) {
	return me.fsTransactor.fileRequestAccess(ethPrivKeyFrom, fileHash)
}

func (me *DappClient) FileRequestSignEstimateGas(ethPrivKeyFrom string, fileHash [32]byte, signer []common.Address) (*bind.TransactOpts, error) {
	return me.fsTransactor.fileRequestSignEstimateGas(ethPrivKeyFrom, fileHash, signer)
}

func (me *DappClient) FileRequestSign(ethPrivKeyFrom string, fileHash [32]byte, filename string, signer []common.Address) (*types.Transaction, error) {
	return me.fsTransactor.fileRequestSign(ethPrivKeyFrom, fileHash, filename, signer)
}

func (me *DappClient) FileRevokePermEstimateGas(ethPrivKeyFrom string, fileHash [32]byte, addr []common.Address) (*bind.TransactOpts, error) {
	return me.fsTransactor.fileRevokePermEstimateGas(ethPrivKeyFrom, fileHash, addr)
}

func (me *DappClient) FileRevokePerm(ethPrivKeyFrom string, fileHash [32]byte, filename string, addr []common.Address) (*types.Transaction, error) {
	return me.fsTransactor.fileRevokePerm(ethPrivKeyFrom, fileHash, filename, addr)
}

func (me *DappClient) FileSetPermEstimateGas(ethPrivKeyFrom string, fileHash [32]byte, addr []common.Address) (*bind.TransactOpts, error) {
	return me.fsTransactor.FileSetPermEstimateGas(ethPrivKeyFrom, fileHash, addr)
}

func (me *DappClient) FileSetPerm(ethPrivKeyFrom string, fileHash [32]byte, filename string, addr []common.Address) (*types.Transaction, error) {
	return me.fsTransactor.fileSetPerm(ethPrivKeyFrom, fileHash, filename, addr)
}

func (me *DappClient) FileSignEstimateGas(ethPrivKeyFrom string, fileHash [32]byte) (*bind.TransactOpts, error) {
	return me.fsTransactor.fileSignEstimateGas(ethPrivKeyFrom, fileHash)
}

func (me *DappClient) FileSign(ethPrivKeyFrom, filename string, fileHash [32]byte) (*types.Transaction, error) {
	return me.fsTransactor.fileSign(ethPrivKeyFrom, filename, fileHash)
}

func (me *DappClient) SetDappVersion(ethPrivKeyFrom string, version [32]byte) (*types.Transaction, error) {
	return me.fsTransactor.setDappVersion(ethPrivKeyFrom, version)
}

func (me *DappClient) SpAdd(ethPrivKeyFrom string, strProv common.Address, urlPrefix [32]byte) (*types.Transaction, error) {
	return me.fsTransactor.spAdd(ethPrivKeyFrom, strProv, urlPrefix)
}

func (me *DappClient) SpUpdate(ethPrivKeyFrom string, strProv common.Address, urlPrefix [32]byte) (*types.Transaction, error) {
	return me.fsTransactor.spUpdate(ethPrivKeyFrom, strProv, urlPrefix)
}

type StorageProvider struct {
	Address string `json:"address"`
	URL     string `json:"url"`
}

func (me *DappClient) StorageProviders() ([]StorageProvider, error) {
	return me.fsClient.storageProviders()
}

func (me *DappClient) SpInfo(strProv common.Address) (string, error) {
	return me.fsClient.spInfo(strProv)
}

func (me *DappClient) SpInfoForFile(fileHash string) (string, error) {
	return me.fsClient.spInfoForFile(fileHash)
}

func isEmptyAddress(addr common.Address) bool {
	for _, a := range addr {
		if a != 0 {
			return false
		}
	}
	return true
}

func (me *DappClient) HasReadRights(fileHash [32]byte, addr common.Address, readFromCache bool) (bool, error) {
	return me.fsClient.hasReadRights(fileHash, addr, readFromCache)
}

func (me *DappClient) HasWriteRights(fileHash [32]byte, addr common.Address, readFromCache bool) (bool, error) {
	return me.fsClient.hasWriteRights(fileHash, addr, readFromCache)
}

func (me *DappClient) FileSigners(fileHash [32]byte, readFromCache bool) ([]common.Address, error) {
	return me.fsClient.fileSigners(fileHash, readFromCache)
}

func (me *DappClient) ContractVersion() string {
	return me.fsClient.contractVersion()
}

func (me *DappClient) Close() error {
	me.stopWorkers()
	if me.fsClient != nil {
		me.fsClient.close()
	}
	if me.fsTransactor != nil {
		me.fsTransactor.close()
	}
	if me.baseClient != nil {
		me.baseClient.close()
	}
	return nil
}
