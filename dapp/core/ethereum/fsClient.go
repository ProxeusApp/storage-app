package ethereum

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/ProxeusApp/storage-app/dapp/core/util"
	"github.com/ProxeusApp/storage-app/spp/fs"

	cache "github.com/ProxeusApp/memcache"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ProxeusApp/storage-app/dapp/core/embdb"
	"github.com/ProxeusApp/storage-app/spp/eth"
)

type (
	fsClient struct {
		baseClient   *baseClient
		pfsAddress   common.Address
		proxeusFSABI abi.ABI

		hasReadRightsCache  *cache.Cache
		hasWriteRightsCache *cache.Cache

		myFileHashesDB *embdb.DB

		proxeusFSContractCaller *eth.ProxeusFSContractCaller
		fileInfoCache           *cache.Cache
		fileSignersCache        *cache.Cache
		fileListCache           *cache.Cache
		fileVerifyCache         *cache.Cache
		genericCache            *cache.Cache
	}

	fileVerify struct {
		valid   bool
		signers []common.Address
	}
	FileHashOrder struct {
		FileHash common.Hash
		SCOrder  int
	}
)

const (
	splist                  = "splist"
	myFileHashesStorageName = "myFileHashes"
)

func NewFsClient(baseClient *baseClient, pfsAddress common.Address) (*fsClient, error) {
	var err error
	fsClient := new(fsClient)
	fsClient.baseClient = baseClient
	fsClient.pfsAddress = pfsAddress
	fsClient.proxeusFSABI, err = abi.JSON(strings.NewReader(eth.ProxeusFSContractABI))
	if err != nil {
		return nil, err
	}
	return fsClient, nil
}

func (me *fsClient) initLocalStorage(storageDir string) (err error) {
	me.fileSignersCache = cache.NewExtendExpiryOnGet(10*time.Minute, true)
	me.fileInfoCache = cache.NewExtendExpiryOnGet(10*time.Minute, true)
	me.hasReadRightsCache = cache.NewExtendExpiryOnGet(10*time.Minute, true)
	me.hasWriteRightsCache = cache.NewExtendExpiryOnGet(10*time.Minute, true)
	me.fileListCache = cache.NewExtendExpiryOnGet(10*time.Minute, true)
	me.fileVerifyCache = cache.NewExtendExpiryOnGet(10*time.Minute, true)
	me.genericCache = cache.New(30 * time.Minute)

	me.myFileHashesDB, err = embdb.Open(storageDir, myFileHashesStorageName)
	return
}

func (me *fsClient) LogAsPaymentReceived(lg *types.Log) *eth.ProxeusFSContractPaymentReceived {
	const PaymentReceived = "PaymentReceived"
	if _, ok := me.isNotProxeusFSEventOrWasExecutedAlready(PaymentReceived, lg); ok {
		return nil
	}
	event := new(eth.ProxeusFSContractPaymentReceived)
	if err := me.eventFromLog(event, lg, PaymentReceived); err != nil {
		return nil
	}

	//check if event is interesting is handled in spp
	event.Raw = *lg
	return event
}

func (me *fsClient) LogAsRequestSign(lg *types.Log, recent bool) (*eth.ProxeusFSContractRequestSign, error) {
	const RequestSign = "RequestSign"
	var eventKey string
	var ok bool
	if eventKey, ok = me.isNotProxeusFSEventOrWasExecutedAlready(RequestSign, lg); ok {
		return nil, nil
	}
	event := new(eth.ProxeusFSContractRequestSign)
	if err := me.eventFromLog(event, lg, RequestSign); err != nil {
		return nil, err
	}

	interesting, err := me.eventInterestingForMe(event.Hash, recent)
	if err != nil {
		return nil, err
	}
	if interesting {
		event.Raw = *lg
		return event, nil
	}
	me.baseClient.alreadyExecutedSuccessfully(eventKey)
	return nil, nil
}

