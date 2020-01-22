package ethereum

import (
	"math/big"
	"path/filepath"
	"sync"

	"git.proxeus.com/core/central/spp/fs"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"

	"git.proxeus.com/core/central/spp/eth"

	"github.com/ethereum/go-ethereum/common"
)

type (
	SppClient struct {
		baseClient    *baseClient
		paymentClient *paymentClient
		fsClient      *fsClient
		eventsHandler func(lg *types.Log, recent bool) error
		pfsAddress    string
	}
)

const (
	dataDir = "data"
)

func NewSppClient(ethClientURL, wsUrl, storageDir, pfsAddress string) (*SppClient, error) {
	var err error

	baseClient := NewBaseClient(&sync.RWMutex{}, wsUrl, ethClientURL, []common.Address{common.HexToAddress(pfsAddress)})
	baseClient.ethconn, err = ethclient.Dial("http://localhost/")

	me := &SppClient{}
	me.baseClient = baseClient
	me.pfsAddress = pfsAddress
	if me.paymentClient, err = NewPaymentClient(baseClient, storageDir, pfsAddress); err != nil {
		return me, err
	}
	me.fsClient, err = NewFsClient(baseClient, common.HexToAddress(pfsAddress))
	if err != nil {
		return me, err
	}

	return me, err
}

func (me *SppClient) SpInfoForFile(fileHash string) (string, error) {
	return me.fsClient.spInfoForFile(fileHash)
}
func (me *SppClient) FileInfo(fileHash [32]byte, readFromCache bool) (fi fs.FileInfo, err error) {
	return me.fsClient.FileInfo(fileHash, readFromCache)
}
func (me *SppClient) GetFilePayment(fhash common.Hash) (*big.Int, error) {
	return me.paymentClient.GetFilePayment(fhash)
}
func (me *SppClient) HasWriteRights(fileHash [32]byte, addr common.Address, readFromCache bool) (bool, error) {
	return me.fsClient.hasWriteRights(fileHash, addr, readFromCache)
}
func (me *SppClient) HasReadRights(fileHash [32]byte, addr common.Address, readFromCache bool) (bool, error) {
	return me.fsClient.hasReadRights(fileHash, addr, readFromCache)
}

func (me *SppClient) LogAsPaymentReceived(lg *types.Log) *eth.ProxeusFSContractPaymentReceived {
	return me.paymentClient.LogAsPaymentReceived(lg)
}

func (me *SppClient) HeaderByNumber(number *big.Int) (*types.Header, error) {
	return me.baseClient.HeaderByNumber(number)
}

func (me *SppClient) HandlePaymentReceivedEvent(fileHash common.Hash, xesAmount *big.Int) error {
	return me.paymentClient.HandlePaymentReceivedEvent(fileHash, xesAmount)
}

func (me *SppClient) initLocalStorage(storageDir string) error {
	var err error

	if storageDir == "" {
		storageDir = "."
	}
	storageDir = filepath.Join(storageDir, dataDir)
	if err != nil {
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

func (me *SppClient) updateEthInterfaces() error {
	var err error

	me.fsClient.proxeusFSContractCaller, err = eth.NewProxeusFSContractCaller(common.HexToAddress(me.pfsAddress), me.baseClient.ethconn)
	if err != nil {
		return err
	}

	return nil
}

func (me *SppClient) InitListeners(storDir, ethAddress string, eventHandler func(lg *types.Log, recent bool) error) error {
	if err := me.initLocalStorage(storDir); err != nil {
		return err
	}
	me.baseClient.initListeners(ethAddress, nil, nil, nil, eventHandler, me.updateEthInterfaces)
	me.baseClient.startWorkers()

	return nil
}

func (me *SppClient) Close() error {
	if me.paymentClient != nil {
		me.paymentClient.close()
	}
	if me.baseClient != nil {
		me.baseClient.close()
	}
	if me.fsClient != nil {
		me.fsClient.close()
	}

	return nil
}
