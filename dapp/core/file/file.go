package file

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/color"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/disintegration/imaging"

	"github.com/ProxeusApp/storage-app/dapp/core/file/archive"
	"github.com/ProxeusApp/storage-app/dapp/core/file/crypt"

	"github.com/ProxeusApp/storage-app/dapp/core/notification"
	channelhub "github.com/ProxeusApp/storage-app/web"

	"context"
	"net/http"

	cache "github.com/ProxeusApp/memcache"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pborman/uuid"

	"github.com/ProxeusApp/storage-app/dapp/core/account"
	"github.com/ProxeusApp/storage-app/spp/client"
	"github.com/ProxeusApp/storage-app/spp/config"
	"github.com/ProxeusApp/storage-app/spp/fs"
)

type (
	Handler struct {
		cfg            *config.Configuration
		fileDir        string
		userAccountDir string

		FileMetaHandler     *handler
		stopAll             chan bool
		waitWorkerGrp       sync.WaitGroup
		workersLock         sync.Mutex
		workersRunning      bool
		notificationManager *notification.Manager
		chanHub             *channelhub.ChannelHub
		uploader            *Uploader

		wallet   *account.Wallet
		dummyAcc *account.Account

		uploadDownloadSync         map[string]*uploadDownloadStatus
		uploadDownloadSyncMutex    sync.Mutex
		archiveFileMutex           sync.Mutex
		fileDownloadScheduledCache *cache.Cache
		closing                    bool
	}

	uploadDownloadStatus struct {
		mutex     sync.Mutex
		cancel    context.CancelFunc
		closeSync sync.Mutex
	}

	Register struct {
		StorageProviderAddress string
		FileReader             io.Reader
		FileName               string
		FileKind               int
		FileSize               int64
		ThumbReader            io.Reader
		ThumbName              string
		DurationDays           int
	}

	Pending struct {
		CurrentAddress  string
		ArchiveFilePath string
		ReadyForUpload  bool
		UploadedToSPP   bool
		FileName        string
		TxHash          string
		SpAddress       string
		SpUrl           string
		FileHash        string
		Percentage      float32
		DurationDays    int
	}

	ReqFile struct {
		Account  *account.Account
		SpUrl    string
		FileHash string
	}

	notified struct {
		statusPending bool
		statusEnd     bool
	}

	TransferProgressCallback func(float32)

	EncryptedArchive struct {
		FileHash     string
		AbsolutePath string // Absolute position on disk
		Size         int64  // file length in bytes
	}
)

var (
	ErrEmptySpURL          = errors.New("empty spUrl")
	ErrDownloadTimeout     = errors.New("download timeout")
	ErrPGPDecryptionFailed = errors.New("pgp decryption failed")
	ErrUploadTimeout       = errors.New("upload timeout")
)

const (
	plain       = "plain"
	encrypted   = "encrypted"
	archiveName = "archive"

	filesName = "files"

	StatusDownload = "download"
	StatusUpload   = "upload"
	StatusSuccess  = "success"
	StatusPending  = "pending"
	StatusFail     = "fail"
)

func NewHandler(cfg *config.Configuration, wallet *account.Wallet, storageDir, userAccountDir string, accountGetter func() *account.Account) (*Handler, error) {
	if accountGetter == nil {
		return nil, os.ErrInvalid
	}
	fh := &Handler{cfg: cfg, fileDir: filepath.Join(storageDir, filesName), userAccountDir: userAccountDir}
	err := fh.ensure(fh.fileDir)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	fh.FileMetaHandler, err = newFileMeta(userAccountDir)
	if err != nil {
		return nil, err
	}
	fh.fileDownloadScheduledCache = cache.New(30 * time.Second)
	fh.uploadDownloadSync = make(map[string]*uploadDownloadStatus)
	fh.wallet = wallet
	fh.uploader, err = NewUploader(cfg, accountGetter, &fh.waitWorkerGrp, &fh.closing, &fh.stopAll, &fh.uploadDownloadSync, &fh.uploadDownloadSyncMutex,
		wallet, userAccountDir)

	return fh, err
}

func (me *Handler) setupWorkers() {
	me.workersLock.Lock()
	if !me.workersRunning {
		me.closing = false
		me.workersRunning = true
		me.stopAll = make(chan bool)
		me.uploader.setupWorkers()
		me.uploader.uploadHandler()
	}
	me.workersLock.Unlock()
}

func (me *Handler) stopWorkers() {
	me.workersLock.Lock()
	if me.workersRunning {
		close(me.stopAll)
		me.uploadDownloadSyncMutex.Lock()
		for _, ds := range me.uploadDownloadSync {
			ds.Close()
		}
		me.uploadDownloadSyncMutex.Unlock()
		me.waitWorkerGrp.Wait()
		me.workersRunning = false
		me.fileDownloadScheduledCache.Clean()
		me.uploader.stopWorkers()
	}
	me.workersLock.Unlock()
}