func (me *fsClient) LogAsNotifySign(lg *types.Log, recent bool) (*eth.ProxeusFSContractNotifySign, error) {
	const NotifySign = "NotifySign"
	var eventKey string
	var ok bool
	if eventKey, ok = me.isNotProxeusFSEventOrWasExecutedAlready(NotifySign, lg); ok {
		return nil, nil
	}
	event := new(eth.ProxeusFSContractNotifySign)
	if err := me.eventFromLog(event, lg, NotifySign); err != nil {
		return nil, err
	}

	//TODO(ave) make more efficient
	//toAddr := strings.ToLower("0x" + hex.EncodeToString(event.Who[:]))
	//if toAddr != me.baseClient.currentAddress {
	//	return nil // request not directed to us
	//}
	interesting, err := me.eventInterestingForMe(event.Hash, recent)
	if err != nil {
		return nil, err
	}
	if interesting {
		event.Raw = *lg
		return event, nil
	}
	me.baseClient.alreadyExecutedSuccessfully(eventKey)
	return nil, nil
}

func (me *fsClient) LogAsDeleted(lg *types.Log, recent bool) (*eth.ProxeusFSContractDeleted, error) {
	const Deleted = "Deleted"
	var eventKey string
	var ok bool
	if eventKey, ok = me.isNotProxeusFSEventOrWasExecutedAlready(Deleted, lg); ok {
		return nil, nil
	}
	event := new(eth.ProxeusFSContractDeleted)
	if err := me.eventFromLog(event, lg, Deleted); err != nil {
		return nil, err
	}
	interesting, err := me.eventInterestingForMe(event.Hash, recent)
	if err != nil {
		return nil, err
	}
	if interesting {
		event.Raw = *lg
		me.cleanSigningRequest(event.Hash)
		return event, nil
	}
	me.baseClient.alreadyExecutedSuccessfully(eventKey)
	return nil, nil
}

func (me *fsClient) LogAsUpdatedEvent(lg *types.Log, recent bool) (*eth.ProxeusFSContractUpdatedEvent, error) {
	const UpdatedEvent = "UpdatedEvent"
	var eventKey string
	var ok bool
	if eventKey, ok = me.isNotProxeusFSEventOrWasExecutedAlready(UpdatedEvent, lg); ok {
		return nil, nil
	}
	event := new(eth.ProxeusFSContractUpdatedEvent)
	if err := me.eventFromLog(event, lg, UpdatedEvent); err != nil {
		return nil, err
	}
	interesting, err := me.eventInterestingForMe(event.NewHash, recent)
	if err != nil {
		return nil, err
	}
	if interesting {
		event.Raw = *lg
		return event, nil
	} else {
		var fhash common.Hash
		fhash = event.NewHash
		if me.isMyFileHash(fhash) {
			me.FileInfo(fhash, false)
			//file was unshared
			me.baseClient.notify(&PendingTx{TxHash: lg.TxHash.Hex(), FileHash: strings.ToLower(fhash.Hex()), Type: Event}, lg.TxHash.Hex(), StatusSuccess)
			me.delMyFileHash(fhash)
		}
		fh := strings.ToLower(fhash.Hex())
		evData := me.getSigningRequestFromCache([]byte(fh + "_" + me.baseClient.currentAddress))
		if evData != nil {
			me.baseClient.delEvent(fh, me.baseClient.currentAddress)
			me.baseClient.notify(evData, evData.TxHash, StatusFail)
		}
	}
	me.baseClient.alreadyExecutedSuccessfully(eventKey)
	return nil, nil
}

func (me *fsClient) fileList(readFromCache bool) (fis []*FileHashOrder, err error) {
	if me.baseClient.currentAddress == "" {
		return nil, os.ErrPermission
	}
	addr := me.baseClient.currentAddress
	if readFromCache {
		err = me.fileListCache.Get(addr, &fis)
		if err == nil {
			return
		}
	}
	ctx, cancel := me.baseClient.ctxWithTimeout()
	var files [][32]byte
	opts := &bind.CallOpts{Pending: false, From: common.HexToAddress(addr), Context: ctx}
	files, err = me.proxeusFSContractCaller.FileList(opts)
	cancel()
	if err == nil {
		fis = make([]*FileHashOrder, len(files))
		for i, v := range files {
			f := &FileHashOrder{FileHash: v, SCOrder: i}
			fis[i] = f
			me.fileListCache.Put(f.FileHash.Hex()+addr, f)
			me.putMyFileHash(v)
		}
		me.fileListCache.Put(addr, fis)
	}
	return fis, err
}

func (me *fsClient) fileHashToFileHashOrder(fileHash common.Hash) (fho *FileHashOrder) {
	if me.baseClient.currentAddress == "" {
		return
	}
	addr := me.baseClient.currentAddress
	k := fileHash.Hex() + addr
	err := me.fileListCache.Get(k, &fho)
	if err != nil {
		//order does not exist, fetch it
		me.fileList(false)
		me.fileListCache.Get(k, &fho)
	}
	return
}

