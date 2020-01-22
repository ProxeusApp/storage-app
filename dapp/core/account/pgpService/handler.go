package pgpService

import (
	"log"
	"os"
	"path/filepath"
	"sync"
)

type PGPServiceHandler struct {
	signer             func(ethAddr string, challenge []byte) ([]byte, error)
	pgpClient          *Client
	serviceInput       chan eth_pgp
	stopping           bool
	currentAccount     string //current eth account address
	currentAccountSync sync.Mutex
}

type eth_pgp struct {
	ethAddr      string
	pgpPublicKey string
}

func NewPGPServiceHandler(dbPath, url string, signer func(ethAddr string, challenge []byte) ([]byte, error)) (*PGPServiceHandler, error) {
	dbPath = filepath.Join(dbPath, "pgp")
	pgpClient, err := NewClient(url)
	if err != nil {
		return nil, err
	}
	if signer == nil {
		return nil, os.ErrInvalid
	}
	me := &PGPServiceHandler{signer: signer}
	me.pgpClient = pgpClient
	me.serviceInputHandler()
	return me, err
}

func (me *PGPServiceHandler) serviceInputHandler() {
	me.serviceInput = make(chan eth_pgp, 20)
	go func() {
		log.Println("PGP pks/add handler started...")
		defer func() {
			me.stopping = true
			log.Println("PGP pks/add handler stopped...")
		}()

		for {
			select {
			case keyPairs, ok := <-me.serviceInput:
				if !ok {
					return
				}
				me.add(keyPairs.ethAddr, keyPairs.pgpPublicKey)
			}
		}
	}()
}

func (me *PGPServiceHandler) GetClient() *Client {
	return me.pgpClient
}

func (me *PGPServiceHandler) UpdatePGPPublicKey(ethAddr, pgpPublicKey string) {
	defer func() {
		if r := recover(); r != nil { //can happen when closing
			log.Println("Recovered in UpdatePGPPublicKey", r)
		}
	}()
	if len(ethAddr) == 42 {
		me.serviceInput <- eth_pgp{ethAddr: ethAddr, pgpPublicKey: pgpPublicKey}
	}
}

func (me *PGPServiceHandler) add(ethAddr, pgpPublicKey string) {
	if me.stopping {
		return
	}
	if len(ethAddr) != 42 {
		return
	}
	keep := func() {
		log.Printf("PGP Handler: failed to add the public key of %s on %s, will try later\n", ethAddr, me.pgpClient.url)
	}
	registeredPGP, err := me.pgpClient.Lookup(ethAddr)
	if me.stopping {
		return
	}
	if err != nil || registeredPGP != pgpPublicKey {
		challenge, err := me.pgpClient.Challenge()
		if err != nil {
			keep()
			return
		}
		if me.stopping {
			return
		}
		sig, err := me.signer(ethAddr, []byte(challenge.Challenge))
		if err != nil {
			keep()
			return
		}
		if me.stopping {
			return
		}
		var ok bool
		ok, err = me.pgpClient.Add(pgpPublicKey, challenge.Token, string(sig))
		if err != nil || !ok {
			keep()
			return
		}
		if me.stopping {
			return
		}
		log.Printf("PGP Handler: added pgp public key for address %s to PGP public service %s\n", ethAddr, me.pgpClient.url)
		return
	}
	log.Printf("PGP Handler: public key already exists for address %s to PGP public service %s\n", ethAddr, me.pgpClient.url)
}

func (me *PGPServiceHandler) Close() {
	close(me.serviceInput)
}