func (me *Handler) FileMatchesQuery(fileHash, query string) bool {
	queryMatch := false
	q := strings.ToLower(query)
	fileMeta, _ := me.FileMetaHandler.Get(fileHash)

	if query == "" {
		queryMatch = true
	} else if fileMeta != nil {
		queryMatch = strings.Contains(strings.ToLower(fileMeta.FileName), q) || strings.Contains(strings.ToLower(fileHash), q)
	} else {
		queryMatch = strings.Contains(strings.ToLower(fileHash), q)
	}

	return queryMatch
}

// Tries only if the download isn't scheduled already
func (me *Handler) TryDownload(spUrl, fileHash string) {
	var addedRecently bool

	k := spUrl + fileHash
	err := me.fileDownloadScheduledCache.Get(k, &addedRecently)
	if err != nil {
		me.fileDownloadScheduledCache.Put(k, true)
	}

	if !addedRecently && !me.closing {
		go func() {
			// Delaying file request to increase chance of success while the file is already on the smart contract
			// but still being uploaded to the SPP
			time.Sleep(15 * time.Second)
			_, err := me.RequestFileFromSpp(spUrl, fileHash)
			me.RemoveFileFromDiskKeepMeta(fileHash)
			if err != nil {
				log.Println("[fileHandler][TryDownload] error while requesting file from SPP:", err)
			}
		}()
	}
}

func (me *Handler) FileNameByHash(fileHash string, fromMetaOnly bool, spUrl string) string {
	if !me.wallet.HasActiveAndUnlockedAccount() {
		return ""
	}

	fileMeta, err := me.FileMetaHandler.Get(fileHash)
	if err == nil && fileMeta.FileName != fileHash {
		return fileMeta.FileName
	}

	if fromMetaOnly {
		return ""
	}

	filePath, err := me.RequestFileFromSpp(spUrl, fileHash)
	if err != nil {
		return ""
	}

	fileName := filepath.Base(filePath)

	if err = me.RemovePlainFromDisk(fileHash, fileName); err != nil {
		log.Printf("[fileHandler][FileNameByHash] Error removing file, failed to cleanup plain file for hash: %s, err: %s\n", fileHash, err.Error())
	}

	return fileName
}

func (me *Handler) SetListener(f func(stype, fhash, spUrl, txHash, status, name string, percentage float32) error, notificationManager *notification.Manager, chanHub *channelhub.ChannelHub) {
	me.uploader.setListener(f)
	me.notificationManager = notificationManager
	me.chanHub = chanHub
	me.setupWorkers()
}

func (me *Handler) NotifyLastState() {
	me.uploader.notifyLastState()
}

func (me *Handler) metaFileDir(fhash string) (string, error) {
	metaPath := filepath.Join(me.userAccountDir, filesMetaName, fhash)
	err := me.ensure(metaPath)
	return metaPath, err
}

func (me *Handler) mainAndPlainFileDir() (string, string, error) {
	var plainDir string
	tmpMainFileDir := filepath.Join(me.fileDir, uuid.NewRandom().String())
	err := me.ensure(tmpMainFileDir)
	if err != nil {
		log.Println("PrepareRegister ensure tmpMainFileDir error ", err)
		return tmpMainFileDir, plainDir, err
	}
	plainDir = filepath.Join(tmpMainFileDir, plain)
	err = me.ensure(plainDir)
	if err != nil {
		log.Println("PrepareRegister ensure plainDir error ", err)
		return tmpMainFileDir, plainDir, err
	}

	return tmpMainFileDir, plainDir, nil
}

func (me *Handler) moveFileTmpToNewMain(tmpMainFileDir, newMainFileDir string, reg Register) {
	var err error
	//prevent from file exists while renaming
	if err = os.RemoveAll(newMainFileDir); err != nil {
		log.Println("PrepareRegister RemoveAll error ", err.Error())
	}
	if err = os.Rename(tmpMainFileDir, newMainFileDir); err != nil {
		log.Println("PrepareRegister Rename error ", err.Error()) //just log error, we might already be writing directory
	}
}