func (me *fsClient) getMyFileHashKey(fhash common.Hash) []byte {
	return []byte(strings.ToLower(fhash.Hex()) + "_" + me.baseClient.currentAddress)
}

func (me *fsClient) putMyFileHash(fhash common.Hash) {
	me.myFileHashesDB.Put(me.getMyFileHashKey(fhash), fhash.Bytes())
}

func (me *fsClient) isMyFileHash(fhash common.Hash) bool {
	_, err := me.getMyFileHash(fhash)
	return err == nil
}

func (me *fsClient) getMyFileHash(fhash common.Hash) (common.Hash, error) {
	var f common.Hash
	bts, _ := me.myFileHashesDB.Get(me.getMyFileHashKey(fhash))
	if len(bts) > 0 {
		return common.BytesToHash(bts), nil
	}
	return f, os.ErrNotExist
}

func (me *fsClient) delMyFileHash(fhash common.Hash) error {
	return me.myFileHashesDB.Del(me.getMyFileHashKey(fhash))
}

func (me *fsClient) LogAsOwnerChanged(lg *types.Log, recent bool) (*eth.ProxeusFSContractOwnerChanged, error) {
	const OwnerChanged = "OwnerChanged"
	var eventKey string
	var ok bool
	if eventKey, ok = me.isNotProxeusFSEventOrWasExecutedAlready(OwnerChanged, lg); ok {
		return nil, nil
	}
	event := new(eth.ProxeusFSContractOwnerChanged)
	if err := me.eventFromLog(event, lg, OwnerChanged); err != nil {
		return nil, err
	}
	interesting, err := me.eventInterestingForMe(event.Hash, recent)
	if err != nil {
		return nil, err
	}
	if interesting {
		event.Raw = *lg
		return event, nil
	}
	me.baseClient.alreadyExecutedSuccessfully(eventKey)
	return nil, nil
}

func (me *fsClient) eventFromLog(out interface{}, lg *types.Log, eventType string) error {
	pfsLogUnpacker := bind.NewBoundContract(me.pfsAddress, me.proxeusFSABI,
		me.baseClient.ethwsconn, me.baseClient.ethwsconn, me.baseClient.ethwsconn)
	err := pfsLogUnpacker.UnpackLog(out, eventType, *lg)
	if err != nil {
		return err // not our event type
	}
	return nil
}

func (me *fsClient) isNotProxeusFSEventOrWasExecutedAlready(eventName string, lg *types.Log) (eventKey string, doNotProceed bool) {
	if !me.isProxeusFSEvent(eventName, lg) {
		doNotProceed = true
		return
	}
	var ok bool
	if eventKey, ok = me.baseClient.alreadyExecutedRecently(lg); ok {
		doNotProceed = true
		return
	}
	return
}

func (me *fsClient) isProxeusFSEvent(eventName string, lg *types.Log) bool {
	if len(lg.Topics) > 0 {
		if bytes.Equal(me.proxeusFSABI.Events[eventName].Id().Bytes(), lg.Topics[0].Bytes()) {
			return true
		}
	}
	return false
}

func (me *fsClient) eventInterestingForMe(fhash common.Hash, recent bool) (bool, error) {
	if !util.Bytes32Empty(fhash) && me.baseClient.currentAddress != "" {
		//error handling not necessary
		//if err != nil interesting will be false anyway
		//we can catch up with another call
		interesting, err := me.hasReadRights(fhash, common.HexToAddress(me.baseClient.currentAddress), !recent)
		if err != nil {
			log.Println("can't read read rights", err.Error())
			return false, err
		}
		if interesting {
			//store the file to notify if we loose interest in it
			me.putMyFileHash(fhash)
		}
		return interesting, nil
	}
	return false, nil
}

func (me *fsClient) hasReadRights(fileHash [32]byte, addr common.Address, readFromCache bool) (bool, error) {
	if readFromCache {
		var hasReadRights bool
		err := me.hasReadRightsCache.Get(fileHash, &hasReadRights)
		if err == nil {
			return hasReadRights, nil
		}
	}
	ctx, cancel := me.baseClient.ctxWithTimeout()
	opts := &bind.CallOpts{Pending: false, Context: ctx}
	readRights, err := me.proxeusFSContractCaller.FileGetPerm(opts, fileHash, addr, false)
	cancel()
	if err == nil {
		me.hasReadRightsCache.Put(fileHash, readRights)
	}
	return readRights, err
}

