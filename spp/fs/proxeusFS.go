package fs

import (
	"encoding/hex"
	"errors"
	"io"
	"log"
	"math"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/ProxeusApp/storage-app/dapp/core/util"

	"github.com/ProxeusApp/storage-app/spp/service"

	"sync"

	"github.com/ethereum/go-ethereum/common"
	cache "github.com/patrickmn/go-cache"
	uuid "github.com/satori/go.uuid"

	"github.com/ProxeusApp/storage-app/lib/wallet"
	"github.com/ProxeusApp/storage-app/spp/config"
	"github.com/ProxeusApp/storage-app/spp/fs/db"
)

type ProxeusFS struct {
	basePath            string
	c                   *cache.Cache
	contractAddress     common.Address
	spAddress           common.Address
	database            *db.KVStore
	ethconn             FsClientInterface
	providerInfoService service.ProviderInfoService
	fileGblLock         sync.Mutex
	fileMetaHandler     FileMetaHandlerInterface
	testMode            bool
}

type SignMsg struct {
	Token     string `json:"token"`
	Challenge string `json:"challenge"`
}

type SpAddress struct {
	Address string `json:"address"`
	Url     string `json:"url"`
}

type FileInfo struct {
	Id             [32]byte
	Ownr           common.Address
	FileType       *big.Int
	Removed        bool
	Expiry         *big.Int
	IsPublic       bool
	ThumbnailHash  [32]byte
	Fparent        [32]byte
	ReplacesFile   [32]byte
	ReadAccess     []common.Address
	DefinedSigners []common.Address
}

type FsClientInterface interface {
	FileInfo(fileHash [32]byte, readFromCache bool) (FileInfo, error)
	SpInfoForFile(fileHash string) (string, error)
	GetFilePayment(fhash common.Hash) (*big.Int, error)
	HasWriteRights(fileHash [32]byte, addr common.Address, readFromCache bool) (bool, error)
	HasReadRights(fileHash [32]byte, addr common.Address, readFromCache bool) (bool, error)
	Close() error
}

const (
	downloadingSuffix = ".downloading"
	tempFolderName    = "tmp"
)

var (
	ErrNotSatisfiable      = errors.New("try again later")
	ErrInvalidSignature    = errors.New("login.error.invalidSignature")
	ErrNoPermission        = errors.New("proxeusFS permission denied")
	ErrFileRemoved         = errors.New("file has been removed")
	ErrFileNotReady        = errors.New("file isn't ready yet")
	ErrPaymentDoesNotMatch = errors.New("spp: received and calculated xes do not match")
)

func NewProxeusFS(cfg *config.Configuration, ethConn FsClientInterface, fileMetaHandler FileMetaHandlerInterface, providerInfoService service.ProviderInfoService) (*ProxeusFS, error) {
	if cfg.StorageDir != "" {
		err := os.MkdirAll(cfg.StorageDir, 0700)
		if err != nil {
			return nil, err
		}
	}
	pfs := &ProxeusFS{
		basePath:            cfg.StorageDir,
		spAddress:           common.HexToAddress(cfg.StorageProviderAddress),
		contractAddress:     common.HexToAddress(cfg.ContractAddress),
		c:                   cache.New(5*time.Minute, 10*time.Minute),
		ethconn:             ethConn,
		providerInfoService: providerInfoService,
		fileMetaHandler:     fileMetaHandler,
	}
	return pfs, nil
}

func (me *ProxeusFS) CreateSignInChallenge() (*SignMsg, error) {
	msg := wallet.CreateSignInChallenge("Sign this message to login: ")
	u, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	tokenStr := u.String()
	res := &SignMsg{Token: tokenStr, Challenge: msg}
	me.c.Set(tokenStr, res, cache.DefaultExpiration)
	return res, nil
}