// Prepare the files on the filesystem but doesn't do anything with it yet.
// Use PrepareRegisterAndScheduleUpload if you want to schedule an upload too
func (me *Handler) PrepareRegister(reg Register, publicKeys [][]byte) (EncryptedArchive, error) {
	var (
		encryptedArchive                         EncryptedArchive
		err                                      error
		fhash, plainDir, tmpMainFileDir, metaDir string
	)

	if tmpMainFileDir, plainDir, err = me.mainAndPlainFileDir(); err != nil {
		return encryptedArchive, err
	}

	filePath := filepath.Join(plainDir, reg.FileName)
	if err = me.storeFileOnDisk(filePath, reg.FileReader); err != nil {
		log.Println("PrepareRegister storeFilesOnTheDisk error ", err.Error())
		return encryptedArchive, err
	}

	if fhash, err = me.hashMainFile(filepath.Join(plainDir, reg.FileName)); err != nil {
		return encryptedArchive, err
	}

	if metaDir, err = me.metaFileDir(fhash); err != nil {
		return encryptedArchive, err
	}

	metaPath := filepath.Join(metaDir, archive.Thumb)
	err = me.storeFileOnDisk(metaPath, reg.ThumbReader)
	if err != nil {
		if err != ErrNoFileDefined {
			log.Println("PrepareRegister storeFilesOnTheDisk error ", err.Error())
			return encryptedArchive, err
		}
	} else {
		const maxSize = 600 //double the size of what is displayed in frontend (for retina displays)
		fileExtension := filepath.Ext(reg.ThumbName)
		if err = me.downsizeThumbnail(metaDir, archive.Thumb, fileExtension, maxSize); err != nil {
			//if unsupported just log any other error, do return error
			if err != imaging.ErrUnsupportedFormat {
				return encryptedArchive, err
			}
			log.Printf("[file][PrepareRegister] will not downsize image, file extension is not supported. Given extension: %s", fileExtension)
		}
	}

	newMainFileDir := filepath.Join(me.fileDir, fhash)
	me.moveFileTmpToNewMain(tmpMainFileDir, newMainFileDir, reg)

	if encryptedArchive, err = me.encryptAndArchive(reg, fhash, newMainFileDir, publicKeys); err != nil {
		return encryptedArchive, err
	}

	if err = me.RemovePlainFromDisk(fhash, reg.FileName); err != nil {
		log.Printf("PrepareRegister: RemovePlainFromDisk Error removing file, Failed cleanup plain file for hash: %s, err: %s", fhash, err.Error())
	}

	return encryptedArchive, err
}

// fileExtension "jpg" (or "jpeg"), "png", "gif", "tif" (or "tiff") and "bmp" are supported.
func (me *Handler) downsizeThumbnail(metaDir, newFilename, fileExtension string, maxSize int) error {
	if _, err := imaging.FormatFromFilename(fileExtension); err != nil {
		return err
	}
	thumbSrc, err := imaging.Open(filepath.Join(metaDir, newFilename))
	if err != nil {
		return err
	}

	x := thumbSrc.Bounds().Dx()
	y := thumbSrc.Bounds().Dy()

	min := math.Min(float64(x), float64(y))
	toSize := int(math.Min(min, float64(maxSize)))
	thumbSrc = imaging.Fill(thumbSrc, toSize, toSize, imaging.Center, imaging.Lanczos)

	dst := imaging.New(toSize, toSize, color.NRGBA{0, 0, 0, 0})
	dst = imaging.Paste(dst, thumbSrc, image.Pt(0, 0))

	//construct tempFilePath with fileExtension in order for imaging to "guess" fileExtension from it
	tempFilePath := filepath.Join(metaDir, fmt.Sprintf("%s%s", newFilename, fileExtension))
	if err = imaging.Save(dst, tempFilePath); err != nil {
		return err
	}
	metaPath := filepath.Join(metaDir, newFilename)
	if err = os.Rename(tempFilePath, metaPath); err != nil {
		return err
	}
	return nil
}

// Prepares the files on disk and schedules an upload
func (me *Handler) PrepareRegisterAndScheduleUpload(reg Register, publicKeys [][]byte, spUrl string) (encryptedArchive EncryptedArchive, err error) {
	archiveFile, err := me.PrepareRegister(reg, publicKeys)
	if err != nil {
		return archiveFile, err
	}

	_, pending, err := me.uploader.scheduleUpload(archiveFile, reg, publicKeys, spUrl, false)
	if err != nil {
		return archiveFile, err
	}

	hasThumbnail := reg.ThumbReader != nil

	err = me.FileMetaHandler.Put(&FileMeta{
		FileHash:     pending.FileHash,
		FileName:     pending.FileName,
		FileKind:     reg.FileKind,
		Uploaded:     false,
		Hidden:       false,
		SpUrl:        spUrl,
		HasThumbnail: hasThumbnail,
	})
	if err != nil {
		return archiveFile, err
	}

	return archiveFile, err
}

func (me *Handler) ensureMainFileDir(mainFileDir, dir string) (string, error) {
	dirPath := filepath.Join(mainFileDir, dir)
	err := me.ensure(dirPath)
	return dirPath, err
}

func (me *Handler) encryptAndArchive(reg Register, fhash, mainFileDir string, publicKeys [][]byte) (EncryptedArchive, error) {
	var (
		encryptedArchive EncryptedArchive
		archiveFilePath  string
		err              error
	)

	if !me.wallet.HasActiveAndUnlockedAccount() {
		return encryptedArchive, os.ErrPermission
	}
	if reg.FileKind == 2 {
		filelocation := filepath.Join(me.cfg.StorageDir, "sharefiles", reg.FileName)
		archiveFilePath = filelocation
	} else {

		if archiveFilePath, err = me.createArchiveForUpload(mainFileDir, fhash, reg, publicKeys); err != nil {
			return encryptedArchive, err
		}
	}

	fileInfo, err := os.Stat(archiveFilePath)
	if err != nil {
		log.Println("encryptAndArchive fileinfo error ", err)
		return encryptedArchive, err
	}
	encryptedArchive.FileHash = fhash
	encryptedArchive.AbsolutePath = archiveFilePath
	encryptedArchive.Size = fileInfo.Size()

	return encryptedArchive, err
}

