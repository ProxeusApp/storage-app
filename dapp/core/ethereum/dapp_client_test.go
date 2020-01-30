package ethereum

import (
	"fmt"
	"log"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/ProxeusApp/storage-app/dapp/core/util"
	"github.com/ProxeusApp/storage-app/spp/config"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

//TODO impl ethereum.Client API tests

var dapp *DappClient

func TestMain(m *testing.M) {
	config.Setup()
	cfg := &config.Config
	var err error

	log.Println(cfg)
	dapp, err = NewDappClient(cfg.EthClientURL, cfg.EthWebSocketURL, cfg.XESContractAddress, cfg.ContractAddress)

	//wait for connection to be ready
	time.Sleep(2 * time.Second)

	if err != nil {
		panic(err)
	}
	dapp.InitListeners("", "0x531da9adba2099718c3c084c9af04319de8297c7", func(tx *PendingTx, txHash, status string) error {
		return nil
	}, func(lg *types.Log, recent bool) error {
		return nil
	})

	ret := m.Run()
	os.Exit(ret)
}

func TestNewClient(t *testing.T) {

	fi, err := dapp.FileInfo(util.StrHexToBytes32("0x21b081f4a38aa9de10e6512e1d51649047316fd6cd97d9d27711843d9da1260f"), true)
	if err != nil {
		log.Println("TestNewClient file 1 failed ", err)
	}
	log.Println(fi)
	fi, err = dapp.FileInfo(util.StrHexToBytes32("0x88c68947a5e5dfbe63a99176f461396aa7ba43341d4938d2cdbc01f120032745"), true)
	if err != nil {
		log.Println("TestNewClient file 2 failed ", err)
	}
	log.Println(fi)
	fi, err = dapp.FileInfo(util.StrHexToBytes32("0xcdf331900373fb6992847b99401a88febae53bd81670a76d8133e8229ba4adcb"), true)
	if err != nil {
		log.Println("TestNewClient file 3 failed ", err)
	}
	log.Println(fi)
	fi, err = dapp.FileInfo(util.StrHexToBytes32("0xd13afc2a3f3ef1255ed1dc178a21a15449e1905985eb22e0ab6e4adae63fa623"), true)
	if err != nil {
		log.Println("TestNewClient file 4 failed ", err)
	}
	log.Println(fi)
}

// FileSettingsService

func TestEthereumBugs(t *testing.T) {

	doXes := func() {
		tx, e := dapp.XESApproveToProxeusFS(
			"b8f30c0313f27fd73617d79257aec18c4521956fc0cad44618b0279b6843dda8",
			big.NewInt(-8446744073709551616))
		fmt.Println("tx error:", e)
		if e == nil {
			fmt.Println("tx hash:", tx.Hash().String())
		}
	}

	doFailingInRuntime := func() {
		tx, e := dapp.FileSetPerm(
			"b8f30c0313f27fd73617d79257aec18c4521956fc0cad44618b0279b6843dda8",
			common.HexToHash("0xa80899bb12e4afe9787425a5e5fe166234b88184"),
			"xxxx",
			[]common.Address{common.HexToAddress("0xa80899bb12e4afe9787425a5e5fe166234b88184")},
		)
		fmt.Println("tx error:", e)
		if e == nil {
			fmt.Println("failing tx hash:", tx.Hash().String())
		}
	}

	dapp.baseClient.nonceManager.OnAccountChange("0xa80899bb12e4afe9787425a5e5fe166234b88184")

	{
		dapp.baseClient.nonceManager.DebugPrint()
		doXes()
	}

	{
		// this case is not increasing nonce
		dapp.baseClient.nonceManager.DebugPrint()
		TestingModeBreakGas = true
		doXes()
		TestingModeBreakGas = false
	}

	{
		dapp.baseClient.nonceManager.DebugPrint()
		doXes()
	}

	{
		// this case is not increasing nonce
		dapp.baseClient.nonceManager.DebugPrint()
		doFailingInRuntime()
		// error: gas required exceeds allowance or always failing transaction
	}

	// create gap
	dapp.baseClient.nonceManager.NextNonce()
	for i := 0; i < 3; i++ {
		// use nonce with gap
		fmt.Println("--nonce with gap (failing)---")
		doXes()
	}

	dapp.baseClient.nonceManager.DebugForceIdle()
	// should be synced again
	{
		dapp.baseClient.nonceManager.DebugPrint()
		doXes()
	}

	dapp.baseClient.nonceManager.DebugForceIdle()
	dapp.baseClient.nonceManager.NextNonce()

	if !dapp.baseClient.nonceManager.DebugNonceEqualsNetwork() {
		t.Fatal("ended up de-synchronised!")
		// either our bug or node problem (see txpool issue, https://medium.com/kinblog/making-sense-of-ethereum-nonce-sense-3858d5588c64)
	}

	fmt.Println("success!")
}