func (me *ProxeusFS) verifyPayment(filehash string, fileSizeBytes int64, duration int) error {
	receivedXes, err := me.ethconn.GetFilePayment(util.StrHexToBytes32(filehash))
	if err != nil {
		return err
	}

	totalXes, err := me.calcTotalXes(duration, big.NewInt(int64(fileSizeBytes)))

	if 0 != receivedXes.Cmp(totalXes) {
		log.Printf("[proxeusFS][verifyPayment] file: %s | expectedXes: %d | receivedXes: %d", filehash, totalXes, receivedXes)
		return ErrPaymentDoesNotMatch
	}
	return err
}

var ErrReplacingExistingFileSize = errors.New("cannot replace file. Size of new and existing file differ to much")

func (me *ProxeusFS) Input(docHash, token, signatureHex string, body io.Reader, duration int) (written int64, err error) {
	addr, err := me.Validate(token, signatureHex)
	if err != nil {
		return 0, err
	}
	ok, err := me.hasPermission(docHash, addr, true)
	if err != nil {
		return 0, err
	}
	if !ok {
		return 0, ErrNoPermission
	}

	newFileSize, err := me.writeFileToDiskTmp(docHash, body)
	if err != nil {
		return
	}
	defer me.removeFileTmp(docHash) //remove tmp file if any

	err, isNewFile := me.checkForExistingFile(docHash, newFileSize)
	if err != nil {
		return 0, err
	}

	tmpPath := me.downloadingPath(docHash)
	untarDst := filepath.Join(me.basePath, tempFolderName, "decompressed", docHash)
	if err = verifyArchivePgpFiles(tmpPath, untarDst); err != nil {
		log.Println("Unencrypted files in archive found", err)
		return 0, err
	}

	//if file was on spp before do not check payment because filesize might change due to re-encryption with different amount of keys
	if isNewFile {
		// Since we don't want to copy the stream's content in memory,
		// we're only going to verify the payment once the file has been written to disk.
		// If the payment doesn't match or an error occurs we remove the file
		var fileInfo FileInfo
		if config.Config.IsTestMode() {
			log.Println("SPP running in TESTMODE. File payment won't be verified")
			fileInfo = FileInfo{
				Id:       util.StrHexToBytes32(docHash),
				Ownr:     common.HexToAddress(addr),
				FileType: big.NewInt(2),
			}
		} else {
			err = me.verifyPayment(docHash, newFileSize, duration)
			if err != nil {
				log.Println("Can't verify payment", err)
				return 0, err
			}

			fileHash := util.StrHexToBytes32(docHash)
			fileInfo, err = me.ethconn.FileInfo(fileHash, false)
			if err != nil {
				return 0, err
			}
		}
		me.fileMetaHandler.Save(fileInfo)
	}

	return newFileSize, me.moveFileTmpToRealDir(docHash)
}

//verify that the new uploaded file to replace the existing one (if any) is approx the same size as existing file
func (me *ProxeusFS) checkForExistingFile(docHash string, newFileSize int64) (err error, isNew bool) {
	existingFileInfo, err := os.Stat(filepath.Join(me.basePath, docHash))
	if err != nil {
		return nil, true
	}

	sizeDiff := math.Abs(float64(newFileSize) - float64(existingFileInfo.Size()))

	//1 sharee-key = ~1000 bytes, we'll allow a difference in size according to 100 sharees
	if sizeDiff > (100 * 1000) {
		log.Printf("[proxeusFS][Input] size existing file: %d | new file: %d", existingFileInfo.Size(), newFileSize)
		return ErrReplacingExistingFileSize, false
	}
	return nil, false
}