func (me *Handler) createArchiveForUpload(mainFileDir, fhash string, reg Register, publicKeys [][]byte) (string, error) {
	var (
		err                                                 error
		archiveFilePath, encryptedDir, archiveDir, plainDir string
	)
	if plainDir, err = me.ensureMainFileDir(mainFileDir, plain); err != nil {
		return archiveFilePath, err
	}

	if encryptedDir, err = me.ensureMainFileDir(mainFileDir, encrypted); err != nil {
		return archiveFilePath, err
	}
	if archiveDir, err = me.ensureMainFileDir(mainFileDir, archiveName); err != nil {
		return archiveFilePath, err
	}

	encryptedFilePath := filepath.Join(encryptedDir, fhash)

	plainFilePath := filepath.Join(plainDir, reg.FileName)
	if err = crypt.EncryptFile(encryptedFilePath, plainFilePath, publicKeys); err != nil {
		log.Println("encryptAndArchive encryptFile error ", err)
		return archiveFilePath, err
	}

	//encrypt thumbnail
	if reg.ThumbName != "" {
		metaDir, err := me.metaFileDir(fhash)
		if err != nil {
			return "", err
		}

		thumbPathPlain := filepath.Join(metaDir, archive.Thumb)
		thumbPathEncrypted := filepath.Join(metaDir, archive.ThumbEncrypted)

		if err = crypt.EncryptFile(thumbPathEncrypted, thumbPathPlain, publicKeys); err != nil {
			return archiveFilePath, err
		}
		defer func() {
			if err = os.Remove(thumbPathEncrypted); err != nil {
				log.Println("[file][createArchiveForUpload] error: ", err.Error())
			}
		}()
	}

	//encrypt proxeus meta
	pm := &archive.ProxeusMeta{
		FileNameMap: map[string]string{
			fhash: reg.FileName,
		},
	}
	bts, err := json.Marshal(pm)
	var proxMetaArmoured []byte
	if proxMetaArmoured, err = crypt.Encrypt(bts, publicKeys); err != nil {
		return archiveFilePath, err
	}

	return me.archiveFileThumbAndMeta(reg, fhash, plainDir, archiveDir, encryptedFilePath, proxMetaArmoured)
}

func (me *Handler) archiveFileThumbAndMeta(reg Register, fhash, plainDir, archiveDir, encryptedFilePath string, proxMetaArmoured []byte) (string, error) {
	toTarList := []string{encryptedFilePath}
	if reg.ThumbName != "" {
		metaDir, err := me.metaFileDir(fhash)
		if err != nil {
			return "", err
		}
		toTarList = append(toTarList, filepath.Join(metaDir, archive.ThumbEncrypted))
	}
	archiveFilePath := filepath.Join(archiveDir, fhash)

	err := me.archiveFiles(proxMetaArmoured, toTarList, &archiveFilePath)
	return archiveFilePath, err
}

func (me *Handler) Register(txHash, fileHash string, rdyForUpload bool) error {
	return me.uploader.register(txHash, fileHash, rdyForUpload)
}

func (me *Handler) RemovePlainFromDisk(fileHash, filename string) error {
	if !me.wallet.HasActiveAndUnlockedAccount() || fileHash == "" {
		return nil
	}

	return os.Remove(filepath.Join(me.fileDir, fileHash, plain, filename))
}

func (me *Handler) RemoveFileAndMetaFromDisk(fileHash string) {
	me.removeFromDisk(fileHash, true)
}

func (me *Handler) RemoveFileFromDiskKeepMeta(fileHash string) {
	me.removeFromDisk(fileHash, false)
}

func (me *Handler) removeFromDisk(fileHash string, deleteMeta bool) {
	if !me.wallet.HasActiveAndUnlockedAccount() || fileHash == "" {
		return
	}

	if _, err := os.Stat(filepath.Join(me.fileDir, fileHash)); err == nil {
		log.Printf("[fileHandler][removeFromDisk] Removing file with hash: %s", fileHash)
		_ = os.RemoveAll(filepath.Join(me.fileDir, fileHash))
	}

	if !deleteMeta {
		return
	}
	_ = me.uploader.removePending(fileHash)

	metaFileDir, err := me.metaFileDir(fileHash)
	if err != nil {
		log.Printf("[fileHandler][removeFromDisk] Unable to delete meta data for file: %s, error: %s", fileHash, err.Error())
		return
	}
	_ = os.RemoveAll(metaFileDir)
}

// Only checking on disk
func (me *Handler) HasThumbnail(spUrl, fileHash string) bool {
	_, err := me.Thumbnail(spUrl, fileHash, true)
	return err == nil
}

func (me *Handler) Thumbnail(spUrl, fileHash string, fromCacheOnly bool) (string, error) {
	if !me.wallet.HasActiveAndUnlockedAccount() {
		return "", os.ErrPermission
	}
	fp, err := me.thumbnailFromDisk(fileHash)
	if err == nil {
		return fp, nil
	}
	if fromCacheOnly {
		return "", os.ErrNotExist
	}
	_, err = me.RequestFileFromSpp(spUrl, fileHash)
	if err != nil {
		return "", err
	}
	return me.thumbnailFromDisk(fileHash)
}

