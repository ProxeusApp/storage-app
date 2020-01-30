package account

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ProxeusApp/pgp"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ProxeusApp/storage-app/dapp/core/account/pgpService"
	"github.com/ProxeusApp/storage-app/dapp/core/embdb"
)

type (
	AddressBook struct {
		book             map[string]*AddressBookEntry
		rwLoadLock       sync.RWMutex
		db               embdb.DataStore
		pgpServiceClient *pgpService.Client
		stopchan         chan bool
	}
	NameSorter []AddressBookEntry
)

const AddressBookDBName = "addr_book"

func (a NameSorter) Len() int           { return len(a) }
func (a NameSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a NameSorter) Less(i, j int) bool { return a[i].Name < a[j].Name }

func NewAddressBook(addressBookStore embdb.DataStore, pgpServiceClient *pgpService.Client) (*AddressBook, error) {
	ab := &AddressBook{book: map[string]*AddressBookEntry{}, pgpServiceClient: pgpServiceClient}

	ab.db = addressBookStore

	ab.stopchan = make(chan bool)
	err := ab.ensureSyncRoutine()
	if err != nil {
		return nil, err
	}
	return ab, nil
}

func (me *AddressBook) ensureSyncRoutine() (err error) {
	ticker := time.NewTicker(time.Second * 30)
	go func() {
		me.rwLoadLock.Lock()
		all, err := me.db.All()
		if err == nil {
			for _, ethAddr := range all {
				ethAddress := string(ethAddr)
				abe, err := me.loadAddrBookEntry(ethAddress)
				if err == nil {
					me.book[ethAddress] = abe
				}
			}
		}
		me.rwLoadLock.Unlock()
		defer func() {
			ticker.Stop()
			close(me.stopchan)
		}()
		for {
			select {
			case <-ticker.C:
				me.syncValidatedWithPGPService()
			case <-me.stopchan:
				return
			}
		}
	}()
	return
}

func (me *AddressBook) List(ethAddr string) ([]AddressBookEntry, error) {
	if ethAddr == "" {
		return nil, os.ErrInvalid
	}
	ethAddr = strings.ToLower(ethAddr)
	me.rwLoadLock.RLock()
	addrBookEntries := make([]AddressBookEntry, 0)
	for _, abe := range me.book {
		if ethAddr == abe.ETHAddress || !abe.Hidden {
			addrBookEntries = append(addrBookEntries, AddressBookEntry{
				Name:                    abe.Name,
				ETHAddress:              abe.ETHAddress,
				PGPPublicKey:            abe.PGPPublicKey,
				ValidatedWithPGPService: abe.ValidatedWithPGPService,
			})
		}
	}
	me.rwLoadLock.RUnlock()
	sort.Sort(NameSorter(addrBookEntries))
	return addrBookEntries, nil
}

func (me *AddressBook) syncValidatedWithPGPService() {
	me.rwLoadLock.RLock()
	for _, abe := range me.book {
		me.syncValidatedWithPGPServiceEntry(abe)
	}
	me.rwLoadLock.RUnlock()
}

func (me *AddressBook) syncValidatedWithPGPServiceEntry(abe *AddressBookEntry) {
	if abe == nil {
		return
	}
	now := time.Now()
	if !abe.ValidatedWithPGPService && now.After(abe.lastPGPServiceCheck.Add(time.Minute*3)) {
		pgpPublicKey, err := me.pgpServiceClient.Lookup(abe.ETHAddress)
		if err == nil && len(pgpPublicKey) > 0 {
			if abe.PGPPublicKey == pgpPublicKey {
				log.Printf("PGP received the same public key from %s with address %s -> updating the validate flag", me.pgpServiceClient.GetURL(), abe.ETHAddress)
				me.rwLoadLock.RUnlock()
				me.rwLoadLock.Lock()
				abe.ValidatedWithPGPService = true
				//to ensure they are not synced at once to spread the pgp service load
				abe.lastPGPServiceCheck = time.Now().Add(time.Minute * time.Duration(me.rndBetween(20, 200)))
				me.rwLoadLock.Unlock()
				me.rwLoadLock.RLock()
			} else {
				if pgp.ValidatePublicKey([]byte(pgpPublicKey)) {
					log.Printf("PGP received a valid public key from %s with address %s -> adding it to our book", me.pgpServiceClient.GetURL(), abe.ETHAddress)
					me.rwLoadLock.RUnlock()
					me.rwLoadLock.Lock()
					abe.PGPPublicKey = pgpPublicKey
					abe.ValidatedWithPGPService = true
					//to ensure they are not synced at once to spread the pgp service load
					abe.lastPGPServiceCheck = time.Now().Add(time.Minute * time.Duration(me.rndBetween(60, 200)))
					me.rwLoadLock.Unlock()
					me.rwLoadLock.RLock()
				} else {
					log.Printf("PGP received an invalid public key from %s with address %s", me.pgpServiceClient.GetURL(), abe.ETHAddress)
				}
			}
		} else {
			log.Printf("PGP service request error: %s %s [%s] size of the PGP public key %d", me.pgpServiceClient.GetURL(), abe.ETHAddress, err, len(pgpPublicKey))
		}
	}
}

