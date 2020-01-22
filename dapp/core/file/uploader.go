package file

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strings"
	"sync"
	"time"

	cache "github.com/ProxeusApp/memcache"

	"git.proxeus.com/core/central/dapp/core/account"
	"git.proxeus.com/core/central/dapp/core/embdb"
	"git.proxeus.com/core/central/spp/client"
	"git.proxeus.com/core/central/spp/config"
	"git.proxeus.com/core/central/spp/fs"
)

const (
	PendingDBName = "pending"
)

type (
	Uploader struct {
		cfg                      *config.Configuration
		pendingDB                *embdb.DB
		wallet                   *account.Wallet
		fileUploadScheduledCache *cache.Cache
		accountGetter            func() *account.Account
		uploadTrigger            chan *Pending
		pendingTrigger           chan bool
		waitWorkerGrp            *sync.WaitGroup
		closing                  *bool
		stopAll                  *chan bool
		uploadDownloadSync       *map[string]*uploadDownloadStatus
		uploadDownloadSyncMutex  *sync.Mutex
		listener                 func(stype, fhash, spUrl, txHash, status, name string, percentage float32) error
		listenerLock             sync.RWMutex
	}
)

func NewUploader(cfg *config.Configuration, accountGetter func() *account.Account, waitWorkerGrp *sync.WaitGroup, closing *bool, stopAll *chan bool,
	uploadDownloadSync *map[string]*uploadDownloadStatus, uploadDownloadSyncMutex *sync.Mutex, wallet *account.Wallet, userAccountDir string) (*Uploader, error) {

	var err error

	up := &Uploader{}
	up.cfg = cfg
	up.accountGetter = accountGetter
	up.waitWorkerGrp = waitWorkerGrp
	up.closing = closing
	up.stopAll = stopAll
	up.uploadDownloadSync = uploadDownloadSync
	up.uploadDownloadSyncMutex = uploadDownloadSyncMutex
	up.wallet = wallet
	up.pendingDB, err = embdb.Open(userAccountDir, PendingDBName)
	up.fileUploadScheduledCache = cache.NewExtendExpiryOnGet(5*time.Minute, true)

	return up, err
}

func (me *Uploader) setupWorkers() {
	me.uploadTrigger = make(chan *Pending, 200)
	me.pendingTrigger = make(chan bool, 200)
}

func (me *Uploader) pendingTask() {
	if !me.wallet.HasActiveAndUnlockedAccount() {
		return
	}
	keys, err := me.pendingDB.FilterKeySuffix([]byte(me.wallet.GetActiveAccountETHAddress()))
	if err == nil {
		for _, k := range keys {
			bts, _ := me.pendingDB.Get(k)
			if len(bts) > 0 {
				p := Pending{}
				err = json.Unmarshal(bts, &p)
				if err != nil {
					log.Printf("error when trying to unmarshal value of %s\n", string(k))
					continue
				}
				me.pendingUploadTask(&p)
			}
		}
	}
}

func (me *Uploader) pendingUploadTask(pending *Pending) {
	if !me.wallet.HasActiveAndUnlockedAccount() || me.wallet.GetActiveAccountETHAddress() != pending.CurrentAddress ||
		!pending.ReadyForUpload {
		return
	}

	log.Println("[Uploader][pendingUploadTask] Pending upload task ", pending)

	var (
		noti *notified
		err  error
	)
	if err = me.fileUploadScheduledCache.Get(pending.FileHash, &noti); err != nil {
		log.Printf("[Uploader][pendingUploadTask] %s, pendingTask: %v", err.Error(), pending)
	}

	if noti == nil {
		noti = &notified{statusEnd: false, statusPending: false}
		me.fileUploadScheduledCache.Put(pending.FileHash, noti)
	}
	if noti.statusEnd && noti.statusPending {
		noti.statusEnd = false
		noti.statusPending = false
	}
	if pending.UploadedToSPP {
		//notify done
		if noti.statusEnd {
			return
		}

		noti.statusEnd = true
		err := me.notify(StatusUpload, pending.FileHash, pending.SpUrl, pending.TxHash, StatusSuccess, pending.FileName, 100)
		if err == nil {
			log.Println("[Uploader][pendingUploadTask] Deleting pending file hash from pendingDB ", pending.FileHash)
			_ = me.removePending(pending.FileHash)
		}

		return
	}

	//trigger upload
	if noti.statusPending {
		return
	}
	noti.statusPending = true
	_ = me.notify(StatusUpload, pending.FileHash, pending.SpUrl, pending.TxHash, StatusPending, pending.FileName, pending.Percentage)

	if err = me.sppUpload(me.accountGetter(), *pending); err != nil {
		noti.statusEnd = true
		_ = me.notify(StatusUpload, pending.FileHash, pending.SpUrl, pending.TxHash, StatusFail, pending.FileName, 0)
	}
}