func (me *Handler) thumbnailFromDisk(fileHash string) (string, error) {
	metaDir, err := me.metaFileDir(fileHash)
	if err != nil {
		return "", err
	}
	thumbPath := filepath.Join(metaDir, archive.Thumb)
	_, err = os.Stat(thumbPath)
	if err == nil {
		return thumbPath, nil
	}
	return "", os.ErrNotExist
}

func (me *Handler) ReEncryptFile(spUrl, fileHash string, pgpPubKeys [][]byte) (string, error) {
	filePath, err := me.RequestFileFromSpp(spUrl, fileHash)
	if err != nil {
		return "", err
	}
	mainFileDir := filepath.Join(me.fileDir, fileHash)
	plainDir := filepath.Join(mainFileDir, plain)
	metaDir, _ := me.metaFileDir(fileHash)

	reg := Register{}

	files, err := ioutil.ReadDir(plainDir)

	if _, err = os.Stat(metaDir); err == nil {
		metaFiles, _ := ioutil.ReadDir(metaDir)
		files = append(files, metaFiles...)
	}

	if err != nil {
		log.Print(err)
	}
	count := 0 //TODO improve this to solve the doc collection pillar
	for _, f := range files {
		if f.Name() == archive.Thumb {
			reg.ThumbName = archive.Thumb
		} else {
			if count > 1 {
				return "", errors.New("multiple files not implemented yet")
			}
			count++
			reg.FileName = f.Name()
		}
	}
	if reg.FileName == "" {
		reg.FileName = filepath.Base(filePath)
	}

	var mainf *os.File
	var thumbf *os.File

	mainf, err = os.Open(filepath.Join(plainDir, reg.FileName))
	if err != nil {
		return "", err
	}
	if reg.ThumbName != "" {
		thumbf, err = os.Open(filepath.Join(metaDir, reg.ThumbName))
		if err != nil {
			return "", err
		}
		reg.ThumbReader = thumbf
	}
	reg.FileReader = mainf

	archiveFile, err := me.encryptAndArchive(reg, fileHash, mainFileDir, pgpPubKeys)
	if err != nil {
		return "", err
	}

	filePath, _, err = me.uploader.scheduleUpload(archiveFile, reg, pgpPubKeys, spUrl, true)
	if err != nil {
		log.Println("[fileHandler][ReEncryptFile] error while scheduling upload", err)
	}

	if mainf != nil {
		mainf.Close()
	}
	if thumbf != nil {
		thumbf.Close()
	}
	return filePath, err
}

type File struct {
	FilePath    string
	FileKind    int
	Uploaded    bool
	ContentName string
	Type        string
}

//download from spp
func (me *Handler) RequestFileFromSpp(spUrl, fileHash string) (string, error) {
	if !me.wallet.HasActiveAndUnlockedAccount() {
		return "", os.ErrPermission
	}

	if len(me.cfg.ForceSpp) > 10 {
		log.Println("get file -> forcing SPP URL ", me.cfg.ForceSpp)
		spUrl = me.cfg.ForceSpp
	}

	//sync ----------------------------------------------
	downloadSyncKey := strings.ToLower(spUrl + fileHash)
	me.uploadDownloadSyncMutex.Lock()
	downStatus := me.uploadDownloadSync[downloadSyncKey]
	if downStatus == nil {
		downStatus = &uploadDownloadStatus{}
		me.uploadDownloadSync[downloadSyncKey] = downStatus
	}
	me.uploadDownloadSyncMutex.Unlock()
	//wait if the process was started already by an other thread
	downStatus.mutex.Lock()
	defer downStatus.mutex.Unlock()
	//sync ----------------------------------------------

	mainFileDir := filepath.Join(me.fileDir, fileHash)
	plainDir := filepath.Join(mainFileDir, plain)
	archiveDir := filepath.Join(mainFileDir, archiveName)

	fileMeta, err := me.FileMetaHandler.Get(fileHash)
	name := fileHash
	if err == nil {
		name = fileMeta.FileName
	}
	me.uploader.notify(StatusDownload, fileHash, spUrl, "", StatusPending, name, 0.0)

	f, err := me.requestPlainFromSPP(spUrl, mainFileDir, plainDir, archiveDir, fileHash, downStatus, func(percentage float32) {
		me.uploader.notify(StatusDownload, fileHash, spUrl, "", StatusPending, name, percentage)
	})
	if err != nil {
		me.uploader.notify(StatusDownload, fileHash, spUrl, "", StatusFail, name, 0)
	}

	me.uploader.notify(StatusDownload, fileHash, spUrl, "", StatusSuccess, name, 100)

	log.Printf("File from SPP: %s, Version: %s, Uploaded here:  %t", fileHash, strconv.Itoa(f.FileKind), f.Uploaded)
	if f.FileKind == 2 && !f.Uploaded {
		m := map[string]interface{}{
			"status":    "pending",
			"name":      "workflow_request",
			"txHash":    fileHash,
			"hash":      fileHash,
			"fileName":  f.ContentName,
			"xesAmount": "0",
			"who":       ""}
		n, err := me.notificationManager.AddOrUpdateAndAppendEventData("workflow_request", fileHash, m, nil)
		if err != nil {
			fmt.Println(err)
		}
		err = me.chanHub.Broadcast("global", n)
		if err != nil {
			fmt.Println(err)
		}
	}
	if err != nil {
		log.Println("[fileHandler][RequestFileFromSpp] error while requesting file:", err)
		return f.FilePath, err
	}

	return f.FilePath, err
}