func (me *fsClient) hasWriteRights(fileHash [32]byte, addr common.Address, readFromCache bool) (bool, error) {
	if readFromCache {
		var hasWriteRights bool
		err := me.hasWriteRightsCache.Get(fileHash, &hasWriteRights)
		if err == nil {
			return hasWriteRights, nil
		}
	}
	ctx, cancel := me.baseClient.ctxWithTimeout()
	opts := &bind.CallOpts{Pending: false, Context: ctx}
	writeRights, err := me.proxeusFSContractCaller.FileGetPerm(opts, fileHash, addr, true)
	cancel()
	if err == nil {
		me.hasWriteRightsCache.Put(fileHash, writeRights)
	}
	return writeRights, err
}

func (me *fsClient) cleanSigningRequest(fhash common.Hash) {
	if me.baseClient.currentAddress == "" {
		return
	}
	//----remove signing request cache if the file is removed or read access as been revoked
	fh := strings.ToLower(fhash.Hex())
	evData := me.getSigningRequestFromCache([]byte(fh + "_" + me.baseClient.currentAddress))
	if evData != nil {
		me.baseClient.delEvent(fh, me.baseClient.currentAddress)
		me.baseClient.notify(evData, evData.TxHash, StatusFail)

	}
	//----remove signing request cache if the file is removed or read access as been revoked
}

func (me *fsClient) isSigningRequestValid(ptx *PendingTx) (bool, error) {
	currentAddrStr := me.baseClient.currentAddress
	currentAddr := common.HexToAddress(me.baseClient.currentAddress)
	fhash := util.StrHexToBytes32(ptx.FileHash)
	fi, err := me.FileInfo(fhash, true)
	if err != nil {
		return false, err
	}
	if fi.Removed {
		if ptx != nil {
			me.baseClient.delEvent(ptx.FileHash, currentAddrStr)
			me.baseClient.notify(ptx, ptx.TxHash, StatusFail)
		}
		return false, err
	}
	//get status from bc
	signers, err := me.fileSigners(fhash, true)
	if err != nil {
		return false, err
	}
	for _, a := range signers {
		if bytes.Equal(currentAddr.Bytes(), a.Bytes()) {
			if ptx != nil {
				me.baseClient.delEvent(ptx.FileHash, currentAddrStr)
				return false, nil
			}
		}
	}
	return true, nil
}

func (me *fsClient) listAllSigningRequests() []*PendingTx {
	if me.baseClient.currentAddress == "" {
		return nil
	}

	keys, err := me.baseClient.eventsDB.FilterKeySuffix([]byte(me.baseClient.currentAddress))
	if err != nil {
		return nil
	}

	res := make([]*PendingTx, 0)
	for _, a := range keys {
		bts, err := me.baseClient.eventsDB.Get(a)
		if err == nil {
			ptx := me.baseClient.deserializePendingTx(bts)
			if ptx != nil {
				if valid, _ := me.isSigningRequestValid(ptx); valid {
					res = append(res, ptx)
				}
			}
		}
	}
	return res
}

func (me *fsClient) fileSigners(fileHash [32]byte, readFromCache bool) ([]common.Address, error) {
	if readFromCache {
		var singers []common.Address
		err := me.fileSignersCache.Get(fileHash, &singers)
		if err == nil {
			return singers, nil
		}
	}
	ctx, cancel := me.baseClient.ctxWithTimeout()
	opts := &bind.CallOpts{Pending: false, Context: ctx}
	res, err := me.proxeusFSContractCaller.FileSigners(opts, fileHash)
	cancel()
	if err == nil {
		me.fileSignersCache.Put(fileHash, res)
	}
	return res, err
}

func (me *fsClient) FileInfo(fileHash [32]byte, readFromCache bool) (fi fs.FileInfo, err error) {
	if readFromCache {
		err = me.fileInfoCache.Get(fileHash, &fi)
		if err == nil {
			return
		}
	}
	ctx, cancel := me.baseClient.ctxWithTimeout()
	opts := &bind.CallOpts{Pending: false, Context: ctx}
	fi, err = me.proxeusFSContractCaller.FileInfo(opts, fileHash)
	cancel()
	if err == nil {
		if fi.FileType.Int64() == NONE {
			return fi, ErrFileNotFound
		}
		//log.Println("fileinfo update cache", Bytes32ToHexStr(fi.Id))
		me.fileInfoCache.Put(fileHash, fi)
	}
	return
}

