package core

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/ProxeusApp/storage-app/dapp/core/util"

	"github.com/ProxeusApp/storage-app/dapp/core/embdb"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ProxeusApp/storage-app/spp/client"
	"github.com/ProxeusApp/storage-app/spp/client/models"

	"github.com/ProxeusApp/storage-app/dapp/core/updater"

	cache "github.com/ProxeusApp/memcache"
	"github.com/ethereum/go-ethereum/common"

	"github.com/atotto/clipboard"

	"github.com/ProxeusApp/storage-app/dapp/core/account"
	"github.com/ProxeusApp/storage-app/dapp/core/ethereum"
	"github.com/ProxeusApp/storage-app/dapp/core/file"
	"github.com/ProxeusApp/storage-app/dapp/core/notification"
	"github.com/ProxeusApp/storage-app/spp/config"
	channelhub "github.com/ProxeusApp/storage-app/web"
)

type App struct {
	cfg         *config.Configuration
	wallet      *account.Wallet
	addressBook *account.AddressBook
	ETHClient   *ethereum.DappClient
	ChanHub     *channelhub.ChannelHub
	fileHandler *file.Handler
	stopchan    chan bool
	stopWg      sync.WaitGroup

	notificationManager *notification.Manager
	fileInfoChan        chan *ethereum.FileHashOrder

	pushMsgs     bool
	SharedWithMe bool
	MyFiles      bool
	SignedByMe   bool
	ExpiredFiles bool
	SearchTxt    string

	sessionHandler *cache.Cache

	searchLock      sync.RWMutex
	accountDirLock  sync.Mutex //handle if logging in and user closes the dapp simultaneously
	isLoggedInState bool
	accountDB       *embdb.DB
	accountDBMutex  *sync.Mutex
}

type accountCache struct {
	EthBalance *big.Int
}

type Options struct {
	Share              bool `json:"share"`
	SendSigningRequest bool `json:"sendSigningRequest"`
	Remove             bool `json:"remove"`
	Revoke             bool `json:"revoke"`
}

type FileInfo struct {
	ID                                   common.Hash                 `json:"id"`
	FileType                             *big.Int                    `json:"fileType"`
	Owner                                *account.AddressBookEntry   `json:"owner"`
	Expiry                               *big.Int                    `json:"expiry"`
	GraceSeconds                         int                         `json:"graceSeconds"`
	Expired                              bool                        `json:"expired"`
	InGracePeriod                        bool                        `json:"inGracePeriod"`
	AboutToExpire                        bool                        `json:"aboutToExpire"`
	IsPublic                             bool                        `json:"isPublic"`
	Filename                             string                      `json:"filename"`
	HasThumbnail                         bool                        `json:"hasThumbnail"`
	ReplacesFile                         common.Hash                 `json:"replacesFile"`
	Fparent                              common.Hash                 `json:"fparent"`
	Removed                              bool                        `json:"removed"`
	SignatureStatus                      int                         `json:"signatureStatus"` //1 no signature required, 2 signatures missing, 3 signed
	SCOrder                              int                         `json:"scOrder"`
	UndefinedSigners                     int                         `json:"undefinedSigners"`
	UndefinedSignersLeft                 int                         `json:"undefinedSignersLeft"`
	SentSignRequestsFileUndefinedSigners []*account.AddressBookEntry `json:"sentSignRequestsFileUndefinedSigners"`
	ReadAccess                           []*account.AddressBookEntry `json:"readAccess"`
	DefinedSigners                       []*account.AddressBookEntry `json:"definedSigners"`
	Signers                              []*account.AddressBookEntry `json:"signers"`
}

type EventMsg struct {
	GroupID string      `json:"grpID"`
	Data    interface{} `json:"data"`
	Type    string      `json:"type"`
}

type AccountInfo struct {
	Name       string `json:"name"`
	Address    string `json:"address"`
	ETHBalance string `json:"ethBalance"`
	XESBalance string `json:"balance"`
	Allowance  string `json:"allowance"`
}

type GasEstimate struct {
	GasPrice *big.Int `json:"gasPrice"`
	GasLimit uint64   `json:"gasLimit"`
}

type Quote struct {
	FileHash  string          `json:"fileHash"`
	Providers []QuoteProvider `json:"providers"`
}

type QuoteProvider struct {
	Provider      models.StorageProviderInfo `json:"provider"`
	PriceSize     string                     `json:"priceSize"`
	PriceDuration string                     `json:"priceDuration"`
	PriceTotal    string                     `json:"priceTotal"`
	Available     bool                       `json:"available"`
}

const (
	AccountDBName = "account"
)

var ErrPGPPublicKeyMissing = errors.New("PGP public key missing")

func NewApp(cfg *config.Configuration, chanHub *channelhub.ChannelHub, sessionTimeoutDuration time.Duration) (*App, error) {
	storageDir := cfg.StorageDir
	if storageDir == "" {
		storageDir = "./"
	}

	storageDir = filepath.Join(storageDir, "proxeus")
	log.Println("storageDir set to", storageDir)
	acfg := &account.Config{StorageDir: storageDir, PGPServiceURL: cfg.PGPPublicServiceURL, FileSuffix: ".proxeusks"}
	_, err := os.Stat(cfg.StorageDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(cfg.StorageDir, 0750)
		if err != nil {
			return nil, err
		}
	}
	wallet, err := account.LoadWallet(acfg)
	if err != nil {
		return nil, err
	}

	app := &App{cfg: cfg, ChanHub: chanHub, wallet: wallet, sessionHandler: cache.NewExtendExpiryOnGet(sessionTimeoutDuration, true),
		accountDBMutex: new(sync.Mutex)}
	if app.ChanHub != nil {
		app.ChanHub.ChannelSubscribed = app.subscribedListener
		app.ChanHub.ChannelUnsubscribed = app.unsubscribedListener
	}

	app.ETHClient, err = ethereum.NewDappClient(cfg.EthClientURL, cfg.EthWebSocketURL, cfg.XESContractAddress, cfg.ContractAddress)
	if err != nil {
		return nil, err
	}

	return app, nil
}

func (me *App) pushNotification(n *notification.Notification) {
	err := me.push(EventMsg{Type: "notification", Data: n})
	if err != nil {
		log.Println("push err", err)
	}
}

func (me *App) pushSignRequest(ptx *ethereum.PendingTx, add bool) (EventMsg, error) {
	fhash := util.StrHexToBytes32(ptx.FileHash)
	fi, err := me.ETHClient.FileInfo(fhash, true)
	if err != nil {
		log.Println("fileInfo err", err)
	}

	dat := map[string]interface{}{
		"owner":    fi.Ownr.Hex(),
		"txHash":   ptx.TxHash,
		"fileHash": ptx.FileHash,
	}
	fHash := util.Bytes32ToHexStr(fi.Id)

	spUrl, err := me.ETHClient.SpInfoForFile(fHash)
	if err == nil {
		fileName := me.getFileNameByHash(fHash, true)
		if fileName != "" {
			dat["filename"] = fileName
			dat["hasThumbnail"] = me.fileHandler.HasThumbnail(spUrl, fHash)
		} else {
			dat["filename"] = fHash
			dat["hasThumbnail"] = false
		}
	}
	t := "remove_signing_request"
	if add {
		t = "add_signing_request"
	}
	eventMsg := EventMsg{Type: t, Data: dat}
	err = me.push(eventMsg)

	if err != nil {
		log.Println("push err", err)
	}
	return eventMsg, err
}

func (me *App) getFileNameByHash(fileHash string, fromMetaOnly bool) string {
	spUrl := ""
	if !fromMetaOnly {
		fileMeta, err := me.fileHandler.FileMetaHandler.Get(fileHash)
		if err == nil && fileMeta.SpUrl != "" {
			spUrl = fileMeta.SpUrl
		} else {
			if spUrl, err = me.ETHClient.SpInfoForFile(fileHash); err != nil {
				log.Printf("[core][app][getFileNameByHash] error while getting SP info for file %s: %s", fileHash, err.Error())
				return ""
			}
		}
	}

	return me.fileHandler.FileNameByHash(fileHash, fromMetaOnly, spUrl)
}

func (me *App) toTimestamp(durationDays int) *big.Int {
	timeStamp := time.Now().Add(time.Hour * 24 * time.Duration(durationDays))
	return big.NewInt(int64(timeStamp.Unix()))
}

func (me *App) setupEventWorkers() {
	const maxWorkers = 4
	me.stopchan = make(chan bool, maxWorkers)
	me.fileInfoChan = make(chan *ethereum.FileHashOrder, 600)
	for i := 0; i < maxWorkers; i++ {
		me.eventRoutine()
	}
}

func (me *App) eventRoutine() (err error) {
	me.stopWg.Add(1)
	go func() {
		log.Println("eventRoutine started")
		defer func() {
			me.stopWg.Done()
			log.Println("eventRoutine ended")
		}()
		for {
			select {
			case fh, ok := <-me.fileInfoChan:
				if !ok {
					return
				}
				me.fileInfoUpdate(fh)
			case <-me.stopchan:
				return
			}
		}
	}()
	return
}

func (me *App) getFileExpiryInfo(fileMeta *file.FileMeta) (time.Duration, time.Duration, time.Duration) {
	now := time.Now()
	expiry := time.Unix(fileMeta.Expiry, 0)
	expiryDiff := expiry.Sub(now)
	aboutToExpireThreshold := 10 * 24 * time.Hour // 10 days
	graceDuration := time.Duration(fileMeta.GraceSeconds) * time.Second

	return expiryDiff, aboutToExpireThreshold, graceDuration
}

func (me *App) getStorageProviderInfoByUrl(spUrl string) (models.StorageProviderInfo, error) {
	if len(me.cfg.ForceSpp) > 10 {
		spUrl = me.cfg.ForceSpp
	}

	spInfo, err := client.ProviderInfo(spUrl)
	if err != nil {
		log.Println("[core][app][getStorageProviderInfoByUrl] Couldn't get storage provider info for", spUrl, err)
		return spInfo, err
	}

	return spInfo, err
}