func (me *Handler) requestArchiveFromDisk(mainFileDir, archiveDir, fileHash string) (File, error) {
	return me.requestFromDisk(archiveName, fileHash)
}

func (me *Handler) requestEncryptedFromDisk(mainFileDir, encryptedDir, fileHash string) (File, error) {
	return me.requestFromDisk(encrypted, fileHash)
}

func (me *Handler) requestFromDisk(fileTypeDir, fileHash string) (File, error) {
	if !me.wallet.HasActiveAndUnlockedAccount() {
		return File{}, os.ErrPermission
	}

	mainFileDir := filepath.Join(me.fileDir, fileHash)
	_, err := os.Stat(mainFileDir)
	if err != nil {
		return File{}, os.ErrNotExist
	}

	f, err := me.tryGetFileFromDB(fileTypeDir, fileHash)
	if err != nil {
		return File{}, err
	}

	return f, nil
}

func (me *Handler) requestPlainFromDisk(mainFileDir, plainDir, archiveDir, fileHash string) (File, error) {
	f, err := me.requestFromDisk(plain, fileHash)
	if err == nil {
		return f, err
	}

	//compatibility check for files without archive
	_, err = os.Stat(filepath.Join(plainDir, fileHash))
	if err == nil {
		return File{filepath.Join(plainDir, fileHash), 0, true, "", plain}, nil
	}
	err = me.ensure(plainDir)
	if err != nil {
		return File{}, err
	}

	f, err = me.getPlainFileFromArchive(fileHash, plainDir, archiveDir, false)
	if err == nil {
		return f, err
	}

	return File{}, err
}

func (me *Handler) requestOnlyArchiveFromSPP(spUrl, mainFileDir, archiveDir, fileHash string,
	force bool, downStatus *uploadDownloadStatus, transferProgressCallback TransferProgressCallback) (File, error) {
	err := me.ensure(mainFileDir)
	if err != nil {
		return File{}, err
	}
	err = me.ensure(archiveDir)
	if err != nil {
		return File{}, err
	}

	archiveFilePath := filepath.Join(archiveDir, fileHash)
	fi, err := os.Stat(archiveFilePath)

	if err != nil || fi.Size() == 0 || force {
		err = me.downloadArchiveFromSpp(spUrl, fileHash, archiveFilePath, downStatus, transferProgressCallback)
		if err != nil {
			log.Printf("requestArchiveFromSPP: Error downloadArchiveFromSpp spUrl: %s, fileHash: %s, err: %s",
				spUrl, fileHash, err.Error())
		}
	}

	return File{archiveFilePath, 0, true, "", archiveName}, nil
}

func (me *Handler) requestPlainFromSPP(spUrl, mainFileDir, plainDir, archiveDir, fileHash string, downStatus *uploadDownloadStatus,
	transferProgressCallback TransferProgressCallback) (File, error) {

	f, err := me.requestOnlyArchiveFromSPP(spUrl, mainFileDir, archiveDir, fileHash, false, downStatus, transferProgressCallback)
	if err != nil {
		log.Println("[fileHandler][requestPlainFromSPP] error while requesting only archive from SPP:", err)
		return File{}, err
	}

	err = me.ensure(plainDir)
	if err != nil {
		log.Println("[fileHandler][requestPlainFromSPP] error while ensure plain dir:", err)
		return File{}, err
	}

	f, err = me.getPlainFileFromArchive(fileHash, plainDir, archiveDir, true)
	if err != nil {
		log.Println("[fileHandler][requestPlainFromSPP] error while getting plain file from archive:", err)
		return File{}, err
	}
	return f, err
}

