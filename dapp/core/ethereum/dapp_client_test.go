package ethereum

import (
	"fmt"
	"log"
	"math/big"
	"testing"

	"git.proxeus.com/core/central/dapp/core/util"

	"github.com/ethereum/go-ethereum/core/types"

	"github.com/ethereum/go-ethereum/common"

	"git.proxeus.com/core/central/spp/config"
)

//TODO impl ethereum.Client API tests

func TestNewClient(t *testing.T) {
	config.Setup()
	cfg := &config.Config

	log.Println(cfg)
	m, err := NewDappClient(cfg.EthClientURL, cfg.EthWebSocketURL, cfg.XESContractAddress, cfg.ContractAddress)

	if err != nil {
		panic(err)
	}
	m.InitListeners("", "0x531da9adba2099718c3c084c9af04319de8297c7", func(tx *PendingTx, txHash, status string) error {
		return nil
	}, func(lg *types.Log, recent bool) error {
		return nil
	})

	fi, err := m.FileInfo(util.StrHexToBytes32("0x5ab9d98c2ef825f3aafdacc7d6fa8296e75b387ff3f8a02d31884cd40d906cea"), true)
	if err != nil {
		log.Println("1 failed ", err)
	}
	log.Println(fi)
	fi, err = m.FileInfo(util.StrHexToBytes32("0x5ab9d98c2ef825f3aafdacc7d6fa8296e75b387ff3f8a02d31884cd40d906cea"), true)
	if err != nil {
		log.Println("2 failed ", err)
	}
	log.Println(fi)
	fi, err = m.FileInfo(util.StrHexToBytes32("0x5ab9d98c2ef825f3aafdacc7d6fa8296e75b387ff3f8a02d31884cd40d906cea"), true)
	if err != nil {
		log.Println("3 failed ", err)
	}
	log.Println(fi)
	fi, err = m.FileInfo(util.StrHexToBytes32("0x5ab9d98c2ef825f3aafdacc7d6fa8296e75b387ff3f8a02d31884cd40d906cea"), true)
	if err != nil {
		log.Println("4 failed ", err)
	}
	log.Println(fi)
}

func TestEthereumBugs(t *testing.T) {
	config.Setup()
	cfg := &config.Config
	c, err := NewDappClient(cfg.EthClientURL, cfg.EthWebSocketURL, cfg.XESContractAddress, cfg.ContractAddress)
	if err != nil {
		t.FailNow()
	}

	doXes := func() {
		tx, e := c.XESApproveToProxeusFS(
			"b8f30c0313f27fd73617d79257aec18c4521956fc0cad44618b0279b6843dda8",
			big.NewInt(-8446744073709551616))
		fmt.Println("tx error:", e)
		if e == nil {
			fmt.Println("tx hash:", tx.Hash().String())
		}
	}

	doFailingInRuntime := func() {
		tx, e := c.FileSetPerm(
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

	c.baseClient.nonceManager.OnAccountChange("0xa80899bb12e4afe9787425a5e5fe166234b88184")

	{
		c.baseClient.nonceManager.DebugPrint()
		doXes()
	}

	{
		// this case is not increasing nonce
		c.baseClient.nonceManager.DebugPrint()
		TestingModeBreakGas = true
		doXes()
		TestingModeBreakGas = false
	}

	{
		c.baseClient.nonceManager.DebugPrint()
		doXes()
	}

	{
		// this case is not increasing nonce
		c.baseClient.nonceManager.DebugPrint()
		doFailingInRuntime()
		// error: gas required exceeds allowance or always failing transaction
	}

	// create gap
	c.baseClient.nonceManager.NextNonce()
	for i := 0; i < 3; i++ {
		// use nonce with gap
		fmt.Println("--nonce with gap (failing)---")
		doXes()
	}

	c.baseClient.nonceManager.DebugForceIdle()
	// should be synced again
	{
		c.baseClient.nonceManager.DebugPrint()
		doXes()
	}

	if !c.baseClient.nonceManager.DebugNonceEqualsNetwork() {
		t.Fatal("ended up de-synchronised!")
		// either our bug or node problem (see txpool issue, https://medium.com/kinblog/making-sense-of-ethereum-nonce-sense-3858d5588c64)
	}

	fmt.Println("sucess!")
}