func (me *AddressBook) rndBetween(min, max int) int {
	return rand.Intn(max-min) + min
}

func (me *AddressBook) syncEntry(abe *AddressBookEntry) {
	if abe != nil && !abe.ValidatedWithPGPService {
		me.rwLoadLock.RLock()
		me.syncValidatedWithPGPServiceEntry(abe)
		me.rwLoadLock.RUnlock()
	}
}

func (me *AddressBook) QuickInsertByETHAddr(name, ethAddr, pgpPublicKey string) (*AddressBookEntry, error) {
	err := me.validateEntry(&name, &ethAddr, &pgpPublicKey)
	if err != nil {
		return nil, err
	}
	ethAddr = strings.ToLower(ethAddr)
	me.rwLoadLock.RLock()
	abe := me.book[ethAddr]
	me.rwLoadLock.RUnlock()
	if abe == nil {
		abe, err := me.loadAddrBookEntry(ethAddr)
		if err != nil {
			return nil, err
		}
		if abe == nil {
			abe = &AddressBookEntry{
				Name:         name,
				ETHAddress:   ethAddr,
				PGPPublicKey: pgpPublicKey,
				Hidden:       true,
			}
			me.rwLoadLock.Lock()
			me.book[ethAddr] = abe
			me.rwLoadLock.Unlock()
			me.insertAddrBookEntry(abe)
		} else {
			//repair PGP
			if len(pgpPublicKey) > 0 && pgpPublicKey != abe.PGPPublicKey {
				abe.PGPPublicKey = pgpPublicKey
				abe.ValidatedWithPGPService = false
				me.insertOrUpdateAddrBookEntry(abe)
			}
		}
	}
	return abe, nil
}

func (me *AddressBook) Create(name, ethAddr string) (*AddressBookEntry, error) {
	err := me.validateEntry(&name, &ethAddr, nil)
	if err != nil {
		return nil, err
	}
	pgpPublicKey := ""

	if me.pgpServiceClient != nil {
		pgpPublicKey, err = me.pgpServiceClient.Lookup(ethAddr)
		if err != nil {
			pgpPublicKey = ""
		}
	}
	abe := NewAddressBookEntry(name, ethAddr, pgpPublicKey)
	err = me.insertAddrBookEntry(abe)
	return abe, err // Fails if already existing
}

// Updates an AddressBookEntry
func (me *AddressBook) Update(name, ethAddr, pgpPublicKey string) (*AddressBookEntry, error) {
	err := me.validateEntry(&name, &ethAddr, &pgpPublicKey)
	if err != nil {
		return nil, err
	}
	ethAddr = strings.ToLower(ethAddr)
	abe := me.provideAddrBookEntry(ethAddr)
	me.rwLoadLock.Lock()
	abe.Name = name
	if len(pgpPublicKey) > 0 {
		abe.PGPPublicKey = pgpPublicKey
		abe.ValidatedWithPGPService = false
	}
	abe.Hidden = false
	me.rwLoadLock.Unlock()
	return abe, me.updateAddrBookEntry(abe) // Updates if existing
}

func (me *AddressBook) Hide(ethAddr string) (err error) {
	addr := strings.ToLower(ethAddr)
	me.rwLoadLock.RLock()
	abe := me.book[addr]
	me.rwLoadLock.RUnlock()
	if abe == nil {
		abe, err = me.loadAddrBookEntry(ethAddr)
		if err != nil {
			return err
		}
	}
	abe.Hidden = true
	me.updateAddrBookEntry(abe)
	return nil
}

// Returns a stored AddressBookEntry. The difference with Get() is that it doesn't create one if non-existing
func (me *AddressBook) Stored(ethAddr string) (*AddressBookEntry, error) {
	return me.loadAddrBookEntry(ethAddr)
}