func (me *fsClient) handleRequestUndefinedSignEvent(fileHash, txHash common.Hash, toAddress string) error {
	curAddr := me.baseClient.currentAddress
	if curAddr == "" {
		return os.ErrPermission
	}
	fHash := strings.ToLower(fileHash.Hex())
	evd := me.GetSigningRequestUndefinedSignersCache(fHash)
	if evd != nil {
		if len(evd.Who) != 0 {
			for _, who := range evd.Who {
				if who == toAddress {
					return nil
				}
			}
		}
		evd.Who = append(evd.Who, toAddress)
	} else {
		evd = &PendingTx{}
		evd.CurrentAddress = curAddr
		evd.Who = []string{toAddress}
		evd.FileHash = fHash
		evd.Type = PendingTypeSignRequest
	}
	bts, err := json.Marshal(evd)
	if err != nil {
		return err
	}

	key := me.getSigningRequestUndefinedSignersCacheKey(fHash)
	return me.baseClient.eventsDB.Put(key, bts)
}

func (me *fsClient) handleRequestSignEvent(fileHash, txHash common.Hash) error {
	if me.baseClient.currentAddress == "" {
		return os.ErrPermission
	}
	curAddr := me.baseClient.currentAddress
	currentAddr := common.HexToAddress(curAddr)
	//get status from cache

	fHash := strings.ToLower(fileHash.Hex())
	evKey := []byte(fHash + "_" + curAddr)
	evd := me.getSigningRequestFromCache(evKey)

	fi, err := me.FileInfo(fileHash, false)
	if err != nil {
		return err
	}
	if fi.Removed {
		if evd != nil {
			me.baseClient.delEvent(fHash, curAddr)
			me.baseClient.notify(evd, evd.TxHash, StatusFail)
		}
		return nil
	}
	//get status from bc
	signers, err := me.fileSigners(fileHash, false)
	if err != nil {
		return err
	}
	for _, a := range signers {
		if bytes.Equal(currentAddr.Bytes(), a.Bytes()) {
			if evd != nil {
				me.baseClient.delEvent(fHash, curAddr)
				me.baseClient.notify(evd, evd.TxHash, StatusFail)
			}
			return nil
		}
	}
	if evd == nil {
		evd = &PendingTx{}
		evd.CurrentAddress = curAddr
		evd.FileHash = fHash
		evd.Type = EventSigningRequest
		evd.TxHash = strings.ToLower(txHash.Hex())
		bts, err := json.Marshal(evd)
		if err == nil {
			me.baseClient.eventsDB.Put(evKey, bts)
		}
	}
	return me.baseClient.notify(evd, evd.TxHash, StatusPending)
}

func (me *fsClient) getSigningRequestUndefinedSignersCacheKey(fileHash string) []byte {
	curAddr := me.baseClient.currentAddress
	if curAddr == "" {
		return nil
	}
	return []byte(fmt.Sprintf("%s_%s_%s", fileHash, curAddr, PendingTypeSignRequest))
}

func (me *fsClient) GetSigningRequestUndefinedSignersCache(fileHash string) *PendingTx {
	curAddr := me.baseClient.currentAddress
	if curAddr == "" {
		return nil
	}
	eventKey := me.getSigningRequestUndefinedSignersCacheKey(fileHash)
	if eventKey == nil {
		return nil
	}
	return me.getSigningRequestFromCache(eventKey)
}

func (me *fsClient) getSigningRequestFromCache(key []byte) *PendingTx {
	b, err := me.baseClient.eventsDB.Get(key)
	if err == nil && len(b) > 0 {
		ptx := me.baseClient.deserializePendingTx(b)
		if ptx != nil {
			log.Printf("[fsClient][getSigningRequestFromCache] txHash: %s for fileHash: %s, fileName: %s", ptx.TxHash, ptx.FileHash, ptx.FileName)
			return ptx
		}
	}
	return nil
}

func (me *fsClient) contractVersion() string {
	ctx, cancel := me.baseClient.ctxWithTimeout()
	opts := &bind.CallOpts{Pending: false, Context: ctx}
	v, err := me.proxeusFSContractCaller.DappVersion(opts)
	cancel()
	if err != nil {
		return "?"
	}
	return string(bytes.Trim(v[:], "\x00"))
}

