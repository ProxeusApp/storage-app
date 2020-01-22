package ethereum

import (
	"errors"
	"log"
	"math/big"

	"git.proxeus.com/core/central/spp/fs"

	"github.com/ethereum/go-ethereum/common"
)

type ClientMock struct {
	fileMetaHandler fs.FileMetaHandlerInterface
}

func NewClientMock(fileMetaHandler fs.FileMetaHandlerInterface) *ClientMock {
	me := new(ClientMock)
	me.fileMetaHandler = fileMetaHandler
	return me
}

func (me *ClientMock) FileInfo(fileHash [32]byte, readFromCache bool) (fi fs.FileInfo, err error) {

	fileMeta, err := me.fileMetaHandler.Get(fileHash)
	if fileMeta == nil || err != nil {
		return fs.FileInfo{}, errors.New("SPP file meta not found")
	}

	return fs.FileInfo{Id: fileMeta.FileHash, Expiry: fileMeta.Expiry}, nil
}

func (me *ClientMock) SpInfoForFile(fileHash string) (string, error) {
	log.Fatal("ClientMock::SpInfoForFile() not implemented")
	return "", nil
}
func (me *ClientMock) GetFilePayment(fhash common.Hash) (*big.Int, error) {
	log.Fatal("ClientMock::GetFilePayment() not implemented")
	return nil, nil
}

func (me *ClientMock) HasWriteRights(fileHash [32]byte, addr common.Address, readFromCache bool) (bool, error) {
	log.Fatal("ClientMock::HasWriteRights() not implemented")
	return false, nil
}
func (me *ClientMock) HasReadRights(fileHash [32]byte, addr common.Address, readFromCache bool) (bool, error) {
	log.Fatal("ClientMock::HasReadRights() not implemented")
	return false, nil
}
func (me *ClientMock) Close() error {
	log.Fatal("ClientMock::Close() not implemented")
	return nil
}
