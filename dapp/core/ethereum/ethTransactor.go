package ethereum

import (
	"context"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	"github.com/ProxeusApp/storage-app/dapp/core/ethglue"
	"github.com/ProxeusApp/storage-app/spp/config"
)

type (
	ethTransactor struct {
		baseClient              *baseClient
		xesTokenTransactorMutex sync.Mutex
		nonceManager            *ethglue.NonceManager
		tokenTransactorMutex    *sync.Mutex
	}
)

func NewEthTransactor(baseClient *baseClient, nonceManager *ethglue.NonceManager, tokenTransactorMutex *sync.Mutex) *ethTransactor {
	ethTransactor := new(ethTransactor)
	ethTransactor.baseClient = baseClient
	ethTransactor.nonceManager = nonceManager
	ethTransactor.tokenTransactorMutex = tokenTransactorMutex
	return ethTransactor
}

func (me *ethTransactor) ethTransferEstimateGas() (gasPrice *big.Int, gasLimit uint64, err error) {
	// Default gas limit for ETH transfer
	gasLimit = uint64(21000)
	gasPrice, err = me.baseClient.ethconn.SuggestGasPrice(context.Background())
	if err != nil {
		return gasPrice, gasLimit, err
	}

	return gasPrice, gasLimit, nil
}

func (me *ethTransactor) ethTransfer(ethPrivKeyFrom, ethAddressTo string, ethAmount *big.Int) (*types.Transaction, error) {
	signedTx, err := me.ethTransaction(ethPrivKeyFrom, ethAddressTo, ethAmount)
	if err != nil {
		return signedTx, err
	}
	me.baseClient.putXesTx("eth-transfer", common.HexToAddress(me.baseClient.currentAddress), signedTx, ethAmount)
	return signedTx, nil
}

func (me *ethTransactor) ethTransaction(ethPrivKeyFrom string, ethAddressTo string, ethAmount *big.Int) (*types.Transaction, error) {
	privateKey, err := crypto.HexToECDSA(ethPrivKeyFrom)
	if err != nil {
		return nil, err
	}
	// Default gas limit for ETH transfer
	gasLimit := uint64(21000)
	gasPrice, err := me.baseClient.ethconn.SuggestGasPrice(context.Background())

	me.xesTokenTransactorMutex.Lock()
	defer me.xesTokenTransactorMutex.Unlock()
	defer me.nonceManager.OnError(err)

	nonce := me.nonceManager.NextNonce()
	tx := types.NewTransaction(nonce.Uint64(), common.HexToAddress(ethAddressTo), ethAmount, gasLimit, gasPrice, nil)
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(config.GetChainId()), privateKey)
	if err != nil {
		return nil, err
	}

	err = me.baseClient.ethconn.SendTransaction(context.Background(), signedTx)

	return signedTx, err
}