func (me *Uploader) register(txHash, fileHash string, rdyForUpload bool) error {
	if !me.wallet.HasActiveAndUnlockedAccount() {
		return os.ErrPermission
	}

	bts, err := me.getPending(fileHash)
	if err != nil {
		return err
	}
	p := Pending{}
	if len(bts) > 0 {
		err = json.Unmarshal(bts, &p)
		if err != nil {
			log.Println("register unmarshal error", len(bts), string(bts))
			return err
		}
	}

	p.TxHash = txHash
	p.ReadyForUpload = rdyForUpload
	if bts, err = json.Marshal(p); err != nil {
		return err
	}

	err = me.putPending(fileHash, bts)
	if err == nil && rdyForUpload {
		log.Println("[Uploader][register] triggering uploadTrigger")
		me.uploadTrigger <- &p
	}
	return err
}

func (me *Uploader) notifyLastState() {
	go func() {
		defer func() {
			if r := recover(); r != nil { //can happen when closing
				log.Println("PendingTrigger file channel closed", r)
			}
		}()
		me.fileUploadScheduledCache.Clean()
		me.pendingTrigger <- true
	}()
}

func (me *Uploader) uploadHandler() {
	me.waitWorkerGrp.Add(1)
	go func() {
		log.Println("uploader started")
		ticker := time.NewTicker(time.Second * 60)

		defer func() {
			*me.closing = true
			ticker.Stop()
			me.waitWorkerGrp.Done()
			log.Println("uploader stopped")
		}()
		for {
			select {
			case <-ticker.C:
				me.pendingTask()
			case _, ok := <-me.pendingTrigger:
				if !ok {
					return
				}
				me.pendingTask()
			case p, ok := <-me.uploadTrigger:
				if !ok {
					return
				}
				me.pendingUploadTask(p)
			case <-*me.stopAll:
				return
			}
		}
	}()
}

// Returns filehash after having scheduled an upload of the encrypted file
func (me *Uploader) scheduleUpload(archiveFile EncryptedArchive, reg Register, publicKeys [][]byte, spUrl string, readyForUpload bool) (string, *Pending, error) {
	currentAccountETHAddr := strings.ToLower(me.wallet.GetActiveAccountETHAddress())
	pending := &Pending{
		CurrentAddress:  currentAccountETHAddr,
		ArchiveFilePath: archiveFile.AbsolutePath,
		FileName:        reg.FileName,
		UploadedToSPP:   false,
		ReadyForUpload:  readyForUpload,
		SpUrl:           spUrl,
		FileHash:        archiveFile.FileHash,
		DurationDays:    reg.DurationDays,
	}
	bts, err := json.Marshal(pending)
	if err != nil {
		return pending.ArchiveFilePath, pending, err
	}
	err = me.pendingDB.Put(me.getPendingDbKey(archiveFile.FileHash), bts)
	if err != nil {
		return "", pending, err
	}
	log.Println("[Uploader][scheduleUpload] triggering uploadTrigger")
	me.uploadTrigger <- pending
	return archiveFile.AbsolutePath, pending, nil
}