func (me *App) fileInfoUpdate(fileHashOrder *ethereum.FileHashOrder) {
	me.searchLock.RLock()
	grpID := fmt.Sprintf("%v-%v-%v-%v-%v", me.MyFiles, me.SharedWithMe, me.SignedByMe, me.ExpiredFiles, me.SearchTxt)
	me.searchLock.RUnlock()

	activeAccountETHAddress := me.GetActiveAccountETHAddress()
	if activeAccountETHAddress == "" {
		return //no active account
	}

	fi, err := me.ETHClient.FileInfo(fileHashOrder.FileHash, true)
	if err != nil || fi.FileType.Int64() == ethereum.THUMBNAIL || util.Bytes32Empty(fi.Id) {
		return
	}
	iAmTheOwner := strings.ToLower(fi.Ownr.Hex()) == activeAccountETHAddress
	if me.SharedWithMe && iAmTheOwner || me.MyFiles && !iAmTheOwner {
		return
	}

	nfi := &FileInfo{}
	nfi.ID = fi.Id
	fileHash := nfi.ID.Hex()
	if !fi.Removed {
		if !me.fileHandler.FileMatchesQuery(fileHash, me.SearchTxt) {
			return
		}

		fileMetaChanged := false
		fileMeta, err := me.fileHandler.FileMetaHandler.Get(fileHash)
		if err == file.ErrFileMetaNotFound {
			fileMeta = &file.FileMeta{
				FileHash: fileHash,
				FileName: fileHash,
			}
		} else if err != nil {
			return
		}
		if fileMeta.Expiry == 0 {
			fileMeta.Expiry = fi.Expiry.Int64()
			fileMetaChanged = true
		}
		if fileMeta.SpUrl == "" {
			spUrl, err := me.ETHClient.SpInfoForFile(fileHash)
			if err != nil {
				return
			}
			fileMeta.SpUrl = spUrl
			fileMetaChanged = true
		}
		if fileMeta.GraceSeconds == 0 {
			spInfo, err := me.getStorageProviderInfoByUrl(fileMeta.SpUrl)
			if err == nil {
				fileMeta.GraceSeconds = spInfo.GraceSeconds
				fileMetaChanged = true
			}
		}

		expiryDiff, aboutToExpireThreshold, graceDuration := me.getFileExpiryInfo(fileMeta)

		var notifType string

		if expiryDiff < graceDuration*-1 {
			if !fileMeta.Expired {
				notif, err := me.notificationManager.FindByFileHashAndType(fileHash, "signing_request")
				if err == nil && notif != nil {
					if fileMeta.FileName != fileMeta.FileHash && fileMeta.FileName != "" {
						notif.Data["fileName"] = fileMeta.FileName
					}

					if err = me.handleExecutedEventSigningRequest(notif, notif.Data, true, notif.Pending); err != nil {
						log.Printf("[app][fileInfoUpdate] could not handleExecutedEventSigningRequest for %s,  err: %s", fileHash, err.Error())
					}
				}
				fileMeta.Expired = true
				fileMetaChanged = true
			}
			nfi.Expired = true
			notifType = "file_expired"
		} else if expiryDiff < 0 {
			nfi.InGracePeriod = true
			notifType = "file_grace_period"
		} else if expiryDiff < aboutToExpireThreshold {
			nfi.AboutToExpire = true
			notifType = "file_about_to_expire"
		}

		if !fileMeta.Expired {
			nfi.HasThumbnail = fileMeta.HasThumbnail

			if fileMeta.FileName == fileMeta.FileHash || fileMeta.FileName == "" {
				log.Printf("[app][fileInfoUpdate] %s not expired and no fileName in meta, try download from SPP",
					fileHash)
				me.fileHandler.TryDownload(fileMeta.SpUrl, fileMeta.FileHash)
			}
		} else if fileMeta.HasThumbnail {
			//if file is expired but had a thumbnail only show if thumbnail is already saved locally (do not trigger download)
			if thumbPath, err := me.GetThumbnail(fileHash, true); thumbPath != "" && err == nil {
				nfi.HasThumbnail = true
			}
		}

		nfi.GraceSeconds = fileMeta.GraceSeconds
		nfi.Filename = fileMeta.FileName

		if notifType != "" {
			notifData := map[string]interface{}{
				"fileHash":     fileMeta.FileHash,
				"fileName":     fileMeta.FileName,
				"expiry":       fileMeta.Expiry,
				"graceSeconds": fileMeta.GraceSeconds,
			}

			notif, err := me.notificationManager.AddOrUpdate(notifType, map[string]string{"fileHash": fileMeta.FileHash}, notifData)
			if err != nil {
				log.Printf("[core][app][fileInfoUpdate] error while add or update a notification of type %s for fileHash %s\n", notifType, fileMeta.FileHash)
				return
			}
			if !notif.Dismissed {
				me.push(EventMsg{Type: "notification", Data: notif})
			}
		}

		if fileMetaChanged {
			_ = me.fileHandler.FileMetaHandler.Put(fileMeta)
		}
	}

	nfi.SCOrder = fileHashOrder.SCOrder
	nfi.Removed = fi.Removed
	nfi.Expiry = fi.Expiry
	nfi.FileType = fi.FileType
	nfi.Fparent = fi.Fparent
	nfi.ReplacesFile = fi.ReplacesFile
	nfi.IsPublic = fi.IsPublic
	nfi.Owner = me.ensureItExistsInOurAddressBook(fi.Ownr.Hex())
	iAmInReadAccess := me.loopAddrs(fi.ReadAccess, &nfi.ReadAccess)

	if !iAmTheOwner && !iAmInReadAccess && !nfi.Removed {
		// if we had read access before but not anymore we need to make it disappear again
		// currently used for file unshare
		nfi.Removed = true
	}

	var fileMeta *file.FileMeta
	if nfi.Removed {
		//remove it from our disk
		me.fileHandler.RemoveFileAndMetaFromDisk(fileHash)
	} else {
		fileMeta, err = me.fileHandler.FileMetaHandler.Get(fileHash)
		if err == nil {
			// if the user locally "removed" the file, don't show it
			if fileMeta.Hidden {
				nfi.Removed = true
			}
		}
	}

	if !me.ExpiredFiles && nfi.Expired || me.ExpiredFiles && !nfi.Expired {
		return
	}

	me.loopAddrs(fi.DefinedSigners, &nfi.DefinedSigners)
	signers, err := me.ETHClient.FileSigners(fileHashOrder.FileHash, true)
	if err == nil {
		me.loopAddrs(signers, &nfi.Signers)
	}

	iAmSigner := false
	for _, s := range signers {
		if strings.ToLower(s.Hex()) == activeAccountETHAddress {
			iAmSigner = true
		}
	}

	if me.SharedWithMe && iAmSigner || me.SignedByMe && !iAmSigner {
		return
	}

	if fi.FileType.Int64() == ethereum.UNDEFINEDSIGNERS {
		if iAmTheOwner {
			signRequestAddresses := me.ETHClient.GetSentSignRequestAddrForFileWithUndefinedSigners(fileHash)
			me.loopAddrs(signRequestAddresses, &nfi.SentSignRequestsFileUndefinedSigners)
		}
		nfi.UndefinedSigners = len(signers)
		nfi.UndefinedSignersLeft = nfi.UndefinedSigners - len(nfi.Signers)
	}

	//1 no signatories required, 2 signatures missing, 3 signed
	if len(fi.DefinedSigners) == 0 && len(signers) == 0 {
		nfi.SignatureStatus = 1
	} else if len(signers) == len(nfi.Signers) {
		nfi.SignatureStatus = 3
	} else {
		nfi.SignatureStatus = 2
	}

	me.searchLock.RLock()
	doNotProceed := grpID != fmt.Sprintf("%v-%v-%v-%v-%v", me.MyFiles, me.SharedWithMe, me.SignedByMe, me.ExpiredFiles, me.SearchTxt)
	me.searchLock.RUnlock()

	if doNotProceed || me.GetActiveAccountETHAddress() != activeAccountETHAddress {
		return
	}
	if iAmTheOwner && fileMeta != nil && !fileMeta.Uploaded && !fileMeta.Expired {
		log.Printf("[app][fileInfoUpdate] file %s is not shown in list: fileMeta.Uploaded %t | fileMeta.Expired %t",
			fileHash, fileMeta.Uploaded, fileMeta.Expired)
		return
	}

	me.push(EventMsg{Type: "fileInfo", Data: nfi, GroupID: grpID})
}

func (me *App) ListAccounts() []*account.AccFile {
	return me.wallet.All()
}

var BuildVersion string

const (
	ReleaseVersion           = "3.0.0"
	ReleaseUpdateUrl         = "https://github.com/ProxeusApp/dappRelease/raw/master/versions_v1"
	ReleaseUpdateFallbackUrl = "https://proxeus.com/update_fallback_v1"
	userAccountAppDir        = "account"
)

func updateUrls() []string {
	return []string{ReleaseUpdateUrl, ReleaseUpdateFallbackUrl}
}

var ErrEthClientNotInitialized = errors.New("version cannot be determined. eth client has not been initialized")

func (me *App) Versions() (interface{}, error) {
	if me.ETHClient == nil {
		return nil, ErrEthClientNotInitialized
	}
	return updater.Versions(ReleaseVersion, me.ETHClient.ContractVersion(), updateUrls())
}

func (me *App) DownloadUpdate() error {
	return updater.Download(updateUrls())
}

func (me *App) ApplyUpdate() error {
	return updater.Apply()
}

func (me *App) AccountImport(r io.Reader, password string) (acc *account.Account, alreadyExists bool, err error) {
	return me.wallet.Import(r, password)
}

func (me *App) AccountExport() (string, error) {
	return me.wallet.Export()
}

func (me *App) AccountExportByAddress(ethAddr, pw string) (string, error) {
	return me.wallet.ExportAccount(ethAddr, pw)
}

func (me *App) AccountRemove(ethAddr, pw string) error {
	err := me.wallet.Remove(ethAddr, pw)
	if err != nil {
		return err
	}

	accountDir := filepath.Join(me.defaultStorageDir(), ethAddr)
	if err != nil {
		return err
	}

	return os.RemoveAll(accountDir)
}

func (me *App) Login(ethAddr, pw string) error {
	var err error
	if len(pw) == 0 {
		return os.ErrPermission
	}
	if err = me.wallet.Login(ethAddr, pw); err != nil {
		return err
	}
	if err = me.onLogin(); err != nil {
		return err
	}
	me.sessionStart()
	if me.hasNoActiveAccount() {
		return ErrNoActiveAccount
	}
	return err
}

func (me *App) LoginWithNew(name string, pw string) error {
	var err error
	if err = me.wallet.LoginWithNewAccount(name, pw); err != nil {
		return err
	}
	if err = me.onUserLogin(true); err != nil {
		return err
	}
	me.sessionStart()
	if me.hasNoActiveAccount() {
		return ErrNoActiveAccount
	}
	_, err = me.addressBook.QuickInsertByETHAddr(me.wallet.GetActiveAccountName(), me.GetActiveAccountETHAddress(), me.wallet.GetActiveAccountPGPKey())
	return err
}

func (me *App) LoginWithImportedKeystore(ethAddr, password string) error {
	var err error
	if err := me.wallet.Login(ethAddr, password); err != nil {
		return err
	}
	if err = me.onUserLogin(true); err != nil {
		return err
	}
	me.sessionStart()
	if me.hasNoActiveAccount() {
		return ErrNoActiveAccount
	}
	_, err = me.addressBook.QuickInsertByETHAddr(me.wallet.GetActiveAccountName(), me.GetActiveAccountETHAddress(), me.wallet.GetActiveAccountPGPKey())
	return err
}

func (me *App) LoginWithETHPriv(ethPriv, name, pw string) error {
	var err error
	if err = me.wallet.LoginWithETHPriv(ethPriv, name, pw); err != nil {
		return err
	}
	if err = me.onLogin(); err != nil {
		return err
	}
	me.sessionStart()
	if me.hasNoActiveAccount() {
		return ErrNoActiveAccount
	}
	_, err = me.addressBook.QuickInsertByETHAddr(me.wallet.GetActiveAccountName(), me.GetActiveAccountETHAddress(), me.wallet.GetActiveAccountPGPKey())
	return err
}

func (me *App) LoginWithETHPrivAndPGPPriv(ethPriv, name, pw, pgpPriv, pgppw string) error {
	var err error
	if err = me.wallet.LoginWithETHPrivAndPGPPriv(ethPriv, name, pw, pgpPriv, pgppw); err != nil {
		return err
	}
	if err = me.onLogin(); err != nil {
		return err
	}
	me.sessionStart()
	if me.hasNoActiveAccount() {
		return ErrNoActiveAccount
	}
	_, err = me.addressBook.QuickInsertByETHAddr(me.wallet.GetActiveAccountName(), me.GetActiveAccountETHAddress(), me.wallet.GetActiveAccountPGPKey())
	return err
}

func (me *App) onLogin() error {
	return me.onUserLogin(false)
}

