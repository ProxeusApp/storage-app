package fs

import (
	"encoding/json"
	"errors"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"

	"git.proxeus.com/core/central/dapp/core/embdb"
)

type (
	FileMetaHandlerInterface interface {
		Get(fileHash common.Hash) (*sppFileMeta, error)
		All() ([]*sppFileMeta, error)
		Remove(fileHash common.Hash)
		Save(fileInfo FileInfo)
		Manipulate(manipulatedFileInfo FileInfoMock)
	}

	fileMetaHandler struct {
		sppFileMetaDB *embdb.DB
	}

	sppFileMeta struct {
		FileHash common.Hash
		Expiry   *big.Int
	}
)

const sppFileMetaStorageName = "sppFileMeta"

func NewFileMetaHandler(storageDir string) (*fileMetaHandler, error) {
	var err error
	me := &fileMetaHandler{}

	me.sppFileMetaDB, err = embdb.Open(storageDir, sppFileMetaStorageName)

	return me, err
}

func (me *fileMetaHandler) put(fileMeta sppFileMeta) error {
	bts, err := json.Marshal(fileMeta)
	if err != nil {
		return err
	}

	return me.sppFileMetaDB.Put([]byte(strings.ToLower(fileMeta.FileHash.Hex())), bts)
}

func (me *fileMetaHandler) Manipulate(manipulatedFileInfo FileInfoMock) {
	fileInfo := FileInfo{
		Id:     manipulatedFileInfo.Id,
		Expiry: manipulatedFileInfo.Expiry,
	}

	me.Save(fileInfo)
}

var ErrSppFileMetaNotFound = errors.New("SPP file meta not found")

func (me *fileMetaHandler) Get(fileHash common.Hash) (*sppFileMeta, error) {
	bts, _ := me.sppFileMetaDB.Get([]byte(strings.ToLower(fileHash.Hex())))

	if len(bts) > 0 {
		fileMeta := sppFileMeta{}
		err := json.Unmarshal(bts, &fileMeta)
		if err != nil {
			log.Println("[fileMetaHandler][Get] deserialize sppFileMetaDB error: ", err, string(bts))
			return nil, err
		}
		return &fileMeta, err
	}

	return nil, ErrSppFileMetaNotFound
}

func (me *fileMetaHandler) delete(fileHash common.Hash) error {
	return me.sppFileMetaDB.Del([]byte(strings.ToLower(fileHash.Hex())))
}

func (me *fileMetaHandler) Save(fileInfo FileInfo) {
	fileMeta, err := me.Get(fileInfo.Id)
	if err == ErrSppFileMetaNotFound {
		fileMeta = &sppFileMeta{FileHash: fileInfo.Id, Expiry: fileInfo.Expiry}
	} else if err != nil {
		log.Println("[fileMetaHandler][Save] couldn't get file meta information: ", err)
		return
	} else {
		// Update existing file meta informations
		fileMeta.Expiry = fileInfo.Expiry
	}

	err = me.put(*fileMeta)
	if err != nil {
		log.Println("[fileMetaHandler][Save] error during putSppFileMeta: ", err)
		return
	}
}

func (me *fileMetaHandler) Remove(fileHash common.Hash) {
	fileMeta, err := me.Get(fileHash)
	if err != nil {
		log.Println("[fileMetaHandler][Remove] couldn't get file meta information: ", err)
		return
	}

	err = me.delete(fileMeta.FileHash)
	if err != nil {
		log.Println("[fileMetaHandler][Remove] error during delSppFileMeta", err)
		return
	}
}

func (me *fileMetaHandler) All() ([]*sppFileMeta, error) {
	_, vals, _ := me.sppFileMetaDB.AllWithValues()

	var fileMetas []*sppFileMeta

	for _, val := range vals {
		if len(val) > 0 {
			fileMeta := sppFileMeta{}
			err := json.Unmarshal(val, &fileMeta)
			if err != nil {
				log.Println("[fileMetaHandler][All] deserialize sppFileMetaDB error: ", err, string(val))
				return nil, err
			}

			fileMetas = append(fileMetas, &fileMeta)
		}
	}

	return fileMetas, nil
}

func (me *fileMetaHandler) close() {
	me.sppFileMetaDB.Close()
}