func (me *ProxeusFS) Output(docHashString, token, signatureHex string, force bool) (n string, err error) {
	docHash, err := strHashToBytes32(docHashString)
	if err != nil {
		return "", err
	}
	var addr string
	addr, err = me.Validate(token, signatureHex)
	if err != nil {
		return "", err
	}
	// Retrieve file first

	var fi FileInfo

	if config.Config.IsTestMode() {
		fi = FileInfo{
			Id:       docHash,
			Ownr:     common.HexToAddress(addr),
			FileType: big.NewInt(2),
		}
	} else {
		fi, err = me.ethconn.FileInfo(docHash, false)
		if err != nil {
			return "", err
		}

		if fi.Removed {
			return "", ErrFileRemoved
		}

		// Then check its permissions
		readRights, err := me.hasPermission(docHashString, addr, false)
		if err != nil {
			return "", err
		}
		if !readRights {
			return "", ErrNoPermission
		}
	}

	oldPath := filepath.Join(me.basePath, docHashString+downloadingSuffix)
	_, err = os.Stat(oldPath)
	if !os.IsNotExist(err) {
		err = ErrNotSatisfiable
		return
	}

	// Check if file meta exists, add otherwise
	_, err = me.fileMetaHandler.Get(docHash)
	if err == ErrSppFileMetaNotFound {
		me.fileMetaHandler.Save(fi)
	} else if err != nil {
		log.Println("proxeusFS::Output(): couldn't get file meta information: ", err)
	}

	return filepath.Join(me.basePath, docHashString), nil
}

func (me *ProxeusFS) Validate(token, signatureHex string) (addr string, err error) {
	if x, found := me.c.Get(token); found {
		signMsg := x.(*SignMsg)
		addr, err = wallet.VerifySignInChallenge(signMsg.Challenge, signatureHex)
		if err != nil {
			return
		}
		me.c.Delete(token)
	}
	return
}

func (me *ProxeusFS) writeFileToDisk(filename string, body io.Reader) (written int64, err error) {
	if written, err = me.writeFileToDiskTmp(filename, body); err != nil {
		return 0, err
	}

	return written, me.moveFileTmpToRealDir(filename)
}

func (me *ProxeusFS) moveFileTmpToRealDir(filename string) error {
	newPath := filepath.Join(me.basePath, filename)
	oldPath := me.downloadingPath(filename)
	return os.Rename(oldPath, newPath)
}

func (me *ProxeusFS) removeFileTmp(filename string) error {
	return os.Remove(me.downloadingPath(filename))
}

func (me *ProxeusFS) writeFileToDiskTmp(filename string, body io.Reader) (written int64, err error) {
	var f *os.File

	downloadingPath := me.downloadingPath(filename)
	//check if already
	_, err = os.Stat(downloadingPath)
	if !os.IsNotExist(err) {
		err = os.ErrInvalid
		return
	}
	f, err = os.OpenFile(downloadingPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		return
	}
	return io.Copy(f, body)
}

func (me *ProxeusFS) downloadingPath(filename string) string {
	downloadingFilename := filename + downloadingSuffix
	return filepath.Join(me.basePath, downloadingFilename)
}

func (me *ProxeusFS) removeFileFromDisk(filename string) (err error) {
	absPath := filepath.Join(me.basePath, filepath.Base(filename))
	err = os.Remove(absPath)
	if err == nil {
		log.Println("Removed file: ", absPath)
	}
	return
}

func strHashToBytes32(documentHashHex string) (documentHashBytesFixed [32]byte, err error) {
	documentHashBytes, err := hex.DecodeString(documentHashHex[2:])
	if err != nil {
		log.Printf("failed to decode document hash to bytes: %v", err)
		return
	}
	// convert the documentHash from hex to fixed size byte[]
	copy(documentHashBytesFixed[:], documentHashBytes[:32])
	return
}