func (me *App) onUserLogin(isNew bool) error {
	if me.hasNoActiveAccount() {
		return ErrNoActiveAccount
	}
	me.isLoggedInState = true
	accStorageDir, err := me.storageDirAccount()
	if err != nil {
		return err
	}
	userAccountAppDir, err := me.userAccountAppDir()
	if err != nil {
		return err
	}

	if !isNew {
		err = me.ensureCompatibility(accStorageDir, userAccountAppDir)
		if err != nil {
			return err
		}
		if err := me.decryptUserData(); err != nil {
			log.Println(err) //just log error and continue if no encrypted file found
			if !os.IsNotExist(err) {
				return err
			}
		}
	}

	me.accountDB, err = embdb.Open(userAccountAppDir, AccountDBName)
	if err != nil {
		return err
	}

	addressBookStorageDir := "."
	if userAccountAppDir != "" {
		addressBookStorageDir = userAccountAppDir
	}
	addressBookDB, err := embdb.Open(addressBookStorageDir, account.AddressBookDBName)
	if err != nil {
		return err
	}

	if me.wallet.GetPGPClient() == nil {
		return os.ErrInvalid
	}
	me.addressBook, err = account.NewAddressBook(addressBookDB, me.wallet.GetPGPClient())
	if err != nil {
		return err
	}
	me.fileHandler, err = file.NewHandler(me.cfg, me.wallet, accStorageDir, userAccountAppDir, me.ActiveAccount)
	if err != nil {
		log.Println("[app][onUserLogin] error initializing file handler: " + err.Error())
		return err
	}

	me.setupEventWorkers()
	me.setListeners()
	me.startAccountTicker()

	return nil
}

func (me *App) startAccountTicker() {
	ticker := time.NewTicker(15 * time.Second)
	go func() {
		defer ticker.Stop()
		log.Println("[app][startAccountTicker] started...")
		for {
			select {
			case <-ticker.C:
				me.getAndUpdateAccountEthBalance()
			case <-me.stopchan:
				log.Println("[app][startAccountTicker] stopping...")
				return
			}
		}
	}()
}

func (me *App) HasActiveAndUnlockedAccount() bool {
	return me.wallet.HasActiveAndUnlockedAccount()
}

func (me *App) GetActiveAccountETHAddress() string {
	return me.wallet.GetActiveAccountETHAddress()
}

//ensure backwards compatibility with oder versions. if userAccountAppDir is not found create it and move data into it
func (me *App) ensureCompatibility(storageDirAccount, userAccountAppDir string) error {
	if _, err := os.Stat(userAccountAppDir); !os.IsNotExist(err) {
		return nil //plain account folder already exists
	}
	if _, err := os.Stat(fmt.Sprintf("%s_%s", userAccountAppDir, "locked")); !os.IsNotExist(err) {
		return nil //encrypted account folder already exists
	}

	if err := os.MkdirAll(userAccountAppDir, 0750); err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(storageDirAccount, "address_book")); err == nil {
		if err := os.Rename(filepath.Join(storageDirAccount, "address_book", account.AddressBookDBName),
			filepath.Join(userAccountAppDir, account.AddressBookDBName)); err != nil {
			return err
		}
		if err := os.RemoveAll(filepath.Join(storageDirAccount, "address_book")); err != nil {
			return err
		}
	}
	if _, err := os.Stat(filepath.Join(storageDirAccount, file.PendingDBName)); err == nil {
		if err := os.Rename(filepath.Join(storageDirAccount, file.PendingDBName), filepath.Join(userAccountAppDir, file.PendingDBName)); err != nil {
			return err
		}
	}
	if _, err := os.Stat(filepath.Join(storageDirAccount, notification.NotificationDBName)); err == nil {
		if err := os.Rename(filepath.Join(storageDirAccount, notification.NotificationDBName),
			filepath.Join(userAccountAppDir, notification.NotificationDBName)); err != nil {
			return err
		}
	}
	if _, err := os.Stat(filepath.Join(storageDirAccount, ethereum.TransactionsDBName)); err == nil {
		if err := os.Rename(filepath.Join(storageDirAccount, ethereum.TransactionsDBName), filepath.Join(userAccountAppDir, ethereum.TransactionsDBName)); err != nil {
			return err
		}
	}
	//"filenamemap" is the old directory name
	if _, err := os.Stat(filepath.Join(storageDirAccount, "filenamemap")); err == nil {
		if err := os.Remove(filepath.Join(storageDirAccount, "filenamemap")); err != nil {
			return err
		}
	}
	return nil
}

func (me *App) setListeners() {
	go func() {
		if me.hasNoActiveAccount() {
			log.Println("Error in setListeners. ActiveAccount is not set")
		} else {
			if me.notificationManager != nil {
				me.notificationManager.Close()
			}
			userAccountAppDir, err := me.userAccountAppDir()
			if err != nil {
				log.Println("Error initializing userAccountAppDir" + err.Error())
			} else {
				me.notificationManager, _ = notification.New(userAccountAppDir, me.GetActiveAccountETHAddress())
			}

			me.ETHClient.InitListeners(userAccountAppDir, me.GetActiveAccountETHAddress(), me.ethereumListener, me.ETHClient.DefaultEventsHandler)

			me.fileHandler.SetListener(me.fileListener, me.notificationManager, me.ChanHub)
			me.setupClipboardListener()
		}
	}()
}

func (me *App) setupClipboardListener() {
	go func() {
		currentValue := ""
		for {
			newValue, err := clipboard.ReadAll()
			if err != nil {
				log.Println(err)
			}
			protocol := "proxeus://"
			if newValue != currentValue && strings.HasPrefix(newValue, protocol) {
				currentValue = newValue
				// proxeus://localhost:1323?dropID=cbf5d879-e962-42b8-8e05-6b0ff7846d82&recipients=0x35c2b5bd7b2f8754f121a1e2946b9199ae00163a
				location, dropID, _, err := ParseProxeusProtocol(newValue)
				if err == nil {
					dest := base64.URLEncoding.EncodeToString([]byte(newValue))
					err = me.downloadFromDroparea(filepath.Join(me.cfg.StorageDir, "sharefiles", string(dest)), location+"/api/drop_area/"+dropID)
					if err != nil {
						log.Println(err)
						clipboard.WriteAll("")
						continue
					}
					event := EventMsg{Type: "share_process", Data: map[string]interface{}{
						"link": dest,
					}}
					me.push(event)
				}
				clipboard.WriteAll("")
			}
			time.Sleep(2 * time.Second)
		}
	}()
}

func (me *App) RemoveNotification(id string) error {
	if me.notificationManager != nil {
		return me.notificationManager.MarkAsDismissed(id)
	}
	return os.ErrPermission
}

func (me *App) MarkAllNotificationsAsRead() error {
	if me.notificationManager != nil {
		res, err := me.notificationManager.List()
		if err == nil {
			for idx := range res {
				n := res[idx]
				n.Unread = false
				me.notificationManager.Put(n.ID, *n)
			}
			return nil
		}
	}
	return os.ErrPermission
}

func (me *App) UpdateNotification(n notification.Notification) error {
	if me.notificationManager != nil {
		err := me.notificationManager.Put(n.ID, n)
		return err
	}
	return os.ErrPermission
}

func (me *App) UpdateAccountInfo() error {
	if !me.HasActiveAndUnlockedAccount() {
		return nil
	}
	ai := AccountInfo{}
	addr := me.GetActiveAccountETHAddress()
	if addr == "" {
		return os.ErrInvalid
	}

	ethBalance, err := me.getAndUpdateAccountEthBalance()
	if err == nil {
		ai.ETHBalance = ethBalance.String()
	}

	ai.Address = addr
	xesBalance, err := me.ETHClient.BalanceXESof(addr)
	if err == nil {
		ai.XESBalance = xesBalance.String()
	}
	allowance, err := me.ETHClient.ProxeusFSAllowance(addr)
	if err == nil {
		ai.Allowance = allowance.String()
	}
	return me.push(EventMsg{Type: "account", Data: ai})
}

func (me *App) getAndUpdateAccountEthBalance() (*big.Int, error) {
	if me.hasNoActiveAccountDoNotSignalUserActivity() { //do not call hasNoActiveAccount, else SignalUserActivity will be called recurring by tickerFunc
		return nil, ErrNoActiveAccount
	}
	newEthBalance, err := me.ETHClient.BalanceETHof(me.GetActiveAccountETHAddress())
	if err != nil {
		return newEthBalance, err
	}

	a := new(accountCache)
	me.accountDBMutex.Lock()
	defer me.accountDBMutex.Unlock()
	bts, err := me.accountDB.Get([]byte(me.GetActiveAccountETHAddress()))

	if err == nil && bts != nil {
		if err = json.Unmarshal(bts, &a); err != nil {
			return newEthBalance, err
		}
		if a.EthBalance != nil && newEthBalance.Cmp(a.EthBalance) != 0 {
			//notify if balance changed
			if err = me.notifyNewAccountBalance(a.EthBalance, newEthBalance); err != nil {
				log.Println("[app][getAndUpdateAccountEthBalance] notify err: ", err.Error())
			}
		}
	}

	a.EthBalance = newEthBalance
	if err = me.saveAccount(a); err != nil {
		return newEthBalance, err
	}

	return newEthBalance, nil
}

func (me *App) notifyNewAccountBalance(oldEthBalance, newEthBalance *big.Int) error {
	var (
		notificationName, txType string
		ethAmountChange          *big.Int
	)
	if oldEthBalance.Cmp(newEthBalance) < 0 {
		log.Printf("[app][getAndUpdateAccountEthBalance] increased from: %d to %d, diff: %d ", oldEthBalance, newEthBalance, new(big.Int).Sub(newEthBalance, oldEthBalance))
		notificationName = "eth-increase"
		txType = "tx_eth_increase"
		ethAmountChange = new(big.Int).Sub(newEthBalance, oldEthBalance)
	} else if oldEthBalance.Cmp(newEthBalance) > 0 {
		log.Printf("[app][getAndUpdateAccountEthBalance] decreased from: %d to %d, diff: %d ", oldEthBalance, newEthBalance, new(big.Int).Sub(oldEthBalance, newEthBalance))
		notificationName = "eth-decrease"
		txType = "tx_eth_decrease"
		ethAmountChange = new(big.Int).Sub(oldEthBalance, newEthBalance)
	} else {
		return os.ErrInvalid
	}

	m := map[string]interface{}{
		"status":    ethereum.StatusSuccess,
		"name":      notificationName,
		"ethAmount": ethAmountChange,
	}

	n, err := me.notificationManager.Add(txType, m)
	if err != nil {
		return err
	}
	return me.push(EventMsg{Type: "notification", Data: n})
}

func (me *App) saveAccount(a *accountCache) error {
	accountBytes, err := json.Marshal(a)
	if err != nil {
		return err
	}
	if me.hasNoActiveAccountDoNotSignalUserActivity() { //do not call hasNoActiveAccount, else SignalUserActivity will be called recurring by tickerFunc
		return ErrNoActiveAccount
	}
	return me.accountDB.Put([]byte(me.GetActiveAccountETHAddress()), accountBytes)
}

// Updates a wallet's information. For now only name can be updated
func (me *App) UpdateAccount(accountInfo *AccountInfo) error {
	account := me.wallet.FindAccount(accountInfo.Address)
	if account == nil {
		return errors.New("account not found")
	}
	account.SetName(accountInfo.Name)
	return account.Store()
}

func (me *App) GetThumbnail(fileHash string, fromCacheOnly bool) (string, error) {
	if me.hasNoActiveAccount() {
		return "", os.ErrPermission
	}
	spUrl, err := me.ETHClient.SpInfoForFile(fileHash)
	if err != nil {
		return "", err
	}
	return me.fileHandler.Thumbnail(spUrl, fileHash, fromCacheOnly)
}