func (me *fsClient) xesAmountPerFile(prvs []common.Address) (*big.Int, error) {
	ctx, cancel := me.baseClient.ctxWithTimeout()
	opts := &bind.CallOpts{Pending: false, Context: ctx}
	b, err := me.proxeusFSContractCaller.XESAmountPerFile(opts, prvs)
	cancel()
	return b, err
}

func (me *fsClient) FileVerify(fileHash common.Hash, readFromCache bool) (bool, []common.Address, error) {
	if me.baseClient.currentAddress == "" {
		return false, nil, os.ErrPermission
	}
	addr := me.baseClient.currentAddress
	if readFromCache {
		var fv *fileVerify
		err := me.fileVerifyCache.Get(fileHash, &fv)
		if err == nil {
			return fv.valid, fv.signers, nil
		}
	}
	ctx, cancel := me.baseClient.ctxWithTimeout()
	opts := &bind.CallOpts{Pending: false, From: common.HexToAddress(addr), Context: ctx}
	valid, signers, err := me.proxeusFSContractCaller.FileVerify(opts, fileHash)
	cancel()
	if err == nil {
		fv := &fileVerify{valid: valid, signers: signers}
		me.fileVerifyCache.Put(fileHash, fv)
	}
	return valid, signers, err
}

// Returns Service Provider's url if present in smart contract and caches result for next calls
func (me *fsClient) spInfo(strProv common.Address) (string, error) {
	var spUrl string
	err := me.genericCache.Get(strProv.Hex(), &spUrl)
	if err == nil {
		return spUrl, nil
	}
	ctx, cancel := me.baseClient.ctxWithTimeout()
	url, err := me.proxeusFSContractCaller.SpInfo(&bind.CallOpts{Pending: false, Context: ctx}, strProv)
	cancel()
	if err == nil {
		strUrl := strings.TrimRight(string(url[:]), "\x00")
		me.genericCache.Put(strProv.Hex(), strUrl)
		return strUrl, nil
	}
	return "", err
}

func (me *fsClient) spInfoForFile(fileHash string) (string, error) {
	var spUrl string
	err := me.genericCache.Get(fileHash, &spUrl)
	if err == nil {
		return spUrl, nil
	}
	l, err := me.SpList()
	if err != nil {
		return "", err
	}
	fhash := util.StrHexToBytes32(fileHash)
	for _, a := range l {
		ctx, cancel := me.baseClient.ctxWithTimeout()
		yes, err := me.proxeusFSContractCaller.FileHasSP(&bind.CallOpts{Pending: false, Context: ctx}, fhash, a)
		cancel()
		if err == nil && yes {
			url, err := me.spInfo(a)
			if err == nil {
				me.genericCache.Put(fileHash, url)
				return url, nil
			}
		}
	}
	return "", err
}

func (me *fsClient) storageProviders() ([]StorageProvider, error) {
	l, err := me.SpList()
	if err != nil {
		return nil, err
	}
	res := make([]StorageProvider, 0, len(l))
	for _, a := range l {
		url, _ := me.spInfo(a)
		res = append(res, StorageProvider{Address: a.Hex(), URL: url})
	}
	return res, nil
}

func (me *fsClient) SpList() ([]common.Address, error) {
	var spList []common.Address
	err := me.genericCache.Get(splist, &spList)
	if err == nil {
		return spList, nil
	}
	ctx, cancel := me.baseClient.ctxWithTimeout()
	list, err := me.proxeusFSContractCaller.SpList(&bind.CallOpts{Pending: false, Context: ctx})
	cancel()
	if err == nil {
		me.genericCache.Put(splist, list)
		return list, nil
	}
	return nil, err
}

func (me *fsClient) close() {
	if me.fileListCache != nil {
		me.fileListCache.Close()
	}
	if me.fileSignersCache != nil {
		me.fileSignersCache.Close()
	}
	if me.fileInfoCache != nil {
		me.fileInfoCache.Close()
	}
	if me.myFileHashesDB != nil {
		me.myFileHashesDB.Close()
	}
	if me.hasReadRightsCache != nil {
		me.hasReadRightsCache.Close()
	}
	if me.hasWriteRightsCache != nil {
		me.hasWriteRightsCache.Close()
	}
	if me.fileVerifyCache != nil {
		me.fileVerifyCache.Close()
	}
	if me.genericCache != nil {
		me.genericCache.Close()
	}
}