// Returns a stored AddressBookEntry or an AddressBookEntry synchronized with the PGP service
func (me *AddressBook) Get(ethAddr string) *AddressBookEntry {
	if me.isInvalidETHAddr(ethAddr) {
		return nil
	}
	return me.provideAddrBookEntry(strings.ToLower(ethAddr))
}

func (me *AddressBook) validateEntry(name, ethAddr, pgpPublicKey *string) error {
	if *name == "" {
		return errors.New("name: invalid")
	}
	if me.isInvalidETHAddr(*ethAddr) {
		return errors.New("ethAddress: invalid")
	}
	if pgpPublicKey != nil && len(*pgpPublicKey) > 0 {
		if !pgp.ValidatePublicKey([]byte(*pgpPublicKey)) {
			return errors.New("pgpPublicKey: invalid")
		}
	}
	return nil
}

func (me *AddressBook) isInvalidETHAddr(addr string) bool {
	return len(addr) < 40 || len(addr) > 42 && !common.IsHexAddress(addr) || me.IsEmptyAddr(addr)
}

func (me *AddressBook) IsEmptyAddr(addr string) bool {
	return addr == "0x0000000000000000000000000000000000000000"
}

/*
	Provides an AddressBookEntry by searching in application memory first,
	then tries to retrieve it from the DB and as last step creates a new one which is then synced with PGP service
*/
func (me *AddressBook) provideAddrBookEntry(ethAddr string) *AddressBookEntry {
	me.rwLoadLock.RLock()
	abe := me.book[ethAddr]
	me.rwLoadLock.RUnlock()
	if abe == nil {
		var err error
		abe, err = me.loadAddrBookEntry(ethAddr)
		if err != nil {
			log.Println(err)
		}
		if abe == nil {
			abe = &AddressBookEntry{
				ETHAddress: ethAddr,
			}
		} else {
			me.rwLoadLock.Lock()
			me.book[ethAddr] = abe
			me.rwLoadLock.Unlock()
		}
	}
	if me.pgpServiceClient != nil {
		me.syncEntry(abe)
	}
	return abe
}

// Returns AddressBookEntry, even hidden ones! Returns nil if none has been found
func (me *AddressBook) loadAddrBookEntry(ethAddr string) (*AddressBookEntry, error) {
	abe := &AddressBookEntry{}
	bts, err := me.db.Get([]byte(ethAddr))
	if err != nil {
		return abe, err
	}
	if len(bts) > 0 {
		err = json.Unmarshal(bts, abe)
		return abe, err
	}
	return nil, nil
}

// Only returns non-hidden AddressBookEntry
func (me *AddressBook) loadVisibleAddrBookEntry(ethAddr string) (*AddressBookEntry, error) {
	abe, err := me.loadAddrBookEntry(ethAddr)
	if abe != nil && abe.Hidden == true {
		return nil, err
	}
	return abe, err
}

func (me *AddressBook) insertOrUpdateAddrBookEntry(abe *AddressBookEntry) error {
	me.rwLoadLock.Lock()
	me.book[abe.ETHAddress] = abe
	me.rwLoadLock.Unlock()
	bts, err := json.Marshal(abe)
	if err == nil {
		return me.db.Put([]byte(abe.ETHAddress), bts)
	}
	return err
}

func (me *AddressBook) updateAddrBookEntry(abe *AddressBookEntry) error {
	me.rwLoadLock.RLock()
	defer me.rwLoadLock.RUnlock()
	existingAbe, err := me.loadAddrBookEntry(abe.ETHAddress)
	if err != nil {
		return err
	}
	if existingAbe == nil {
		return new(ErrAddressNotFound)
	}
	bts, err := json.Marshal(abe)
	if err == nil {
		return me.db.Put([]byte(abe.ETHAddress), bts)
	}
	return err
}

// Inserts a new AddressBookEntry and returns an error if it exists already and isnt hidden.
func (me *AddressBook) insertAddrBookEntry(abe *AddressBookEntry) error {
	foundAbe, err := me.loadAddrBookEntry(abe.ETHAddress)
	if err != nil {
		return err
	}
	if foundAbe != nil && foundAbe.Hidden == false {
		return new(ErrAddressExists)
	}
	return me.insertOrUpdateAddrBookEntry(abe)
}

func (me *AddressBook) Close() error {
	me.stopchan <- true
	<-me.stopchan
	for _, abe := range me.book {
		me.insertOrUpdateAddrBookEntry(abe)
	}
	me.db.Close()
	return nil
}