func (me *App) GetFile(fileHash string) (string, error) {
	if me.hasNoActiveAccount() {
		return "", os.ErrPermission
	}
	spUrl, err := me.ETHClient.SpInfoForFile(fileHash)
	if err != nil {
		return "", err
	}
	return me.fileHandler.RequestFileFromSpp(spUrl, fileHash)
}

func (me *App) downloadFromDroparea(filepath string, url string) error {
	out, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0660)
	if os.IsNotExist(err) {
		os.MkdirAll(filepath, 0770)
		out, err = os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0660)
	}
	if err != nil {
		return err
	}
	defer out.Close()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth("proxeus", "pr0x3us!")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}
	return nil
}

//return err if the event should be fired again
func (me *App) ethereumListener(tx *ethereum.PendingTx, txHash, status string) error {
	m := map[string]interface{}{
		"status":    status,
		"name":      tx.Type,
		"txHash":    txHash,
		"hash":      tx.FileHash,
		"fileName":  tx.FileName,
		"xesAmount": tx.XesAmount,
		"who":       tx.Who}
	if tx.Tx != nil {
		m["gasPrice"] = tx.Tx.GasPrice()
		m["gasLimit"] = tx.Tx.Gas()
	}

	if tx.Type == ethereum.Event {
		if status == ethereum.StatusSuccess {
			me.pushFileStr(tx.FileHash)
		}
		return nil
	} else if tx.Type == ethereum.ConnStatusNotification {
		if status == string(ethereum.ConnOnline) {
			me.UpdateAccountInfo()
		}
		return me.push(EventMsg{Type: ethereum.ConnStatusNotification, Data: map[string]interface{}{"status": status}})
	} else if tx.Type == ethereum.EventSigningRequest {
		return me.handleEventSigningRequest(tx, txHash, status, m)
	}
	if tx.Type == ethereum.PendingTypeRegister {
		if status != ethereum.StatusPending {
			if status == ethereum.StatusSuccess {
				n, err := me.notificationManager.AddOrUpdate("tx_register", map[string]string{"txHash": txHash}, m)
				if err != nil {
					return err
				}
				if !n.Dismissed {
					me.push(EventMsg{Type: "notification", Data: n})
				}
				//push this here to make sure it gets to the UI before the file events
				pushErr := me.push(EventMsg{Type: "tx", Data: m})
				fileErr := me.fileHandler.Register(txHash, tx.FileHash, true)
				me.UpdateAccountInfo()
				if fileErr != nil {
					log.Printf("fatal error when trying to kick off the file upload -> txHash %s %s, file hash %s, error [%s] \n", txHash, status, tx.FileHash, fileErr.Error())
				}
				if pushErr != nil || fileErr != nil {
					return os.ErrClosed
				}
				return nil
			}

			me.fileHandler.RemoveFileAndMetaFromDisk(tx.FileHash)
		}
	} else if tx.Type == ethereum.PendingTypeSignRequest {
		if status == ethereum.StatusSuccess {
			n, err := me.notificationManager.AddOrUpdate("tx_signRequest", map[string]string{"txHash": txHash}, m)
			if err != nil {
				return err
			}
			if !n.Dismissed {
				me.push(EventMsg{Type: "notification", Data: n})
			}
			me.push(EventMsg{Type: "tx", Data: m})
			me.UpdateAccountInfo()
			return err
		}
	} else if tx.Type == ethereum.PendingTypeShare {
		if status == ethereum.StatusSuccess {
			n, err := me.notificationManager.AddOrUpdate("tx_share", map[string]string{"txHash": txHash}, m)
			if err != nil {
				return err
			}
			if !n.Dismissed {
				me.push(EventMsg{Type: "notification", Data: n})
			}
			me.push(EventMsg{Type: "tx", Data: m})
			err = me.reEncryptFile(tx.FileHash)
			if err != nil {
				log.Println("Re-encryption error ", err)
			}
			me.UpdateAccountInfo()
			return err
		}
	} else if tx.Type == ethereum.PendingTypeRevoke {
		if status == ethereum.StatusSuccess {
			n, err := me.notificationManager.AddOrUpdate("tx_revoke", map[string]string{"txHash": txHash}, m)
			if err != nil {
				return err
			}
			if !n.Dismissed {
				me.push(EventMsg{Type: "notification", Data: n})
			}
			me.pushFileStr(tx.FileHash)
			me.push(EventMsg{Type: "tx", Data: m})
			me.UpdateAccountInfo()
			return me.reEncryptFile(tx.FileHash)
		}
	} else if tx.Type == ethereum.PendingTypeRemove {
		if status == ethereum.StatusSuccess {
			n, err := me.notificationManager.AddOrUpdate("tx_remove", map[string]string{"txHash": txHash}, m)
			if err != nil {
				return err
			}
			if !n.Dismissed {
				me.push(EventMsg{Type: "notification", Data: n})
			}
			me.pushFileStr(tx.FileHash)
			me.fileHandler.RemoveFileAndMetaFromDisk(tx.FileHash)
			me.fileHandler.FileMetaHandler.Del(tx.FileHash)
			//if file is removed and already signed by current account it will no longer be in eventsDB (not pending anymore)
			//so we'll need to handle it here
			signingReqNotification, err := me.notificationManager.FindByFileHashAndType(tx.FileHash, "signing_request")
			if err == nil && signingReqNotification != nil {
				me.handleExecutedEventSigningRequest(signingReqNotification, signingReqNotification.Data,
					true, false)
				if err == nil {
					log.Printf("[app][fileInfoUpdate] successfully handleExecutedEventSigningRequest for %s",
						tx.FileHash)
				} else {
					log.Printf("[app][fileInfoUpdate] could not handleExecutedEventSigningRequest for %s,  err: %s",
						tx.FileHash, err.Error())
				}
			}
		}
	} else if tx.Type == ethereum.PendingTypeSign {
		if status == ethereum.StatusSuccess {
			n, err := me.notificationManager.AddOrUpdate("tx_sign", map[string]string{"txHash": txHash}, m)
			if err != nil {
				return err
			}
			if !n.Dismissed {
				me.push(EventMsg{Type: "notification", Data: n})
			}
			me.pushFileStr(tx.FileHash)
		}
	} else if tx.Type == "xes-approve" {
		if status == ethereum.StatusSuccess {
			err := me.push(EventMsg{Type: "tx", Data: m})
			n, err := me.notificationManager.AddOrUpdate("tx_xes_approval", map[string]string{"txHash": txHash}, m)
			if err != nil {
				return err
			}
			if !n.Dismissed {
				me.push(EventMsg{Type: "notification", Data: n})
			}
		}
	}

	if tx.Type == ethereum.EventNotifySign {
		if status == ethereum.StatusSuccess {
			n, err := me.notificationManager.AddOrUpdate("ev_notifysign", map[string]string{"txHash": txHash}, m)
			if err != nil {
				return err
			}
			if !n.Dismissed {
				me.push(EventMsg{Type: "notification", Data: n})
			}
			return err
		}
	} else if tx.Type == ethereum.EventXesSend {
		err := me.push(EventMsg{Type: "tx", Data: m})

		n, err := me.notificationManager.AddOrUpdate("tx_xes_send", map[string]string{"txHash": txHash}, m)
		if err != nil {
			return err
		}
		if !n.Dismissed {
			me.push(EventMsg{Type: "notification", Data: n})
		}
	} else if tx.Type == ethereum.EventXesReceive {
		err := me.push(EventMsg{Type: "tx", Data: m})

		n, err := me.notificationManager.AddOrUpdate("tx_xes_receive", map[string]string{"txHash": txHash}, m)
		if err != nil {
			return err
		}
		if !n.Dismissed {
			me.push(EventMsg{Type: "notification", Data: n})
		}
	}

	err := me.push(EventMsg{Type: "tx", Data: m})
	me.UpdateAccountInfo()
	return err
}

func (me *App) handleEventSigningRequest(tx *ethereum.PendingTx, txHash, status string, m map[string]interface{}) error {
	if status == ethereum.StatusPending {
		eventMsg, err := me.pushSignRequest(tx, true)
		if err != nil {
			return err
		}
		n, err := me.notificationManager.AddOrUpdateAndAppendEventData("signing_request", txHash, m, eventMsg.Data)
		if err != nil {
			return err
		}
		_ = me.push(EventMsg{Type: "notification", Data: n})
	} else if status == ethereum.StatusFail || status == ethereum.StatusSuccess {
		me.pushSignRequest(tx, false)
		signingNotification, err := me.notificationManager.FindByTxHashAndType(txHash, "signing_request")
		if err != nil {
			return err
		}
		if err = me.handleExecutedEventSigningRequest(signingNotification, m, false, status == ethereum.StatusFail); err != nil {
			return err
		}
	}
	return nil
}

func (me *App) handleExecutedEventSigningRequest(signingNotification *notification.Notification, eventData map[string]interface{}, markFileRemoved bool,
	pushSignRequestRemoveMsg bool) error {

	signingNotification, err := me.notificationManager.MarkPendingAs(signingNotification.ID, false)
	if err != nil {
		return err
	}
	signingNotification, err = me.notificationManager.MarkUnreadAs(signingNotification.ID, false)
	if err != nil {
		return err
	}
	if markFileRemoved {
		signingNotification, err = me.notificationManager.MarkFileRemovedAs(signingNotification.ID, true)
		if err != nil {
			return err
		}
	}
	if err = me.push(EventMsg{Type: "notification", Data: signingNotification}); err != nil {
		log.Println("[app][handleExecutedEventSigningRequest] could not push signing_request, err: ", err.Error())
	}
	if pushSignRequestRemoveMsg {
		me.notificationManager.Append(map[string]interface{}{
			"filename": signingNotification.Data["fileName"],
			"owner":    signingNotification.Data["owner"],
		}, eventData)
		removeSigningNotification, err := me.notificationManager.Add("signing_request_removed", eventData)
		if err != nil {
			return err
		}
		log.Printf("[app][handleExecutedEventSigningRequest] will push removeSigningNotification with Data: %v", removeSigningNotification.Data)
		if err = me.push(EventMsg{Type: "notification", Data: removeSigningNotification}); err != nil {
			log.Println("[app][handleExecutedEventSigningRequest] could not push signing_request_removed, err: ", err.Error())
		}
	}

	return nil
}

func (me *App) fileListener(stype, fhash, spUrl, txHash, status, name string, percentage float32) error {
	if file.StatusDownload == stype && file.StatusSuccess == status {
		signReqNotif, err := me.notificationManager.FindByFileHashAndType(fhash, "signing_request")
		if err == nil && signReqNotif != nil {
			me.updateNotificationFileName(signReqNotif)
		}
		me.pushFileStr(fhash)
	} else if file.StatusUpload == stype && file.StatusSuccess == status {
		me.RemoveFileFromDiskKeepMeta(fhash)

		fileMeta, err := me.fileHandler.FileMetaHandler.Get(fhash)
		if err != nil {
			return err
		}
		fileMeta.Uploaded = true
		_ = me.fileHandler.FileMetaHandler.Put(fileMeta)
		log.Printf("[app][fileListener] got file uploaded notification for file: %s, will pushFileStr(fileInfoUpdate)", fhash)
		me.pushFileStr(fhash)
	}
	ev := EventMsg{Data: map[string]interface{}{"status": status, "spUrl": spUrl, "name": stype, "fileName": name, "txHash": txHash, "hash": fhash, "percentage": percentage}}
	if stype == file.StatusUpload {
		ev.Type = "fileUpload"
	} else {
		ev.Type = "fileDownload"
	}
	return me.push(ev)
}

