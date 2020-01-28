package account

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"encoding/binary"
	"sort"
	"time"

	"github.com/ProxeusApp/storage-app/dapp/core/account/pgpService"
	"github.com/ProxeusApp/storage-app/dapp/core/embdb"
)

type (
	Wallet struct {
		cfg           *Config
		accs          []*AccFile
		activeAcc     *Account
		pgpHandler    *pgpService.PGPServiceHandler
		lock          sync.RWMutex
		walletUsageDB *embdb.DB
	}
	AccFile struct {
		path       string
		acc        *Account
		lastAccess int64
	}
	TimestampSorter []*AccFile
)

func (a TimestampSorter) Len() int           { return len(a) }
func (a TimestampSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a TimestampSorter) Less(i, j int) bool { return a[i].lastAccess < a[j].lastAccess }

func LoadWallet(cfg *Config) (*Wallet, error) {
	if cfg == nil {
		return nil, os.ErrInvalid
	}
	if cfg.StorageDir == "" {
		cfg.StorageDir = "."
	}
	cfg.StorageDir = filepath.Join(cfg.StorageDir, "wallet")
	var err error
	wallet := &Wallet{accs: []*AccFile{}, cfg: cfg}
	cfg.pgpPublicKeyAddListener = wallet.addedListener
	wallet.pgpHandler, err = pgpService.NewPGPServiceHandler(cfg.StorageDir, cfg.PGPServiceURL, wallet.signerListener)
	if err != nil {
		return nil, err
	}

	wallet.walletUsageDB, err = embdb.Open(cfg.StorageDir, "wusage")
	if err != nil {
		return nil, err
	}

	count := 1
	wallet.accs = make([]*AccFile, 0)
	accIdMap := map[string]bool{}
	fileSuffix := strings.ToLower(cfg.FileSuffix)
	err = filepath.Walk(cfg.StorageDir, func(path string, f os.FileInfo, err error) error {
		if strings.HasSuffix(strings.ToLower(path), fileSuffix) {
			ac, err := NewFastReadETHAddrAndId(cfg, path) //to get the name and eth addr
			if err != nil {
				log.Println(err)
			}
			if ac != nil {
				if id := ac.GetETHAddress(); id != "" {
					if !accIdMap[id] {
						accIdMap[id] = true
						wallet.ensureAccountName(ac)
						count++
						log.Println("loading account", ac.GetETHAddress())
						accFile := &AccFile{path: path, acc: ac}
						wallet.accs = append(wallet.accs, accFile)
						return nil
					}
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	wallet.readAllAccountAccessTimestamps()
	return wallet, err
}

func (me *Wallet) readAllAccountAccessTimestamps() {
	go func() {
		for _, a := range me.accs {
			a.lastAccess = me.readAccountAccessTimestamp(a.acc)
		}
	}()
}

func (me *Wallet) updateAccountAccessTimestamp(acc *Account) int64 {
	now := time.Now().Unix()
	n := uint64(now)
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, n)
	me.walletUsageDB.Put(acc.kpETH.Pub, b)
	return now
}

func (me *Wallet) readAccountAccessTimestamp(acc *Account) int64 {
	b, _ := me.walletUsageDB.Get(acc.kpETH.Pub)
	if len(b) > 0 {
		return int64(binary.LittleEndian.Uint64(b))
	}
	return 0
}

func (me *Wallet) GetPGPClient() *pgpService.Client {
	return me.pgpHandler.GetClient()
}

func (me *Wallet) Import(r io.Reader, password string) (acc *Account, alreadyExists bool, err error) {
	acc, err = NewAccountImport(me.cfg, r)
	if err != nil {
		return nil, false, err
	}

	alreadyExistingAcc, exists := me.alreadyExists(acc)
	if exists {
		log.Println("[Wallet][Import] account already exists:", alreadyExistingAcc.GetETHAddress())
		return alreadyExistingAcc, true, nil
	}

	err = acc.VerifyCredentials(acc.GetFileName(), password)
	if err != nil {
		_ = os.Remove(acc.GetFilePath())
		return nil, false, err
	}

	me.ensureAccountName(acc)
	acf := &AccFile{path: acc.GetFileName(), acc: acc}
	me.accs = append(me.accs, acf)
	me.lock.Lock()
	me.setActiveAccount(acf)
	me.lock.Unlock()
	return acc, false, nil
}

func (me *Wallet) Export() (string, error) {
	me.lock.RLock()
	defer me.lock.RUnlock()
	ac := me.ActiveAccount()
	if ac != nil {
		return ac.GetFilePath(), nil
	}
	return "", os.ErrNotExist
}

func (me *Wallet) ExportAccount(ethAddr, pw string) (string, error) {
	if ethAddr == "" {
		return "", os.ErrInvalid
	}
	var accountToExport *AccFile
	var ac *AccFile
	for _, ac = range me.accs {
		if strings.ToLower(ac.acc.GetETHAddress()) == strings.ToLower(ethAddr) {
			accountToExport = ac
			break
		}
	}
	if accountToExport == nil || accountToExport.acc == nil {
		return "", errors.New("account not found")
	}
	err := accountToExport.acc.VerifyCredentials(accountToExport.path, pw)
	if len(pw) > 0 && err == nil {
		return accountToExport.acc.GetFilePath(), nil
	}
	return "", os.ErrPermission
}

func (me *Wallet) Remove(ethAddr, pw string) error {
	if ethAddr == "" {
		return os.ErrInvalid
	}
	var i int
	var accountToDelete *AccFile
	var ac *AccFile
	for i, ac = range me.accs {
		if strings.ToLower(ac.acc.GetETHAddress()) == strings.ToLower(ethAddr) {
			accountToDelete = ac
			break
		}
	}
	if accountToDelete == nil || accountToDelete.acc == nil {
		return errors.New("account not found")
	}
	err := accountToDelete.acc.Unlock(accountToDelete.path, pw)
	if len(pw) > 0 && err == nil {
		err = os.Remove(accountToDelete.acc.GetFilePath())
		if err != nil {
			return err
		}
		me.accs = append(me.accs[:i], me.accs[i+1:]...)
		return nil
	}
	return os.ErrPermission
}

func (me *Wallet) ensureAccountName(ac *Account) {
	if ac.GetName() == "" {
		ac.SetName(fmt.Sprintf("Account %s", ac.GetETHAddress()[:10]))
	}
}

func (me *Wallet) Login(ethAddr, pw string) error {
	return me.findAcc(ethAddr, func(acf *AccFile) error {
		err := acf.acc.Unlock(acf.path, pw)
		if len(pw) > 0 && err == nil {
			me.lock.Lock()
			me.setActiveAccount(acf)
			me.lock.Unlock()
			return nil
		}
		return err
	})
}

func (me *Wallet) setActiveAccount(acf *AccFile) {
	if me.activeAcc != nil && me.activeAcc != acf.acc {
		me.activeAcc.Close()
	}
	me.activeAcc = acf.acc
	acf.lastAccess = me.updateAccountAccessTimestamp(acf.acc)
}

func (me *Wallet) LoginWithNewAccount(name string, pw string) error {
	me.lock.RLock()
	defer me.lock.RUnlock()
	if pw == "" {
		return errors.New("password can not be empty")
	}
	a, err := NewAccount(me.cfg, pw)
	if err != nil {
		return err
	}
	name = strings.TrimSpace(name)
	if len(name) > 0 {
		a.SetName(name)
	}
	return me.storeAndProvideAName(a)
}

func (me *Wallet) LoginWithETHPriv(ethPriv, name, pw string) error {
	me.lock.RLock()
	defer me.lock.RUnlock()
	if pw == "" {
		return errors.New("password can not be empty")
	}
	a, err := NewAccountImportETHPrivCreatePGP(me.cfg, ethPriv, pw)
	if err != nil {
		return err
	}
	a.SetName(name)
	return me.storeAndProvideAName(a)
}

func (me *Wallet) LoginWithETHPrivAndPGPPriv(ethPriv, name, pw, pgpPriv, pgppw string) error {
	me.lock.RLock()
	defer me.lock.RUnlock()
	if pw == "" {
		return errors.New("password can not be empty")
	}
	a, err := NewAccountImportETHPrivAndPGP(me.cfg, ethPriv, pw, pgpPriv, pgppw)
	if err != nil {
		return err
	}
	a.SetName(name)
	return me.storeAndProvideAName(a)
}

func (me *Wallet) alreadyExists(a *Account) (existingAcc *Account, exists bool) {
	me.findAcc(a.GetETHAddress(), func(acf *AccFile) error {
		existingAcc = acf.acc
		exists = true
		return nil
	})
	return
}

var ErrAccAlreadyExists = errors.New("account already exists")

func (me *Wallet) storeAndProvideAName(a *Account) error {
	if _, exists := me.alreadyExists(a); exists {
		return ErrAccAlreadyExists
	}
	me.ensureAccountName(a)
	a.Store()
	acf := &AccFile{path: a.GetFileName(), acc: a}
	me.accs = append(me.accs, acf)
	me.setActiveAccount(acf)
	return nil
}

func (me *Wallet) HasActiveAndUnlockedAccount() bool {
	me.lock.RLock()
	defer me.lock.RUnlock()
	return me.activeAcc != nil && me.activeAcc.IsUnlocked()
}

func (me *Wallet) All() []*AccFile {
	sort.Sort(TimestampSorter(me.accs))
	return me.accs
}

func (me *Wallet) ActiveAccount() *Account {
	return me.activeAcc
}

func (me *Wallet) FindAccount(ethAddr string) (res *Account) {
	me.findAcc(ethAddr, func(acf *AccFile) error {
		res = acf.acc
		return nil
	})
	return
}

func (me *Wallet) addedListener(acc *Account) {
	me.pgpHandler.UpdatePGPPublicKey(acc.GetETHAddress(), acc.GetPGPPublicKey())
}

func (me *Wallet) signerListener(ethAddr string, challenge []byte) (sig []byte, err error) {
	me.findAcc(ethAddr, func(acf *AccFile) error {
		sig, err = acf.acc.SignWithETH(challenge)
		return err
	})
	return
}

func (me *Wallet) findAcc(ethAddr string, found func(acc *AccFile) error) error {
	for _, a := range me.accs {
		if strings.ToLower(a.acc.GetETHAddress()) == strings.ToLower(ethAddr) {
			return found(a)
		}
	}
	return os.ErrNotExist
}

func (me *Wallet) SignWithETHofActiveAccount(msg []byte) ([]byte, error) {
	me.lock.RLock()
	defer me.lock.RUnlock()
	activeAccount := me.getActiveAndUnlockedAccount()
	if activeAccount == nil {
		return []byte{}, ErrAccountLocked
	}
	return activeAccount.SignWithETH(msg)
}

func (me *Wallet) getActiveAndUnlockedAccount() *Account {
	me.lock.RLock()
	defer me.lock.RUnlock()
	if me.ActiveAccount() == nil || me.ActiveAccount().IsLocked() {
		return nil
	}
	return me.activeAcc
}

func (me *AccFile) Account() *Account {
	if me.acc != nil {
		return me.acc
	}
	return nil
}

func (me *Wallet) GetActiveAccountETHAddress() string {
	me.lock.RLock()
	defer me.lock.RUnlock()
	activeAccount := me.ActiveAccount()
	if activeAccount == nil {
		return ""
	}
	return activeAccount.GetETHAddress()
}

func (me *Wallet) GetActiveAccountETHPrivateKey() string {
	me.lock.RLock()
	defer me.lock.RUnlock()
	activeAccount := me.getActiveAndUnlockedAccount()
	if activeAccount == nil {
		return ""
	}
	return activeAccount.GetETHPriv()
}

func (me *Wallet) GetActiveAccountPGPKey() string {
	me.lock.RLock()
	defer me.lock.RUnlock()
	activeAccount := me.getActiveAndUnlockedAccount()
	if activeAccount == nil {
		return ""
	}
	return activeAccount.GetPGPPublicKey()
}

func (me *Wallet) GetActiveAccountName() string {
	me.lock.RLock()
	defer me.lock.RUnlock()
	activeAccount := me.getActiveAndUnlockedAccount()
	if activeAccount == nil {
		return ""
	}
	return activeAccount.GetName()
}

func (me *Wallet) GetActiveAccountETHCommonAddress() string {
	me.lock.RLock()
	defer me.lock.RUnlock()
	activeAccount := me.getActiveAndUnlockedAccount()
	if activeAccount == nil {
		return ""
	}
	return activeAccount.GetETHCommonAddress().String()
}

func (me *Wallet) GetActiveAccountPGPPrivatePw() []byte {
	me.lock.RLock()
	defer me.lock.RUnlock()
	activeAccount := me.getActiveAndUnlockedAccount()
	if activeAccount == nil {
		return nil
	}
	return activeAccount.GetPGPPrivatePw()
}

func (me *Wallet) GetActiveAccountPGPPrivateKey() []byte {
	me.lock.RLock()
	defer me.lock.RUnlock()
	activeAccount := me.getActiveAndUnlockedAccount()
	if activeAccount == nil {
		return nil
	}
	return activeAccount.GetPGPPrivateKey()
}

func (me *Wallet) Logout() error {
	me.lock.Lock()
	defer me.lock.Unlock()
	if me.activeAcc != nil {
		err := me.activeAcc.Close()
		me.activeAcc = nil
		return err
	}
	return nil
}

func (me *Wallet) Close() error {
	if me.pgpHandler != nil {
		me.pgpHandler.Close()
	}
	if me.activeAcc != nil {
		err := me.Logout()
		return err
	}
	if me.walletUsageDB != nil {
		me.walletUsageDB.Close()
	}
	return nil
}
