package blockchain

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ProxeusApp/storage-app/lib/blockchain/bctypes"
	"github.com/ProxeusApp/storage-app/lib/blockchain/ethereum"
	contract2 "github.com/ProxeusApp/storage-app/lib/blockchain/generator/ethereum/go"
)

func TestEndToEndDeploy(t *testing.T) {
	t.SkipNow()
	const defaultEthURL = "https://ropsten.infura.io/v3/"
	const defaultEthwsURL = "wss://ropsten.infura.io/ws"
	const testwallet = `{"address":"f8ae553b87f695a938236d22c48a0c5b43edddbe",
	"crypto":{
		"cipher":"aes-128-ctr",
		"ciphertext":"2672f1b0529a3e9e93027dc3bc587b44feff0888cdb160039a37d6489fc90817",
		"cipherparams":{"iv":"64f4b93669aa468192e24f8f42e08b01"},
		"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"52a1435e7c39084958665ea459db52953f533f1ba96fb861570bf9557a5a120c"},
		"mac":"70abe56da792f13bfc9983f65ef63b6fcccc156095a1ca04609a4203dd1a6578"
	},"id":"91fa36cf-59c8-4410-9513-28b5ef86ee59","version":3}`

	ethclient := ethereum.NewETHClient(defaultEthURL, defaultEthwsURL)

	//Account
	bufReader := ioutil.NopCloser(bytes.NewReader([]byte(testwallet)))
	account, err := bctypes.LoadFromWallet(bufReader, "test")
	if err != nil {
		fmt.Println("Error opening Account: ")
		t.Fatal(err)
	}
	fmt.Print("Account loaded: ")
	fmt.Println(account.GetAddress())

	deploy := false

	var testcontract contract2.TestcontractContract
	if deploy {
		fmt.Println("Deploying Contract...")
		transaction, err := testcontract.Deploy(ethclient, account)
		if err != nil {
			fmt.Println("Error in deploy transaction: ")
			t.Fatal(err)
		}
		fmt.Print("Tx Hash: ")
		fmt.Println(transaction.Hash.String())

		//Wait for contract = deployed to get address
		receiptTicker := time.NewTicker(3 * time.Second)
		var address common.Address
		defer receiptTicker.Stop()
		fmt.Print("Waiting for tx to be mined")
		for {
			if transaction != nil {
				mined := transaction.Mined
				if mined {
					address = testcontract.Address()
					break
				}
			}
			fmt.Print(".")
			<-receiptTicker.C
		}
		fmt.Print("Deployed at: ")
		fmt.Println(address.String())
	} else {
		address := "C6BAdfc9722E3CB8395622472a257651420dA8b5"
		err := testcontract.BindContract(address, ethclient)
		if err != nil {
			fmt.Println("Error binding contract: ")
			t.Fatal(err)
		}
		fmt.Print("Contract bound from: ")
		fmt.Println(testcontract.Address().String())
	}

	//Setup Event Listener
	events := make(chan bctypes.Event)
	stop := make(chan bool)

	listener := false
	//go func to listen and print event and abort when received
	go func() {
		fmt.Println("Listing for Events...")
		for {
			event := <-events
			fmt.Print(">Event: ")
			fmt.Println(testcontract.UnpackTestEvent(event))
		}
	}()
	if listener || deploy {
		fmt.Println("Setting up EventListener...")
		err = testcontract.SetupEventListener(ethclient, 0, events, stop)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		fmt.Println("Setting up EventListener...")
		err = testcontract.FetchEvents(ethclient, 0, events)
		if err != nil {
			t.Fatal(err)
		}
	}

	//Do Transaction on Contract
	fmt.Println("Calling Test Transaction...")
	calltransaction, err := testcontract.Testfunction(ethclient, account, "Test 2")
	if err != nil {
		fmt.Println("Error in transaction: ")
		t.Fatal(err)
	}
	fmt.Print("Tx Hash: ")
	fmt.Println(calltransaction.Hash.String())

	//Wait for transaction = mined to get result
	functionTicker := time.NewTicker(3 * time.Second)
	defer functionTicker.Stop()
	fmt.Print("Waiting for tx to be mined")
	for {
		mined := calltransaction.Mined
		if mined {
			break
		}
		fmt.Print(".")
		<-functionTicker.C
	}
	fmt.Print("Gas used: ")
	fmt.Println(calltransaction.Result.Receipt.GasUsed)
}