func (me *App) updateNotificationFileName(notification *notification.Notification) {
	fileHash, ok := notification.Data["hash"].(string)
	if !ok {
		return
	}
	if notification.Data["fileName"] != "" && notification.Data["fileName"] != fileHash {
		return
	}
	fileMeta, err := me.fileHandler.FileMetaHandler.Get(fileHash)
	if err != nil || fileMeta.FileName == "" || fileMeta.FileName == fileHash {
		return
	}
	notification.Data["fileName"] = fileMeta.FileName
	notification, err = me.notificationManager.UpdateData(notification.ID, notification.Data)
	if err != nil {
		return
	}
	log.Println("[app][updateNotificationFileName] will update notification with Filename for file: ", fileHash)
	me.push(EventMsg{Type: "notification", Data: notification})
}

func (me *App) subscribedListener(channel *channelhub.Channel, client *channelhub.Client) {
	me.pushMsgs = true

	me.searchLock.Lock()

	me.SharedWithMe = false
	me.MyFiles = false
	me.SignedByMe = false
	me.ExpiredFiles = false
	me.SearchTxt = "" //because frontend cleans it as well

	me.searchLock.Unlock()

	me.fileList()

	if me.notificationManager != nil {
		nList, _ := me.notificationManager.List()
		for _, n := range nList {
			me.push(EventMsg{Type: "notification", Data: n})
		}
	}

	me.fileHandler.NotifyLastState()
	me.ETHClient.NotifyLastState()
	me.UpdateAccountInfo()
}

func (me *App) unsubscribedListener(channel *channelhub.Channel, client *channelhub.Client) {
	me.pushMsgs = false
}

func (me *App) reEncryptFile(fhash string) error {
	if me.hasNoActiveAccount() {
		return os.ErrPermission
	}
	spUrl, err := me.ETHClient.SpInfoForFile(fhash)
	if err != nil {
		return err
	}
	pgpPublicKeys, err := me.collectAllPublicKeysFor(me.ActiveAccount(), fhash)
	if err != nil {
		return err
	}
	_, err = me.fileHandler.ReEncryptFile(spUrl, fhash, pgpPublicKeys)
	_ = me.RemovePlain(fhash)

	return err
}

func (me *App) push(msg EventMsg) error {
	if me.ChanHub != nil && me.pushMsgs {
		return me.ChanHub.Broadcast("global", msg)
	}
	return os.ErrClosed
}

func (me *App) BalanceXES() (*big.Int, error) {
	if me.hasNoActiveAccount() {
		return nil, os.ErrPermission
	}
	return me.ETHClient.BalanceXESof(me.GetActiveAccountETHAddress())
}

func (me *App) ApproveXESToContractEstimateGas(xesValue string) (GasEstimate, error) {
	var gasEstimate GasEstimate

	if me.hasNoActiveAccount() {
		return gasEstimate, os.ErrPermission
	}

	xesValueInt, ok := new(big.Int).SetString(xesValue, 10)
	if ok {
		opts, err := me.ETHClient.XESApproveToProxeusFSEstimateGas(me.wallet.GetActiveAccountETHPrivateKey(), xesValueInt)
		if err != nil {
			return gasEstimate, err
		}

		gasEstimate.GasPrice = opts.GasPrice
		gasEstimate.GasLimit = opts.GasLimit

		return gasEstimate, err
	}

	return gasEstimate, os.ErrInvalid
}

func (me *App) ApproveXESToContract(xesValue string) error {
	if me.hasNoActiveAccount() {
		return os.ErrPermission
	}
	xesValueInt, ok := new(big.Int).SetString(xesValue, 10)
	if ok {
		_, err := me.ETHClient.XESApproveToProxeusFS(me.wallet.GetActiveAccountETHPrivateKey(), xesValueInt)
		return err
	}
	return os.ErrInvalid
}

var ErrInvalidEthAddr = errors.New("Invalid eth address")

func (me *App) SendXESEstimateGas(ethAddressTo, xesAmount string) (GasEstimate, error) {
	var gasEstimate GasEstimate

	if me.hasNoActiveAccount() {
		return gasEstimate, os.ErrPermission
	}

	if !common.IsHexAddress(ethAddressTo) {
		return gasEstimate, ErrInvalidEthAddr
	}

	xesAmountInt, ok := new(big.Int).SetString(xesAmount, 10)
	if ok {
		opts, err := me.ETHClient.XESTransferEstimateGas(me.wallet.GetActiveAccountETHPrivateKey(), ethAddressTo, xesAmountInt)
		if err != nil {
			return gasEstimate, err
		}

		gasEstimate.GasPrice = opts.GasPrice
		gasEstimate.GasLimit = opts.GasLimit

		return gasEstimate, err
	}

	return gasEstimate, os.ErrInvalid
}

func (me *App) SendXES(ethAddressTo, xesAmount string) error {
	if me.hasNoActiveAccount() {
		return os.ErrPermission
	}

	if !common.IsHexAddress(ethAddressTo) {
		return ErrInvalidEthAddr
	}

	xesAmountInt, ok := new(big.Int).SetString(xesAmount, 10)
	if ok {
		_, err := me.ETHClient.XESTransfer(me.wallet.GetActiveAccountETHPrivateKey(), ethAddressTo, xesAmountInt)
		return err
	}
	return os.ErrInvalid
}

func (me *App) SendETHEstimateGas(ethAddressTo, ethAmount string) (GasEstimate, error) {
	var gasEstimate GasEstimate

	if me.hasNoActiveAccount() {
		return gasEstimate, os.ErrPermission
	}

	if !common.IsHexAddress(ethAddressTo) {
		return gasEstimate, ErrInvalidEthAddr
	}

	_, ok := new(big.Int).SetString(ethAmount, 10)
	if ok {
		gasPrice, gasLimit, err := me.ETHClient.ETHTransferEstimateGas()
		if err != nil {
			return gasEstimate, err
		}

		gasEstimate.GasPrice = gasPrice
		gasEstimate.GasLimit = gasLimit

		return gasEstimate, err
	}

	return gasEstimate, os.ErrInvalid
}

func (me *App) SendETH(ethAddressTo, ethAmount string) error {
	if me.hasNoActiveAccount() {
		return os.ErrPermission
	}

	if !common.IsHexAddress(ethAddressTo) {
		return ErrInvalidEthAddr
	}

	ethAmountInt, ok := new(big.Int).SetString(ethAmount, 10)
	if ok {
		_, err := me.ETHClient.ETHTransfer(me.wallet.GetActiveAccountETHPrivateKey(), ethAddressTo, ethAmountInt)
		return err
	}
	return os.ErrInvalid
}

func (me *App) collectAllPublicKeysFor(account *account.Account, fhash string) ([][]byte, error) {
	if account == nil {
		return nil, os.ErrInvalid
	}
	//owners public key
	pubKeys := [][]byte{[]byte(account.GetPGPPublicKey())}
	ownrAddr := account.GetETHAddress()
	fhash32 := util.StrHexToBytes32(fhash)
	fi, err := me.ETHClient.FileInfo(fhash32, false)
	if err != nil {
		return nil, err
	}
	alreadyCollected := map[string]bool{ownrAddr: true}
	for _, addr := range fi.DefinedSigners {
		strAddr := strings.ToLower(addr.Hex())
		if me.addressBook.IsEmptyAddr(strAddr) || alreadyCollected[strAddr] {
			continue
		}
		abe := me.addressBook.Get(strAddr)
		if abe == nil || abe.PGPPublicKey == "" {
			return nil, ErrPGPPublicKeyMissing
		}
		alreadyCollected[strAddr] = true
		pubKeys = append(pubKeys, []byte(abe.PGPPublicKey))
	}

	for _, addr := range fi.ReadAccess {
		strAddr := strings.ToLower(addr.Hex())
		if me.addressBook.IsEmptyAddr(strAddr) || alreadyCollected[strAddr] {
			continue
		}
		abe := me.addressBook.Get(strAddr)
		if abe == nil || abe.PGPPublicKey == "" {
			return nil, ErrPGPPublicKeyMissing
		}
		alreadyCollected[strAddr] = true
		pubKeys = append(pubKeys, []byte(abe.PGPPublicKey))
	}

	return pubKeys, nil
}

// Retrieve storage provider by eth address by checking with the smart contract first,
// then loads the data from the SPP and merges all together
func (me *App) GetStorageProvider(providerAddress string) (models.StorageProviderInfo, error) {
	var providerInfo models.StorageProviderInfo
	// Check whether we have the SPP in our smart contract
	var (
		spUrl string
		err   error
	)
	if len(me.cfg.ForceSpp) > 0 {
		spUrl = me.cfg.ForceSpp
	} else {
		spUrl, err = me.ETHClient.SpInfo(common.HexToAddress(providerAddress))
	}

	if err != nil {
		return providerInfo, err
	}
	// SPP's urls exists in smart contract, retrieve basic info from them
	providerInfo, err = client.ProviderInfo(spUrl)
	// Enrich with ETH data which is trusted
	providerInfo.Address = providerAddress

	return providerInfo, err
}

func (me *App) GetStorageProviders() ([]models.StorageProviderInfo, error) {
	if me.hasNoActiveAccount() {
		return nil, os.ErrPermission
	}
	// Load storage providers from smart contract
	var (
		err                 error
		ethStorageProviders []ethereum.StorageProvider
	)
	if me.cfg.ForceSpp != "" {
		ethStorageProviders = []ethereum.StorageProvider{{Address: me.cfg.StorageProviderAddress, URL: me.cfg.ForceSpp}}
	} else {
		ethStorageProviders, err = me.ETHClient.StorageProviders()
	}
	if err != nil {
		return nil, err
	}
	var spis []models.StorageProviderInfo
	var wg sync.WaitGroup
	var mux sync.Mutex
	for _, ethStorageProvider := range ethStorageProviders {
		// Asynchronally retrieve info from each storage provider
		wg.Add(1)
		go func(ethStorageProvider ethereum.StorageProvider) {
			defer wg.Done()
			clientInfo, err := me.GetStorageProvider(ethStorageProvider.Address)
			if err != nil {
				log.Printf("[app][GetStorageProviders] Can't get info from storage provider, address: %s url: %s, err: %s", ethStorageProvider.Address, ethStorageProvider.URL, err.Error())
			} else {
				mux.Lock()
				spis = append(spis, clientInfo)
				mux.Unlock()
			}
			log.Printf("[app][GetStorageProviders] Provider: %v", ethStorageProvider)
		}(ethStorageProvider)
	}
	wg.Wait()

	return spis, nil
}

func (me *App) Contacts() ([]account.AddressBookEntry, error) {
	if me.hasNoActiveAccount() {
		return nil, os.ErrPermission
	}
	return me.addressBook.List(me.GetActiveAccountETHAddress())
}

func (me *App) ContactCreate(name, ethAddr, pgpPublicKey string) (*account.AddressBookEntry, error) {
	if me.hasNoActiveAccount() {
		return nil, os.ErrPermission
	}
	abe, err := me.addressBook.Stored(ethAddr)
	if err != nil {
		return abe, err
	}
	if abe != nil && abe.Hidden == true {
		// Behind the scenes we don't really delete the AddressBookEntry, we set it to Hidden = true so the
		// frontend tries to create it. In that case we go through the Update function
		return me.addressBook.Update(name, ethAddr, pgpPublicKey)
	}
	return me.addressBook.Create(name, ethAddr)
}

func (me *App) ContactUpdate(name, ethAddr, pgpPublicKey string) (*account.AddressBookEntry, error) {
	if me.hasNoActiveAccount() {
		return nil, os.ErrPermission
	}
	return me.addressBook.Update(name, ethAddr, pgpPublicKey)
}

