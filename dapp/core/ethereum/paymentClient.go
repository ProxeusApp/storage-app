package ethereum

import (
	"errors"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ProxeusApp/storage-app/dapp/core/embdb"
	"github.com/ProxeusApp/storage-app/spp/eth"
)

type (
	paymentClient struct {
		fsClient      *fsClient
		listenerLock  *sync.RWMutex
		filePaymentDB *embdb.DB
	}
)

const (
	filePaymentStorageName = "filePayment"
)

func NewPaymentClient(baseClient *baseClient, storageDir, proxeusFSAddress string) (*paymentClient, error) {
	var err error
	me := &paymentClient{}
	me.fsClient, err = NewFsClient(baseClient, common.HexToAddress(proxeusFSAddress))
	if err != nil {
		return me, err
	}
	me.filePaymentDB, err = embdb.Open(storageDir, filePaymentStorageName)

	return me, err
}

func (me *paymentClient) LogAsPaymentReceived(lg *types.Log) *eth.ProxeusFSContractPaymentReceived {
	return me.fsClient.LogAsPaymentReceived(lg)
}

var ErrFilePaymentNotFound = errors.New("file payment not found")

func (me *paymentClient) GetFilePayment(fhash common.Hash) (*big.Int, error) {
	bts, _ := me.filePaymentDB.Get(me.fsClient.getMyFileHashKey(fhash))
	val := new(big.Int)
	if len(bts) > 0 {
		val.SetBytes(bts)
		return val, nil
	}
	return val, ErrFilePaymentNotFound
}

func (me *paymentClient) putFilePayment(fhash common.Hash, xesAmount *big.Int) error {
	return me.filePaymentDB.Put(me.fsClient.getMyFileHashKey(fhash), xesAmount.Bytes())
}

func (me *paymentClient) HandlePaymentReceivedEvent(fileHash common.Hash, xesAmount *big.Int) error {
	return me.putFilePayment(fileHash, xesAmount)
}

func (me *paymentClient) close() {
	me.filePaymentDB.Close()
}
