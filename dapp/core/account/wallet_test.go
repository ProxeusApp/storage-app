package account

import (
	"testing"

	"github.com/ProxeusApp/storage-app/dapp/core/embdb"
	"github.com/ethereum/go-ethereum/common"
)

func TestWallet_LoginWithNewAccount(t *testing.T) {
	wallet := &Wallet{walletUsageDB: embdb.OpenDummyDB(), cfg: &Config{}}
	err := wallet.LoginWithNewAccount("jesse", "securePassword")
	if err != nil {
		t.Error(err)
	}

	if wallet.activeAcc.name != "jesse" {
		t.Error("Expected active account to be 'jesse' but got: ", wallet.activeAcc.name)
	}

	accFileList := wallet.All()
	if len(accFileList) != 1 {
		t.Error("Expected to find 1 account but got ", len(accFileList))
	}

	if accFileList[0].Account().name != "jesse" {
		t.Error("Expected active account to be 'jesse' but got: ", accFileList[0].Account().name)
	}
}

func TestWallet_Logout(t *testing.T) {
	wallet := &Wallet{}
	addr := common.HexToAddress("0xa80899bb12e4afe9787425a5e5fe166234b88185")
	wallet.activeAcc = &Account{ethAddr: &addr}

	err := wallet.Logout()
	if err != nil {
		t.Error(err)
	}

	if wallet.activeAcc != nil {
		t.Errorf("Expetec activeAcc to be nil")
	}
}