func (me *App) ContactRemove(ethAddr string) error {
	if me.hasNoActiveAccount() {
		return os.ErrPermission
	}
	return me.addressBook.Hide(ethAddr)
}

func (me *App) ContactFind(ethAddr string) *account.AddressBookEntry {
	if me.hasNoActiveAccount() {
		return nil
	}
	return me.addressBook.Get(ethAddr)
}

var ErrFileSizeLimit = errors.New("file exceeds provider file size limit")

// Preparation before archiving file based on the input.
// Returns Public keys
func (me *App) checkFileSizeAndCollectPGPKeys(register file.Register, definedSigners []account.AddressBookEntry, spInfo models.StorageProviderInfo) ([][]byte, error) {
	var err error
	if err = me.fileSizeCheck(register, spInfo); err != nil {
		return nil, err
	}

	return me.collectPGPKeys(definedSigners)
}

func (me *App) fileSizeCheck(register file.Register, spInfo models.StorageProviderInfo) error {
	// != 0 check because of non-responding provider
	if spInfo.MaxFileSizeByte != 0 && register.FileSize > spInfo.MaxFileSizeByte {
		return ErrFileSizeLimit // We can stop this before calling the blockchain
	}
	return nil
}

func (me *App) collectPGPKeys(definedSigners []account.AddressBookEntry) ([][]byte, error) {
	if me.hasNoActiveAccount() {
		return nil, os.ErrPermission
	}

	//owners public key
	pubKeys := [][]byte{[]byte(me.wallet.GetActiveAccountPGPKey())}

	//readers public key
	for _, entry := range definedSigners {
		abe := me.addressBook.Get(entry.ETHAddress)
		if abe != nil && abe.ETHAddress != "" && abe.ETHAddress == me.GetActiveAccountETHAddress() {
			continue
		}
		if abe == nil || abe.PGPPublicKey == "" {
			return nil, ErrPGPPublicKeyMissing
		}
		pubKeys = append(pubKeys, []byte(abe.PGPPublicKey))
	}
	return pubKeys, nil
}

// Archive files and return information about it, but remove the file! This should only be used for simulations like quote requests
func (me *App) ArchiveFile(register file.Register, definedSigners []account.AddressBookEntry, undefinedSignersCount int64, spInfo models.StorageProviderInfo) (encryptedArchive file.EncryptedArchive, err error) {
	pubKeys, err := me.checkFileSizeAndCollectPGPKeys(register, definedSigners, spInfo)
	if err != nil {
		log.Print("[app][ArchiveFile] error: ", err)
		return encryptedArchive, err
	}

	return me.fileHandler.PrepareRegister(register, pubKeys)
}

var ErrorEstimateGasNotImplemented = errors.New("estimate gas not implemented yet")

func (me *App) RegisterFileEstimateGas(reg file.Register, definedSigners []account.AddressBookEntry, undefinedSignersCount int64,
	spInfo models.StorageProviderInfo) (GasEstimate, error) {

	var gasEstimate GasEstimate

	if me.hasNoActiveAccount() {
		return gasEstimate, os.ErrPermission
	}

	encryptedArchiveInfo, err := me.ArchiveFile(reg, definedSigners, undefinedSignersCount, spInfo)
	if err != nil {
		return gasEstimate, err
	}

	fileHash := util.StrHexToBytes32(encryptedArchiveInfo.FileHash)
	replacesFileHash := util.StrHexToBytes32("")
	storageProviders := []common.Address{common.HexToAddress(spInfo.Address)}

	xesAmount, err := spInfo.TotalPriceForFile(reg.DurationDays, big.NewInt(encryptedArchiveInfo.Size))
	if err != nil {
		return gasEstimate, err
	}

	opts, err := &bind.TransactOpts{}, nil

	if reg.FileKind == 2 {
		// CreateFileSharedEstimateGas not implemented yet
		return gasEstimate, ErrorEstimateGasNotImplemented
	}

	if definedSigners != nil && len(definedSigners) > 0 {
		dsignrs := make([]common.Address, 0, len(definedSigners))
		for _, a := range definedSigners {
			dsignrs = append(dsignrs, common.HexToAddress(a.ETHAddress))
		}

		opts, err = me.ETHClient.CreateFileDefinedSignersEstimateGas(me.wallet.GetActiveAccountETHPrivateKey(), fileHash, reg.FileName, dsignrs,
			me.toTimestamp(reg.DurationDays), replacesFileHash, storageProviders, xesAmount)
	} else {
		opts, err = me.ETHClient.CreateFileUndefinedSignersEstimateGas(me.wallet.GetActiveAccountETHPrivateKey(), fileHash, reg.FileName,
			big.NewInt(undefinedSignersCount), me.toTimestamp(reg.DurationDays), replacesFileHash, storageProviders, xesAmount)
	}
	if err != nil {
		return gasEstimate, err
	}

	gasEstimate.GasPrice = opts.GasPrice
	gasEstimate.GasLimit = opts.GasLimit

	return gasEstimate, err
}

func (me *App) ArchiveFileAndRegister(reg file.Register, definedSigners []account.AddressBookEntry, undefinedSignersCount int64, spInfo models.StorageProviderInfo, readers []string) error {
	pubKeys, err := me.checkFileSizeAndCollectPGPKeys(reg, definedSigners, spInfo)
	spUrl, err := me.ETHClient.SpInfo(common.HexToAddress(spInfo.Address))
	if err != nil {
		return err
	}
	encryptedArchive, err := me.fileHandler.PrepareRegisterAndScheduleUpload(reg, pubKeys, spUrl)
	if err != nil {
		return err
	}

	xesAmount, err := spInfo.TotalPriceForFile(reg.DurationDays, big.NewInt(encryptedArchive.Size))
	if err != nil {
		return err
	}

	txHash, err := "", nil

	if reg.FileKind == 2 {
		txHash, err = me.registerFileShared(encryptedArchive.FileHash, reg.FileName, undefinedSignersCount, me.toTimestamp(reg.DurationDays), spInfo.Address, xesAmount, readers)
	} else {
		if definedSigners != nil && len(definedSigners) > 0 {
			txHash, err = me.registerFileWithDefinedSigners(encryptedArchive.FileHash, reg.FileName, definedSigners,
				me.toTimestamp(reg.DurationDays), spInfo.Address, xesAmount)
		} else {
			txHash, err = me.registerFileWithUndefinedSigners(encryptedArchive.FileHash, reg.FileName,
				undefinedSignersCount, me.toTimestamp(reg.DurationDays), spInfo.Address, xesAmount)
		}
	}

	if err != nil {
		log.Println("[app][ArchiveFileAndRegister] error while register file", err)
		me.fileHandler.RemoveFileAndMetaFromDisk(encryptedArchive.FileHash)
		return err
	}
	return me.fileHandler.Register(txHash, encryptedArchive.FileHash, false)
}

func (me *App) registerFileWithDefinedSigners(fileHash string, filename string,
	definedSigners []account.AddressBookEntry, expiry *big.Int, ethAddrSp string, xesAmount *big.Int) (string, error) {

	if me.hasNoActiveAccount() {
		return "", os.ErrPermission
	}
	fhash := util.StrHexToBytes32(fileHash)
	dsignrs := make([]common.Address, 0, len(definedSigners))
	for _, a := range definedSigners {
		dsignrs = append(dsignrs, common.HexToAddress(a.ETHAddress))
	}

	tx, err := me.ETHClient.CreateFileDefinedSigners(me.wallet.GetActiveAccountETHPrivateKey(), fhash, filename, dsignrs,
		expiry, util.StrHexToBytes32(""), []common.Address{common.HexToAddress(ethAddrSp)}, xesAmount)
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), err
}

func (me *App) registerFileShared(fileHash string, filename string, mandatorySigners int64, expiry *big.Int, ethAddrSp string,
	xesAmount *big.Int, readers []string) (string, error) {
	if me.hasNoActiveAccount() {
		return "", os.ErrPermission
	}
	fhash := util.StrHexToBytes32(fileHash)
	replacesFileHash := util.StrHexToBytes32("")
	readersAddrs := make([]common.Address, 0, len(readers))
	for _, r := range readers {
		readersAddrs = append(readersAddrs, common.HexToAddress(r))
	}
	tx, err := me.ETHClient.CreateFileShared(me.wallet.GetActiveAccountETHPrivateKey(), fhash, filename,
		big.NewInt(mandatorySigners), expiry, replacesFileHash, []common.Address{common.HexToAddress(ethAddrSp)}, readersAddrs, xesAmount)

	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), err
}

func (me *App) registerFileWithUndefinedSigners(fileHash string, filename string, mandatorySigners int64, expiry *big.Int,
	ethAddrSp string, xesAmount *big.Int) (string, error) {
	if me.hasNoActiveAccount() {
		return "", os.ErrPermission
	}
	fhash := util.StrHexToBytes32(fileHash)
	replacesFileHash := util.StrHexToBytes32("")
	tx, err := me.ETHClient.CreateFileUndefinedSigners(me.wallet.GetActiveAccountETHPrivateKey(), fhash, filename,
		big.NewInt(mandatorySigners), expiry, replacesFileHash, []common.Address{common.HexToAddress(ethAddrSp)}, xesAmount)

	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), err
}

func (me *App) sendSigningRequestFilePrepare(fileHash string, ethAddrs []string) ([32]byte, []common.Address, error) {
	fhash := util.StrHexToBytes32(fileHash)

	if me.hasNoActiveAccount() {
		return fhash, nil, os.ErrPermission
	}

	//check if we have signed it already
	signers, err := me.ETHClient.FileSigners(fhash, false)
	if err != nil {
		return fhash, nil, err
	}

	addrs, err := me.filterNewAddrsOnly(fhash, signers, ethAddrs)
	if err != nil {
		return fhash, nil, err
	}

	return fhash, addrs, err
}

func (me *App) SendSigningRequestFileEstimateGas(fileHash string, ethAddrs []string) (GasEstimate, error) {
	gasEstimate := GasEstimate{big.NewInt(0), uint64(0)}

	if me.hasNoActiveAccount() || len(ethAddrs) == 0 {
		return gasEstimate, ErrNoActiveAccount
	}

	ethAddrsWithoutMe := me.removeMyAddress(ethAddrs) //no need to share with owner

	fhash, addrs, err := me.sendSigningRequestFilePrepare(fileHash, ethAddrs)
	if err != nil {
		return gasEstimate, err
	}

	//if only sharing with own account, no need to call file share
	shareEstimate := GasEstimate{big.NewInt(0), uint64(0)}
	if len(ethAddrsWithoutMe) > 0 {
		shareEstimate, err = me.ShareFileEstimateGas(fileHash, ethAddrsWithoutMe)
		if err != ErrEmpty && err != nil {
			return gasEstimate, err
		}
	}

	fileSignOpts, err := me.ETHClient.FileRequestSignEstimateGas(me.wallet.GetActiveAccountETHPrivateKey(), fhash, addrs)
	if err != nil {
		return gasEstimate, err
	}

	//show higher gasPrice
	if fileSignOpts.GasPrice.Cmp(shareEstimate.GasPrice) == 1 {
		gasEstimate.GasPrice = fileSignOpts.GasPrice
	} else {
		gasEstimate.GasPrice = shareEstimate.GasPrice
	}

	sumGasLimit := fileSignOpts.GasLimit + shareEstimate.GasLimit
	if shareEstimate.GasLimit < 0 || fileSignOpts.GasLimit < 0 ||
		sumGasLimit < fileSignOpts.GasLimit || sumGasLimit < shareEstimate.GasLimit {
		return gasEstimate, os.ErrInvalid
	}

	gasEstimate.GasLimit = fileSignOpts.GasLimit + shareEstimate.GasLimit

	return gasEstimate, err
}