func (me *Uploader) sppUpload(acc *account.Account, pending Pending) error {
	if !acc.IsUnlocked() {
		return account.ErrAccountLocked
	}
	var spUrl string
	spUrl = pending.SpUrl
	if len(me.cfg.ForceSpp) > 10 {
		log.Println("[Uploader][sppUpload] file Handler upload -> forcing SPP URL ", me.cfg.ForceSpp)
		spUrl = me.cfg.ForceSpp
	}
	if spUrl == "" {
		return ErrEmptySpURL
	}

	//sync ----------------------------------------------
	downloadSyncKey := strings.ToLower(spUrl + pending.FileHash)
	me.uploadDownloadSyncMutex.Lock()

	upDownSync := *me.uploadDownloadSync
	downStatus := upDownSync[downloadSyncKey]
	if downStatus == nil {
		downStatus = &uploadDownloadStatus{}
		upDownSync[downloadSyncKey] = downStatus
	}
	me.uploadDownloadSync = &upDownSync
	me.uploadDownloadSyncMutex.Unlock()
	//wait if the process was started already by an other thread
	downStatus.mutex.Lock()
	defer downStatus.mutex.Unlock()
	//sync ----------------------------------------------

	//try a couple of times in case server is not reachable
	paymentNotFoundCount := 0
	for count := 1; count < 3; count++ {
		if *me.closing {
			return os.ErrClosed
		}

		var sleepDuration int
		if paymentNotFoundCount == 0 {
			sleepDuration = count
		} else {
			sleepDuration = int(math.Pow(float64(paymentNotFoundCount+1), 2)) //extend wait if there is an issue with filePayment
		}
		log.Printf("[Uploader][sppUpload] sleep %d seconds before upload", sleepDuration)
		time.Sleep(time.Second * time.Duration(sleepDuration))

		response, err := client.Challenge(spUrl)
		if err != nil || response == nil {
			log.Printf("[Uploader][sppUpload] Try %d: error connecting to SPP(%s) to request the challenge. Error: %s\n",
				count, spUrl, err.Error())
			continue
		}

		bts, err := ioutil.ReadAll(response.Body)
		response.Body.Close()
		if err != nil {
			log.Printf("[Uploader][sppUpload] Try %d: error when connecting to SPP(%s) Response from spp not readable. Error: %s\n",
				count, spUrl, err.Error())
			continue
		}
		resp := fs.SignMsg{}
		err = json.Unmarshal(bts, &resp)
		if err != nil {
			return err
		}
		sig, err := acc.SignWithETH([]byte(resp.Challenge))
		if err == account.ErrAccountLocked {
			return err
		}
		if *me.closing {
			return os.ErrClosed
		}
		if err != nil {
			log.Printf("[Uploader][sppUpload] Try %d: error when signing the challenge of the SPP(%s) with address %s\n",
				count, spUrl, acc.GetETHAddress())
			continue
		}
		var archiveFile *os.File
		archiveFile, err = os.Open(pending.ArchiveFilePath)
		if err != nil {
			log.Printf("[Uploader][sppUpload] error on open file: %s, err: %s", pending.FileHash, err.Error())
			return err
		}
		stat, err := archiveFile.Stat()
		if err != nil {
			return err
		}
		downStatus.closeSync.Lock()
		ctx, cancel := me.ctxWithCancel()
		downStatus.cancel = func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println("[Uploader][sppUpload] panic when cancelling download")
				}
			}()
			f := archiveFile
			cancel()
			if f != nil {
				f.Close()
			}
		}
		downStatus.closeSync.Unlock()
		var transferProgressCallback TransferProgressCallback = func(percentage float32) {
			pending.Percentage = percentage
			_ = me.notify(StatusUpload, pending.FileHash, pending.SpUrl, pending.TxHash, StatusPending, pending.FileName, percentage)
		}
		_, err = client.InputWithContext(spUrl, pending.FileHash, resp.Token, string(sig), archiveFile, ctx, stat.Size(), transferProgressCallback, pending.DurationDays)
		_ = archiveFile.Close()
		if err != nil {
			if err == client.ErrFilePaymentNotFound {
				paymentNotFoundCount++
				log.Printf("[uploader][sppUpload] ErrFilePaymentNotFound file: %s, paymentNotFoundCount: %d", pending.FileHash, paymentNotFoundCount)
				if paymentNotFoundCount >= 4 {
					break //if paymentNotFoundCount try up to 4 times
				}
				count-- //lower count if payment not found because we will try more than 3 times
				continue
			}
			log.Printf("[uploader][sppUpload] Try %d: error when uploading the file(%s) to the SPP(%s) with address %s. Error: %s\n",
				count, pending.FileHash, spUrl, acc.GetETHAddress(), err.Error())
			continue
		}
		pending.UploadedToSPP = true
		bts, err = json.Marshal(pending)
		if err != nil {
			return err
		}
		err = me.putPending(pending.FileHash, bts)
		if err != nil {
			continue
		}
		downStatus.closeSync.Lock()
		downStatus.cancel = nil
		downStatus.closeSync.Unlock()

		log.Println("[Uploader][sppUpload] (re-)triggering uploadTrigger")
		me.uploadTrigger <- &pending
		return nil
	}
	return ErrUploadTimeout
}

func (me *Uploader) notify(stype, fhash, spUrl, txHash, status, name string, percentage float32) error {
	me.listenerLock.RLock()
	defer me.listenerLock.RUnlock()
	if me.listener != nil {
		return me.listener(stype, fhash, spUrl, txHash, status, name, percentage)
	}
	return os.ErrClosed
}

func (me *Uploader) ctxWithCancel() (context.Context, context.CancelFunc) {
	return context.WithCancel(context.TODO())
}

func (me *Uploader) putPending(fileHash string, bts []byte) error {
	key := me.getPendingDbKey(fileHash)
	return me.pendingDB.Put(key, bts)
}

func (me *Uploader) removePending(fileHash string) error {
	key := me.getPendingDbKey(fileHash)
	return me.pendingDB.Del(key)
}

func (me *Uploader) getPending(fileHash string) ([]byte, error) {
	key := me.getPendingDbKey(fileHash)
	return me.pendingDB.Get(key)
}

//Returns the pendingDb key for pendingDb cache
func (me *Uploader) getPendingDbKey(fileHash string) []byte {
	return []byte(fileHash + "_" + me.wallet.GetActiveAccountETHAddress())
}

func (me *Uploader) setListener(f func(stype, fhash, spUrl, txHash, status, name string, percentage float32) error) {
	me.listenerLock.Lock()
	me.listener = f
	me.listenerLock.Unlock()
}

func (me *Uploader) remListener() {
	me.listenerLock.Lock()
	me.listener = nil
	me.listenerLock.Unlock()
}

func (me *Uploader) close() {
	if me.fileUploadScheduledCache != nil {
		me.fileUploadScheduledCache.Close()
	}
	if me.pendingDB != nil {
		me.pendingDB.Close()
	}
	if me.uploadTrigger != nil {
		close(me.uploadTrigger)
	}
	if me.pendingTrigger != nil {
		close(me.pendingTrigger)
	}
	me.remListener()
}

func (me *Uploader) stopWorkers() {
	me.fileUploadScheduledCache.Clean()
}