// Triggers a download from the service provider
// progressReader has to be passed as a pointer since we won't have all the data we need until the HTTP request is made
func (me *Handler) downloadArchiveFromSpp(spUrl, fileHash, archiveFilePath string,
	downStatus *uploadDownloadStatus, transferProgressCallback TransferProgressCallback) (err error) {

	if !me.wallet.HasActiveAndUnlockedAccount() {
		return account.ErrAccountLocked
	}
	if spUrl == "" {
		return ErrEmptySpURL
	}

	var (
		bts         []byte
		sig         []byte
		archiveFile *os.File
		r           *http.Response
	)

	//try a couple of times in case server is not reachable and each time increment the waiting time before retry
	var sleep int
	for count := 1; count < 5; count++ {
		if me.closing {
			return os.ErrClosed
		}
		if count > 1 {
			sleep = count * 3
			log.Printf("[downloadArchiveFromSpp] Waiting %d seconds before retry", sleep)
			time.Sleep(time.Second * time.Duration(sleep))
		}
		//after sleep immediately check again if closing
		if me.closing {
			return os.ErrClosed
		}
		r, err = client.Challenge(spUrl)
		if err != nil || r == nil {
			log.Printf("try %d: error when connecting to SPP(%s) to request the challenge err %v \n",
				count, spUrl, err)
			continue
		}
		bts, err = ioutil.ReadAll(r.Body)
		r.Body.Close()
		if err != nil {
			log.Printf("try %d: error when connecting to SPP(%s) response not readable %v \n",
				count, spUrl, err)
			continue
		}
		resp := fs.SignMsg{}
		err = json.Unmarshal(bts, &resp)
		if err != nil {
			continue
		}
		sig, err = me.wallet.SignWithETHofActiveAccount([]byte(resp.Challenge))
		if err == account.ErrAccountLocked {
			return err
		}
		if me.closing {
			return os.ErrClosed
		}
		if err != nil {
			log.Printf("try %d: error when signing the challenge of the SPP(%s) with address %s err %v \n",
				count, spUrl, me.wallet.GetActiveAccountETHAddress(), err)
			continue
		}
		archiveFile, err = os.OpenFile(archiveFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
		if err != nil {
			archiveFile.Close()
			continue
		}
		downStatus.closeSync.Lock()
		ctx, cancel := me.uploader.ctxWithCancel()
		downStatus.cancel = func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println("panic when cancelling download")
				}
			}()
			f := archiveFile
			cancel()
			if f != nil {
				f.Close()
			}
		}
		downStatus.closeSync.Unlock()

		_, err = client.OutputWithContext(spUrl, fileHash, resp.Token, string(sig), false, archiveFile, ctx, transferProgressCallback)
		if me.closing {
			return os.ErrClosed
		}
		if err == fs.ErrNotSatisfiable {
			return err
		}
		if err != nil {
			log.Printf("try %d: error when downloading the file(%s) from the SPP(%s) with address %s err %v \n", count, fileHash, spUrl, me.wallet.GetActiveAccountETHAddress(), err)
			os.Remove(archiveFilePath)
			continue
		}
		err = archiveFile.Close()
		if err != nil {
			continue
		}

		downStatus.closeSync.Lock()
		downStatus.cancel = nil
		downStatus.closeSync.Unlock()
		return nil
	}

	return ErrDownloadTimeout
}

func (me *Handler) getPlainFileFromArchive(fileHash, plainDir, archiveDir string, fromSpp bool) (File, error) {
	if !me.wallet.HasActiveAndUnlockedAccount() {
		return File{}, os.ErrPermission
	}

	//check if archive exists
	var err error
	var pm *archive.ProxeusMeta
	var archFile *os.File
	archiveFile := filepath.Join(archiveDir, fileHash)
	_, err = os.Stat(archiveFile)
	if err != nil {
		return File{}, err
	}
	archFile, err = os.Open(archiveFile)
	if err != nil {
		return File{}, err
	}
	pm, err = archive.UntarProxeusArchive(plainDir, archFile, me.wallet.GetActiveAccountPGPPrivatePw(), me.wallet.GetActiveAccountPGPPrivateKey())
	archFile.Close()
	if err != nil {
		log.Println("[fileHandler][getPlainFileFromArchive] error while untar proxeus archive:", err)
		//clean plainDir
		er := os.RemoveAll(plainDir)
		if er == nil {
			er = me.ensure(plainDir)
			if er != nil {
				return File{}, er
			}
		}
		//we are downloading the old file
		//the new file is maybe still being uploaded
		if strings.Contains(err.Error(), "incorrect key") {
			log.Printf("error when decrypting file(%s)\n", fileHash)
			return File{}, ErrPGPDecryptionFailed
		}
	}
	if err == archive.ErrNoProxeusArchive {
		err = crypt.DecryptFile(filepath.Join(plainDir, fileHash), archiveFile, me.wallet.GetActiveAccountPGPPrivatePw(), me.wallet.GetActiveAccountPGPPrivateKey())
		if err != nil {
			return File{}, err
		}
		return File{filepath.Join(plainDir, fileHash), 1, false, "", plain}, nil
	}
	if err != nil {
		return File{}, err
	}
	hasThumbnail := me.moveThumbFromPlainToMeta(plainDir, fileHash)

	for fhash, fname := range pm.FileNameMap {
		fileMetaChanged := false
		fileMeta, err := me.FileMetaHandler.Get(fhash)
		if err == ErrFileMetaNotFound {
			fileMeta = &FileMeta{
				FileHash:     fhash,
				FileName:     fname,
				FileKind:     1,
				Uploaded:     fromSpp,
				Hidden:       false,
				HasThumbnail: hasThumbnail,
			}
			fileMetaChanged = true
		} else if err != nil {
			log.Println("[fileHandler][getPlainFileFromArchive] couldn't get fileMeta for", fhash)
			continue
		} else {
			if fileMeta.FileName != fname {
				fileMeta.FileName = fname
				fileMetaChanged = true
			}
			if fileMeta.HasThumbnail != hasThumbnail {
				fileMeta.HasThumbnail = hasThumbnail
				fileMetaChanged = true
			}
			if !fileMeta.Uploaded && fromSpp {
				fileMeta.Uploaded = true
				fileMetaChanged = true
			}
		}

		if fileMetaChanged {
			err = me.FileMetaHandler.Put(fileMeta)
			if err != nil {
				log.Println("[fileHandler][getPlainFileFromArchive] error while saving fileMeta", fileMeta)
				continue
			}
		}
	}

	f, err := me.tryGetFileFromDB(plain, fileHash)
	if err != nil {
		log.Println("[fileHandler][getPlainFileFromArchive] error while trying to get file from DB:", err)
		return File{}, err
	}

	return f, nil
}