func (me *App) removeMyAddress(ethAddrs []string) []string {
	for i, ethAddr := range ethAddrs {
		if ethAddr == me.GetActiveAccountETHAddress() && ethAddr != "" {
			ethAddrs = append(ethAddrs[:i], ethAddrs[i+1:]...)
		}
	}
	return ethAddrs
}

func (me *App) SendSigningRequestFile(fileHash string, ethAddrs []string) (string, error) {
	if me.hasNoActiveAccount() {
		return "", ErrNoActiveAccount
	}
	fhash, addrs, err := me.sendSigningRequestFilePrepare(fileHash, ethAddrs)
	if err != nil {
		return "", err
	}

	ethAddrsWithoutMe := me.removeMyAddress(ethAddrs) //no need to share with owner

	//if only sharing with own account, no need to call file share
	if len(ethAddrsWithoutMe) > 0 {
		_, err = me.ShareFile(fileHash, ethAddrsWithoutMe)
		if err != ErrEmpty && err != nil {
			return "", err
		}
	}

	filename := me.getFileNameByHash(fileHash, false)

	// create a signing request
	tx, err := me.ETHClient.FileRequestSign(me.wallet.GetActiveAccountETHPrivateKey(), fhash, filename, addrs)
	if err != nil {
		return "", err
	}

	return tx.Hash().Hex(), err
}

var ErrEmpty = errors.New("no new addresses provided")

func (me *App) filterNewAddrsOnly(fhash common.Hash, bcAddrs []common.Address, ethAddrs []string) ([]common.Address, error) {
	addrs := me.ethStrAddrsToCommonAddr(ethAddrs)
	for _, bcAddr := range bcAddrs {
		for i, localAddr := range addrs {
			if bytes.Equal(bcAddr.Bytes(), localAddr.Bytes()) {
				addrs = append(addrs[:i], addrs[i+1:]...)
				break
			}
		}
	}
	if len(addrs) == 0 {
		return nil, ErrEmpty
	}
	return addrs, nil
}

func (me *App) shareFilePrepare(fileHash string, ethAddrs []string) ([32]byte, []common.Address, error) {
	fhash := util.StrHexToBytes32(fileHash)

	if me.hasNoActiveAccount() {
		return fhash, nil, os.ErrPermission
	}

	//ensure we can collect the needed public keys upfront
	_, err := me.collectAllPublicKeysFor(me.ActiveAccount(), fileHash)
	if err != nil {
		return fhash, nil, err
	}
	for _, ethAddr := range ethAddrs {
		//check the ethAddr we want to share it with
		abe := me.addressBook.Get(ethAddr)
		if abe == nil || abe.PGPPublicKey == "" {
			return fhash, nil, ErrPGPPublicKeyMissing
		}
	}

	fi, err := me.ETHClient.FileInfo(fhash, false)
	if err != nil {
		return fhash, nil, err
	}

	addrs, err := me.filterNewAddrsOnly(fhash, fi.ReadAccess, ethAddrs)
	if err != nil {
		return fhash, nil, err
	}

	return fhash, addrs, err
}

func (me *App) ShareFileEstimateGas(fileHash string, ethAddrs []string) (GasEstimate, error) {
	gasEstimate := GasEstimate{big.NewInt(0), uint64(0)}

	if me.hasNoActiveAccount() {
		return gasEstimate, ErrNoActiveAccount
	}
	fhash, addrs, err := me.shareFilePrepare(fileHash, ethAddrs)
	if err != nil {
		return gasEstimate, err
	}

	opts, err := me.ETHClient.FileSetPermEstimateGas(me.wallet.GetActiveAccountETHPrivateKey(), fhash, addrs)
	if err != nil {
		return gasEstimate, err
	}

	gasEstimate.GasPrice = opts.GasPrice
	gasEstimate.GasLimit = opts.GasLimit

	return gasEstimate, err
}

func (me *App) ShareFile(fileHash string, ethAddrs []string) (string, error) {
	if me.hasNoActiveAccount() {
		return "", ErrNoActiveAccount
	}
	fhash, addrs, err := me.shareFilePrepare(fileHash, ethAddrs)
	if err != nil {
		return "", err
	}

	filename := me.getFileNameByHash(fileHash, false)

	tx, err := me.ETHClient.FileSetPerm(me.wallet.GetActiveAccountETHPrivateKey(), fhash, filename, addrs)
	if err != nil {
		return "", err
	}

	return tx.Hash().Hex(), nil
}

func (me *App) ethStrAddrsToCommonAddr(addrs []string) []common.Address {
	cAddrs := make([]common.Address, 0)
	//prevent from duplicates
	addrMap := make(map[string]string)
	for _, a := range addrs {
		addrMap[strings.ToLower(a)] = a
	}
	i := 0
	for k := range addrMap {
		cAddrs = append(cAddrs, common.HexToAddress(k))
		i++
	}
	return cAddrs
}

func (me *App) ListFiles(query string, myFiles, sharedWithMe, signedByMe, expiredFiles bool) error {
	if me.hasNoActiveAccount() {
		return os.ErrPermission
	}
	me.searchLock.Lock()

	me.SharedWithMe = sharedWithMe
	me.MyFiles = myFiles
	me.SignedByMe = signedByMe
	me.ExpiredFiles = expiredFiles
	me.SearchTxt = query

	me.searchLock.Unlock()

	err := me.fileList()

	return err
}

func (me *App) fileList() error {
	files, err := me.ETHClient.FileList(true)
	if err != nil {
		return err
	}
	for _, v := range files {
		if me.hasNoActiveAccount() {
			break
		}
		me.pushFileHashOrder(v)
	}
	return nil
}

//Parses location (dev.proxeus.com), dropID (xyz) and reciepient (0x123) from
// proxeus://dev.proxeus.com/shareprocess?dropID=xyz&recipients=0x123
func ParseProxeusProtocol(input string) (location string, dropID string, recipients []string, err error) {
	protocol := "proxeus://"
	if !strings.HasPrefix(input, protocol) {
		decoded, err := base64.URLEncoding.DecodeString(input)
		if err != nil {
			return "", "", nil, err
		}
		input = string(decoded)
	}
	if !strings.HasPrefix(input, protocol) {
		return "", "", nil, ErrEmpty
	}
	url, err := url.Parse(input)
	if err != nil {
		return "", "", nil, nil
	}
	rs := strings.Split(url.Query().Get("recipients"), ",")
	return config.Config.MainHostedURL, url.Query().Get("dropID"), rs, nil
}

func (me *App) DropFile(location string, filePath string) (string, error) {
	if me.hasNoActiveAccount() {
		return "", os.ErrPermission
	}
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	buf := bufio.NewReader(file)

	req, err := http.NewRequest("POST", location+"/api/drop_area", buf)
	if err != nil {
		return "", err
	}
	req.SetBasicAuth("proxeus", "pr0x3us!")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}

	if resp.StatusCode == http.StatusOK {
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		var respjson map[string]interface{}
		err = json.Unmarshal(bodyBytes, &respjson)
		if err != nil {
			return "", err
		}
		dropId, ok := respjson["dropID"].(string)
		if !ok {
			return "", err
		}
		return dropId, nil
	}
	return "", err
}

func (me *App) revokeFilePrepare(fileHash string, ethAddrs []string) ([32]byte, []common.Address, error) {
	fhash := util.StrHexToBytes32(fileHash)

	if me.hasNoActiveAccount() {
		return fhash, nil, os.ErrPermission
	}

	//ensure we can collect the needed public keys upfront
	_, err := me.collectAllPublicKeysFor(me.ActiveAccount(), fileHash)
	if err != nil {
		return fhash, nil, err
	}

	fi, err := me.ETHClient.FileInfo(fhash, false)
	if err != nil {
		return fhash, nil, err
	}

	addrs := me.ethStrAddrsToCommonAddr(ethAddrs)
	validAddrs := make(map[string]common.Address, 0)
	for _, localAddr := range addrs {
		for _, bcAddr := range fi.ReadAccess {
			if bytes.Equal(bcAddr.Bytes(), localAddr.Bytes()) {
				validAddrs[bcAddr.Hex()] = bcAddr
			}
		}
	}

	if len(validAddrs) > 0 {
		addrs = make([]common.Address, len(validAddrs))
		i := 0
		for _, addr := range validAddrs {
			addrs[i] = addr
			i++
		}

		return fhash, addrs, err
	}

	return fhash, nil, os.ErrInvalid
}

func (me *App) RevokeFileEstimateGas(fileHash string, ethAddrs []string) (GasEstimate, error) {
	var gasEstimate GasEstimate

	if me.hasNoActiveAccount() {
		return gasEstimate, ErrNoActiveAccount
	}
	fhash, addrs, err := me.revokeFilePrepare(fileHash, ethAddrs)
	if err != nil {
		return gasEstimate, err
	}

	opts, err := me.ETHClient.FileRevokePermEstimateGas(me.wallet.GetActiveAccountETHPrivateKey(), fhash, addrs)
	if err != nil {
		return gasEstimate, err
	}

	gasEstimate.GasPrice = opts.GasPrice
	gasEstimate.GasLimit = opts.GasLimit

	return gasEstimate, err
}

func (me *App) RevokeFile(fileHash string, ethAddrs []string) (string, error) {
	if me.hasNoActiveAccount() {
		return "", ErrNoActiveAccount
	}
	fhash, addrs, err := me.revokeFilePrepare(fileHash, ethAddrs)
	if err != nil {
		return "", err
	}

	filename := me.getFileNameByHash(fileHash, false)

	tx, err := me.ETHClient.FileRevokePerm(me.wallet.GetActiveAccountETHPrivateKey(), fhash, filename, addrs)
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), err
}

func (me *App) SignFileEstimateGas(fileHash string) (GasEstimate, error) {
	var gasEstimate GasEstimate

	if me.hasNoActiveAccount() {
		return gasEstimate, os.ErrPermission
	}

	fhash := util.StrHexToBytes32(fileHash)

	opts, err := me.ETHClient.FileSignEstimateGas(me.wallet.GetActiveAccountETHPrivateKey(), fhash)
	if err != nil {
		return gasEstimate, err
	}

	gasEstimate.GasPrice = opts.GasPrice
	gasEstimate.GasLimit = opts.GasLimit

	return gasEstimate, err
}

func (me *App) SignFile(fileHash string) (string, error) {
	if me.hasNoActiveAccount() {
		return "", os.ErrPermission
	}
	fhash := util.StrHexToBytes32(fileHash)

	filename := me.getFileNameByHash(fileHash, false)
	tx, err := me.ETHClient.FileSign(me.wallet.GetActiveAccountETHPrivateKey(), filename, fhash)
	if err != nil {
		return "", err
	}

	return tx.Hash().Hex(), err
}

func (me *App) RemovePlain(fileHash string) error {
	if me.hasNoActiveAccount() {
		return os.ErrPermission
	}

	//important to set fromMetaOnly = true, else we get an endless loop
	fileName := me.getFileNameByHash(fileHash, true)

	err := me.fileHandler.RemovePlainFromDisk(fileHash, fileName)
	if err != nil {
		log.Println("RemovePlain: Error removing file, failed to cleanup plain file for hash: ", fileHash)
		return err
	}
	log.Println("RemovePlain: Successfully removed plain file with hash: ", fileHash)
	return err
}

func (me *App) RemoveFileAndMetaFromDisk(fileHash string) {
	me.fileHandler.RemoveFileAndMetaFromDisk(fileHash)
}

func (me *App) RemoveFileFromDiskKeepMeta(fileHash string) {
	me.fileHandler.RemoveFileFromDiskKeepMeta(fileHash)
}