func (me *ProxeusFS) hasPermission(docHashStr, addr string, write bool) (access bool, err error) {
	docHash, err := strHashToBytes32(docHashStr)
	if err != nil {
		return false, err
	}
	var filePerm, hasSP bool
	if write {
		filePerm, err = me.ethconn.HasWriteRights(docHash, common.HexToAddress(addr), true)
	} else {
		filePerm, err = me.ethconn.HasReadRights(docHash, common.HexToAddress(addr), true)
	}
	if filePerm {
		_, err := me.ethconn.SpInfoForFile(docHashStr)
		if err == nil {
			hasSP = true
		}
	}
	allow := filePerm && hasSP
	if !allow {
		log.Printf("permission denied [due to filePerm %v hasSP %v] fileHash %s addr %s authFrom %s write %v", filePerm, hasSP, docHashStr, addr, me.spAddress.String(), write)
	}
	return allow, err
}

func (me *ProxeusFS) calcTotalXes(durationInDays int, fileSizeByte *big.Int) (*big.Int, error) {
	storageProviderInfo := me.providerInfoService.Get()
	totalXes, err := storageProviderInfo.TotalPriceForFile(durationInDays, fileSizeByte)
	if err != nil {
		return totalXes, err
	}

	priceDuration, _ := storageProviderInfo.PriceForDurationInXesWei(durationInDays)
	priceSize, _ := storageProviderInfo.PriceForSizeInXesWei(fileSizeByte)
	pricePerDayXesWei, _ := storageProviderInfo.PriceDayXESWei()
	pricePerByteXesWei, _ := storageProviderInfo.PriceByteXESWei()

	log.Printf("spp upload: pricing totalXes: %v | durationInDays: %v | fileSizeByte: %v",
		totalXes, durationInDays, fileSizeByte)
	log.Printf("spp upload: priceSize: %v | priceDuration: %v | pricePerByteXesWei: %v | pricePerDayXesWei: %v",
		priceSize, priceDuration, pricePerByteXesWei, pricePerDayXesWei)

	return totalXes, nil
}

func (me *ProxeusFS) CheckForExpiredFiles() {
	start := time.Now()
	log.Println("CheckForExpiredFiles...")

	fileMetas, err := me.fileMetaHandler.All()
	if err != nil {
		return
	}

	now := time.Now().Unix()
	gracePeriod := int64(me.providerInfoService.Get().GraceSeconds)

	for _, fileMeta := range fileMetas {
		expiryWithGrace := fileMeta.Expiry.Int64() + gracePeriod

		if expiryWithGrace > now {
			continue
		}

		log.Println("CheckForExpiredFiles: Expired, would remove, double check expiry from smart contract...", fileMeta.FileHash.Hex())
		fileInfo, err := me.ethconn.FileInfo(fileMeta.FileHash, false)
		if err != nil {
			log.Println("CheckForExpiredFiles: error while double checking expiry date ", err)
			continue
		}

		if fileMeta.Expiry.Int64() != fileInfo.Expiry.Int64() {
			log.Printf("CheckForExpiredFiles: saved expiry date not identical with smart contract! Saved: %d, Smart contract: %d, FileHash: %s", fileMeta.Expiry.Int64(), fileInfo.Expiry.Int64(), fileMeta.FileHash.Hex())
			expiryWithGrace = fileInfo.Expiry.Int64() + gracePeriod

			// Check if expiry date from smart contract still is old enough to be removed, otherwise update file meta
			if expiryWithGrace > now {
				// Update spp file meta
				me.fileMetaHandler.Save(fileInfo)
				continue
			}
		}

		log.Println("CheckForExpiredFiles: actually remove... ", fileMeta.FileHash.Hex())
		err = me.removeFileFromDisk(fileMeta.FileHash.Hex())
		if err != nil {
			log.Println("CheckForExpiredFiles: error while remove file from disk ", err)
			if !os.IsNotExist(err) {
				continue
			}
		}
		// If no error occurred during remove file from disk or file didn't exist, remove meta information
		me.fileMetaHandler.Remove(fileMeta.FileHash)
	}

	elapsed := time.Since(start)

	log.Println("...CheckForExpiredFiles took ", elapsed)
}

func (me *ProxeusFS) Close() (err error) {
	me.ethconn.Close()
	return nil
}