//move thumbnail to metadata if exists
func (me *Handler) moveThumbFromPlainToMeta(plainDir, fileHash string) (hasThumbnail bool) {
	existingThumbPath := filepath.Join(plainDir, archive.Thumb)
	_, err := os.Stat(existingThumbPath)
	if err == nil {
		if metaDir, err := me.metaFileDir(fileHash); err == nil {
			newThumbPath := filepath.Join(metaDir, archive.Thumb)
			if err = os.Rename(existingThumbPath, newThumbPath); err != nil {
				log.Printf("[file][getPlainFileFromArchive] error moving thumb to metadata: %s", err.Error())
			}
			return true
		}
	}
	return false
}

func (me *Handler) tryGetFileFromDB(fileTypeDir, fileHash string) (File, error) {
	fileMeta, err := me.FileMetaHandler.Get(fileHash)
	if err != nil {
		log.Println("[fileHandler][tryGetFileFromDB] fileMeta does not exist", err)
		return File{}, os.ErrNotExist
	}

	filename := fileHash
	if fileTypeDir == plain {
		filename = fileMeta.FileName
	}

	mainFileDir := filepath.Join(me.fileDir, fileHash)
	filePath := filepath.Join(mainFileDir, fileTypeDir, filename)
	_, err = os.Stat(filePath)
	if err != nil {
		log.Println("[fileHandler][tryGetFileFromDB] os.Stat failed", err)
		return File{}, os.ErrNotExist
	}

	return File{
		filePath,
		fileMeta.FileKind,
		fileMeta.Uploaded,
		fileMeta.ContentName,
		fileTypeDir,
	}, nil
}

func (me *Handler) EncryptDirectory(dst string, src string, pgpPublicKeys [][]byte) error {
	return crypt.EncryptDirectory(dst, src, pgpPublicKeys)
}

func (me *Handler) DecryptDirectory(dst string, src string, pw, pgpPrivateKey []byte) error {
	return crypt.DecryptDirectory(dst, src, pw, pgpPrivateKey)
}

func (me *Handler) archiveFiles(proxMetaArmoured []byte, srcList []string, dst *string) error {
	fff, err := os.OpenFile(*dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Println("archiveFiles openFile error", err)
		return err
	}
	defer fff.Close()
	err = archive.TarFileList(proxMetaArmoured, srcList, fff)
	if err != nil {
		return err
	}
	return nil
}

func (me *Handler) hashMainFile(mainFilePath string) (string, error) {
	//read main file to hash
	f, err := os.OpenFile(mainFilePath, os.O_RDONLY, 0600)
	if err != nil {
		return "", err
	}
	fhash, err := me.hashStream(f)
	if err != nil {
		return "", err
	}
	err = f.Close()
	if err != nil {
		return "", err
	}
	return fhash, nil
}

var ErrNoFileDefined = errors.New("No file defined")
var ErrNoBytesWritten = errors.New("No bytes written to file")

func (me *Handler) storeFileOnDisk(fileDst string, reader io.Reader) error {
	if fileDst == "" || reader == nil {
		return ErrNoFileDefined
	}

	f, err := os.OpenFile(fileDst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	written, err := io.Copy(f, reader)
	err = f.Close()
	if err != nil {
		return err
	}
	if written < 1 {
		return ErrNoBytesWritten
	}
	return nil
}

func (me *Handler) hashStream(r io.Reader) (string, error) {
	bts, err := ioutil.ReadAll(r)
	if err != nil {
		return "", err
	}
	hash := crypto.Keccak256(bts)
	docHash := "0x" + hex.EncodeToString(hash)
	return docHash, nil
}

func (me *Handler) ensure(p string) error {
	var err error
	_, err = os.Stat(p)
	if os.IsNotExist(err) {
		err = os.MkdirAll(p, 0750)
		if err != nil {
			return err
		}
	}
	return nil
}

func (me *Handler) Close() error {
	me.stopWorkers()
	me.FileMetaHandler.close()
	me.uploader.close()
	me.fileDownloadScheduledCache.Close()
	return nil
}

func (me *uploadDownloadStatus) Close() error {
	if me.cancel != nil {
		me.closeSync.Lock()
		if me.cancel != nil {
			log.Println("download or upload cancel..")
			me.cancel()
			me.cancel = nil
		}
		me.closeSync.Unlock()
	}
	return nil
}
