package fs

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"github.com/ProxeusApp/storage-app/dapp/core/ethglue"

	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"

	"github.com/ProxeusApp/storage-app/spp/config"
	"github.com/ProxeusApp/storage-app/spp/fs/db"
)

type ProxeusFSDeleteEvent struct {
	basePath        string
	database        *db.KVStore
	contractAddress common.Address
	stopchan        chan bool
	ethWebSocketURL string
	client          *ethclient.Client
}

func NewProxeusFSDeleteEvent(cfg *config.Configuration) (*ProxeusFSDeleteEvent, error) {
	pfsde := &ProxeusFSDeleteEvent{ethWebSocketURL: cfg.EthWebSocketURL, contractAddress: common.HexToAddress(cfg.ContractAddress)}
	pfsde.basePath = cfg.StorageDir
	strDir, err := filepath.Abs(cfg.StorageDir)
	if err != nil {
		return nil, err
	}
	pfsde.database, err = db.NewKVStore(filepath.Join(strDir, "PsppDatabase.db"))
	if err != nil {
		return nil, err
	}
	pfsde.stopchan = make(chan bool)
	return pfsde, nil
}

func (me *ProxeusFSDeleteEvent) removeFileFromDisk(filename string) (err error) {
	absPath := filepath.Join(me.basePath, filepath.Base(filename))
	err = os.Remove(absPath)
	if err == nil {
		log.Println("Removed file: ", absPath)
	}
	return
}

func (me *ProxeusFSDeleteEvent) StartSubscribeDeleteHandler() {
	me.startSubscribeDeleteHandler()
}

func (me *ProxeusFSDeleteEvent) deletedEventsQuery() ethereum.FilterQuery {
	eventSignature := []byte("Deleted(bytes32)")
	esHash := crypto.Keccak256Hash(eventSignature)
	return ethereum.FilterQuery{
		Addresses: []common.Address{me.contractAddress},
		Topics: [][]common.Hash{
			{esHash},
		},
	}
}

func (me *ProxeusFSDeleteEvent) startSubscribeDeleteHandler() {
	var (
		err error
		sub ethereum.Subscription
	)

	//TODO add to config file the url
	me.client, err = ethglue.Dial(me.ethWebSocketURL)
	if err != nil {
		panic(err)
	}

	q := me.deletedEventsQuery()
	/* Lookup last block number from which we deleted files, so we can query only from that block and onwards
	for delete events that we still have not processed, maybe because the spp was offline.*/
	var blockNumber uint64
	err = me.database.Get("vLog-blockNumber", &blockNumber)
	if err == nil && blockNumber > 0 {
		q.FromBlock = big.NewInt(int64(blockNumber))
	}

	logList, err := me.client.FilterLogs(context.Background(), q)
	if err != nil {
		panic(err)
	}

	// async loops
	go func() {
		var logs chan types.Log
		defer func() {
			//Close
			log.Println("stopping SubscribeDeleteHandler")
			sub.Unsubscribe()
			close(logs)
			me.client.Close()
			me.database.Close(false)
			close(me.stopchan)
		}()

		for {
			sub, logs, err = me.safelyEstablishWSConnection()
			if err != nil {
				log.Println("establishing connection for bc events again")
				time.Sleep(time.Second * 4)
				continue
			}
			break
		}

		for _, vLog := range logList {
			h := vLog.Topics[1].Hex()
			me.removeFileFromDisk(h)
		}

		for {
			select {
			case err, ok := <-sub.Err():
				if !ok {
					return
				}
				for {
					sub, logs, err = me.safelyEstablishWSConnection()
					if err != nil {
						log.Println("establishing connection for bc events again")
						time.Sleep(time.Second * 4)
						continue
					}
					break
				}
			case vLog, ok := <-logs:
				if !ok {
					return
				}
				if len(vLog.Topics) > 1 {
					h := vLog.Topics[1].Hex()
					err := me.removeFileFromDisk(h)
					if err != nil {
						log.Println("fsDelete ERROR rm", err)
					}
					// Store last vLog from Delete events so we can reference the last block number for which we deleted files
					err = me.database.Put("vLog-blockNumber", &vLog.BlockNumber)
					if err != nil {
						log.Println("fsDelete ERROR put", err)
					}
				}

			case <-me.stopchan:
				return
			}
		}
	}()
}

func (me *ProxeusFSDeleteEvent) safelyEstablishWSConnection() (sub ethereum.Subscription, logs chan types.Log, err error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
			err = os.ErrClosed
		}
		if err != nil {
			close(logs)
		}
	}()
	logs = make(chan types.Log)
	me.client, err = ethglue.Dial(me.ethWebSocketURL)
	if err != nil {
		return
	}
	query := me.deletedEventsQuery()
	for {
		sub, err = me.client.SubscribeFilterLogs(context.TODO(), query, logs)
		if err != nil {
			log.Println("failed to subscribe for bc events")
			time.Sleep(time.Second * 4)
			continue
		}
		break
	}
	return
}

func (me *ProxeusFSDeleteEvent) Close() {
	me.stopchan <- true
	<-me.stopchan
}