func (me *App) RemoveFileEstimateGas(fileHash string) (GasEstimate, error) {
	var gasEstimate GasEstimate

	if me.hasNoActiveAccount() {
		return gasEstimate, os.ErrPermission
	}

	fhash := util.StrHexToBytes32(fileHash)

	opts, err := me.ETHClient.FileRemoveEstimateGas(me.wallet.GetActiveAccountETHPrivateKey(), fhash)
	if err != nil {
		return gasEstimate, err
	}

	gasEstimate.GasPrice = opts.GasPrice
	gasEstimate.GasLimit = opts.GasLimit

	return gasEstimate, err
}

func (me *App) RemoveFile(fileHash string) (string, error) {
	if me.hasNoActiveAccount() {
		return "", os.ErrPermission
	}
	fhash := util.StrHexToBytes32(fileHash)

	filename := me.getFileNameByHash(fileHash, false)
	tx, err := me.ETHClient.FileRemove(me.wallet.GetActiveAccountETHPrivateKey(), fhash, filename)
	if err != nil {
		return "", err
	}
	return tx.Hash().Hex(), err
}

func (me *App) RemoveFileLocal(fileHash string) error {
	if me.hasNoActiveAccount() {
		return os.ErrPermission
	}

	fileMeta, err := me.fileHandler.FileMetaHandler.Get(fileHash)
	if err != nil {
		if err != file.ErrFileMetaNotFound {
			return err
		}
		fileMeta = new(file.FileMeta)
		fileMeta.FileHash = fileHash
	}

	fileMeta.Hidden = true

	err = me.fileHandler.FileMetaHandler.Put(fileMeta)
	if err != nil {
		return err
	}

	log.Println("[app][RemoveFileLocal] file hidden", fileMeta)
	me.pushFileStr(fileHash)

	return nil
}

func (me *App) loopAddrs(src []common.Address, dst *[]*account.AddressBookEntry) (iAmIncluded bool) {
	myAddr := me.wallet.GetActiveAccountETHCommonAddress()
	if myAddr == "" {
		return false
	}
	for _, r := range src {
		if myAddr == r.String() {
			iAmIncluded = true
		}
		addr := me.ensureItExistsInOurAddressBook(r.Hex())
		if addr == nil {
			continue
		}
		*dst = append(*dst, addr)
	}
	return
}

func (me *App) ensureItExistsInOurAddressBook(ethAddress string) *account.AddressBookEntry {
	abe, _ := me.addressBook.QuickInsertByETHAddr(fmt.Sprintf("Account %s", ethAddress[:5]), ethAddress, "")
	return abe
}

func (me *App) HasActiveAccount() bool {
	return me.HasActiveAndUnlockedAccount()
}

func (me *App) pushFileStr(fh string) {
	me.pushFileHash(util.StrHexToBytes32(fh))
}

func (me *App) pushFileHash(f common.Hash) {
	fho := me.ETHClient.FileHashToFileHashOrder(f)
	if fho != nil && me.isLoggedInState {
		me.fileInfoChan <- fho
	}
}

func (me *App) pushFileHashOrder(f *ethereum.FileHashOrder) {
	if !me.isLoggedInState {
		return
	}
	me.fileInfoChan <- f
}

//todo: the xesAmount per file is no longer calculated in smart contract (can be removed?)
func (me *App) XESAmountPerFile(providers []string) (*big.Int, error) {
	if me.hasNoActiveAccount() {
		return nil, os.ErrPermission
	}
	prvdrs := make([]common.Address, len(providers))
	for i := 0; i < len(prvdrs); i++ {
		prvdrs[i] = common.HexToAddress(providers[i])
	}
	return me.ETHClient.XESAmountPerFile(prvdrs)
}

func (me *App) ActiveAccount() *account.Account {
	return me.wallet.ActiveAccount()
}

func (me *App) hasNoActiveAccount() bool {
	return me.noActiveAccount(true)
}

// should be used when checking for active account inside a ticker func, else session-timeout is never triggered
func (me *App) hasNoActiveAccountDoNotSignalUserActivity() bool {
	return me.noActiveAccount(false)
}

func (me *App) noActiveAccount(signalUserActivity bool) bool {
	if !me.HasActiveAndUnlockedAccount() {
		return true
	}

	if signalUserActivity {
		if err := me.SignalUserActivity(); err != nil {
			log.Println("[app][hasNoActiveAccount] SignalUserActivity failed, err: ", err.Error())
		}
	}
	return false
}

func (me *App) sessionStart() {
	me.sessionHandler.Put("session", true)
	me.sessionHandler.OnExpired = func(key interface{}, val interface{}) {
		/*
			In normal case not OnExpired will trigger the timeout logout but the frontend (App.vue). This is because the frontend calls ping every other minute,
			which calls `SignalUserActivity` and therefore prevents OnExpired to happen. This is just a fail-safe in case frontend does not send ping.
		*/
		log.Println("[app][sessionStart] session-timeout pushed, will lock")
		me.push(EventMsg{Type: "session-timeout"})
		//needs to run in routine to prevent error with lock in ProxeusApp/memcache/cache.go#176 (s.cleanupLock.Lock()) when calling me.sessionHandler.Close
		go me.Logout()
	}
}

func (me *App) SignalUserActivity() error {
	var s bool
	err := me.sessionHandler.Get("session", &s)
	if err != nil {
		log.Println("user activity session err", err)
		return err
	}
	return nil
}

func (me *App) defaultStorageDir() string {
	return filepath.Join(me.cfg.StorageDir, "proxeus")
}

func (me *App) storageDirAccount() (string, error) {
	ethAddress := me.GetActiveAccountETHAddress()
	if ethAddress == "" {
		return "", ErrNoActiveAccount
	}
	return filepath.Join(me.defaultStorageDir(), ethAddress), nil
}

func (me *App) userAccountAppDir() (string, error) {
	var err error
	var storageDirAcc string

	if storageDirAcc, err = me.storageDirAccount(); err != nil {
		return "", err
	}
	return filepath.Join(storageDirAcc, userAccountAppDir), nil
}

// Retrieve quotes for all service providers
func (me *App) Quotes(durationDays int, sizeBytes int64) (Quote, error) {
	quote := Quote{}
	storageProviders, err := me.GetStorageProviders()
	if err != nil {
		return quote, err
	}
	for _, storageProvider := range storageProviders {

		fileSizesBytes := big.NewInt(sizeBytes)

		priceSize, err := storageProvider.PriceForSizeInXesWei(fileSizesBytes)
		if err != nil {
			log.Println("quote skipping storage provider: error PriceForSizeInXesWei", err)
			continue
		}
		priceDuration, err := storageProvider.PriceForDurationInXesWei(durationDays)
		if err != nil {
			log.Println("quote skipping storage provider: error PriceForDurationInXesWei", err)
			continue
		}
		priceTotal, err := storageProvider.TotalPriceForFile(durationDays, fileSizesBytes)
		if err != nil {
			log.Println("quote skipping storage provider: error TotalPriceForFile", err)
			continue
		}

		quoteProvider := QuoteProvider{
			Provider:      storageProvider,
			PriceSize:     priceSize.String(),
			PriceDuration: priceDuration.String(),
			PriceTotal:    priceTotal.String(),
			Available:     durationDays <= storageProvider.MaxStorageDays && sizeBytes <= storageProvider.MaxFileSizeByte,
		}
		quote.Providers = append(quote.Providers, quoteProvider)

		log.Printf(
			"dapp spp quote Provider: URL: %v | PriceDuration: %v | PriceSize: %v | PriceTotal: %v | sizeBytes: %v",
			quoteProvider.Provider.URL, quoteProvider.PriceDuration, quoteProvider.PriceSize, quoteProvider.PriceTotal, sizeBytes)
	}
	return quote, nil
}

//Logout is called when user clicks logout or if session expires
func (me *App) Logout() error {

	log.Println("Logout called")
	if err := me.Close(); err != nil {
		if err == ErrAlreadyLoggingOut {
			return nil
		}
		log.Println("Logout error: ", err.Error())
		return err
	}

	return nil
}

var ErrNoActiveAccount = errors.New("no active account found")

func (me *App) decryptUserData() error {
	me.accountDirLock.Lock()
	defer me.accountDirLock.Unlock()

	encryptedDir, err := me.userAccountAppDir()
	if err != nil {
		return err
	}
	encryptedFile := fmt.Sprintf("%s_%s", encryptedDir, "locked")
	if me.hasNoActiveAccount() {
		log.Println("decryptUserData: No active account found")
		return ErrNoActiveAccount
	}
	if _, err := os.Stat(encryptedFile); err != nil {
		log.Println("decryptUserData: No encrypted userdata found for account: ",
			me.GetActiveAccountETHAddress())
		return err
	}

	return me.fileHandler.DecryptDirectory(encryptedDir, encryptedFile, me.wallet.GetActiveAccountPGPPrivatePw(),
		me.wallet.GetActiveAccountPGPPrivateKey())
}

func (me *App) encryptUserData() error {
	me.accountDirLock.Lock()
	defer me.accountDirLock.Unlock()

	plainDir, err := me.userAccountAppDir()
	if err != nil {
		return err
	}
	encryptedFile := fmt.Sprintf("%s_%s", plainDir, "locked")
	err = me.fileHandler.EncryptDirectory(encryptedFile, plainDir, [][]byte{[]byte(me.wallet.GetActiveAccountPGPKey())})
	if err != nil {
		log.Println("encryptUserData: Error when encrypting userdata. Error: ", err.Error())
	}

	return err
}

var ErrAlreadyLoggingOut = errors.New("app is already logging out")

//Close is called if user clicks close button on window or if user clicks Logout or session expires
func (me *App) Close() error {
	log.Println("[app][Close] closing...")

	//prevent panic when trying to close already closed channels
	if false == me.isLoggedInState {
		return ErrAlreadyLoggingOut
	}
	me.isLoggedInState = false

	if me.accountDB != nil {
		me.accountDB.Close()
	}
	me.closeChannels()

	//close all except me.wallet because its shared between accounts
	me.stopWg.Wait()
	if me.fileHandler != nil {
		if err := me.fileHandler.Close(); err != nil {
			log.Println("[app][Close] Error when calling fileHandler.Close ", err.Error())
		}
	}
	if me.addressBook != nil {
		if err := me.addressBook.Close(); err != nil {
			log.Println("[app][Close] Error when calling addressBook.Close ", err.Error())
		}
	}
	if me.ETHClient != nil {
		if err := me.ETHClient.Close(); err != nil {
			log.Println("[app][Close] Error when calling ETHClient.Close ", err.Error())
		}
	}
	if me.sessionHandler != nil {
		me.sessionHandler.Close()
	}

	if me.notificationManager != nil {
		if err := me.notificationManager.Close(); err != nil {
			log.Println("[app][Close] Error when calling notificationManager.Close() ", err.Error())
		}
		me.notificationManager = nil
	}
	me.pushMsgs = false

	//encrypt data after closing (e.g. after addressBook.Close) because there might be data that is written on close
	if err := me.encryptUserData(); err != nil {
		log.Println("[app][Close] Logout error on encryptUserData: ", err.Error())
	} else {
		log.Println("[app][Close] successfully encrypted user account dir")
	}

	//wallet logout after encryptUserData because we need private key to encrypt directory
	err := me.wallet.Close()
	if err != nil {
		log.Println("[app][Close] Error when calling wallet.Close ", err.Error())
	}
	return err
}

func (me *App) closeChannels() {
	if me.stopchan != nil {
		close(me.stopchan)
	}
	if me.fileInfoChan != nil {
		close(me.fileInfoChan)
	}
}
