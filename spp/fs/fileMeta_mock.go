package fs

import (
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type (
	fileMetaHandlerMock struct {
		sppFileMetaDB  map[common.Hash]*sppFileMeta
		proxeusFSFiles []FileInfo
	}
	FileInfoMock struct {
		Id     [32]byte
		Expiry *big.Int
	}
)

func NewFileMetaClientMock(proxeusFSFileMocks []*FileInfoMock) *fileMetaHandlerMock {
	me := &fileMetaHandlerMock{sppFileMetaDB: make(map[common.Hash]*sppFileMeta)}

	var initialProxeusFSFiles []FileInfo
	for _, proxeusFSFileMock := range proxeusFSFileMocks {
		FileInfo := FileInfo{Id: proxeusFSFileMock.Id, Expiry: proxeusFSFileMock.Expiry}
		initialProxeusFSFiles = append(initialProxeusFSFiles, FileInfo)
		me.Save(FileInfo)
	}

	me.proxeusFSFiles = initialProxeusFSFiles

	return me
}

func (me *fileMetaHandlerMock) Manipulate(manipulatedFileInfo FileInfoMock) {
	fileInfo := FileInfo{
		Id:     manipulatedFileInfo.Id,
		Expiry: manipulatedFileInfo.Expiry,
	}

	me.Save(fileInfo)
}

func (me *fileMetaHandlerMock) Get(fileHash common.Hash) (*sppFileMeta, error) {
	for _, sppFileMeta := range me.sppFileMetaDB {
		if sppFileMeta.FileHash == fileHash {
			return sppFileMeta, nil
		}
	}
	return nil, ErrSppFileMetaNotFound
}

func (me *fileMetaHandlerMock) Save(FileInfo FileInfo) {
	fileMeta, err := me.Get(FileInfo.Id)
	if err == ErrSppFileMetaNotFound {
		fileMeta = &sppFileMeta{FileHash: FileInfo.Id, Expiry: FileInfo.Expiry}
	} else if err != nil {
		log.Println("client::Save(): couldn't get file meta information: ", err)
		return
	} else {
		// Update existing file meta informations
		fileMeta.Expiry = FileInfo.Expiry
	}

	me.sppFileMetaDB[fileMeta.FileHash] = fileMeta
}

func (me *fileMetaHandlerMock) Remove(fileHash common.Hash) {
	delete(me.sppFileMetaDB, fileHash)
}

func (me *fileMetaHandlerMock) All() ([]*sppFileMeta, error) {
	var allSppFileMeta []*sppFileMeta
	for _, sppFileMeta := range me.sppFileMetaDB {
		allSppFileMeta = append(allSppFileMeta, sppFileMeta)
	}

	return allSppFileMeta, nil
}
