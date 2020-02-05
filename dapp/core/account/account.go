package account

import (
	"bytes"
	"crypto/ecdsa"
	b64 "encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ProxeusApp/pgp"

	"github.com/ethereum/go-ethereum/accounts/keystore"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/pborman/uuid"
)

type (
	Config struct {
		StorageDir              string
		PGPServiceURL           string
		FileSuffix              string
		pgpPublicKeyAddListener func(ac *Account)
	}
	Account struct {
		name        string
		unlocked    bool
		key         *keystore.Key
		cfg         *Config
		ethAddr     *common.Address
		kpPGP       KeyPair
		kpETH       KeyPair
		localdataDB string
	}
	KeyPair struct {
		Pub    []byte
		Priv   []byte
		PrivPw []byte
	}
)

func NewAccount(cfg *Config, pw string) (me *Account, err error) {
	me = &Account{cfg: cfg}
	bpw := []byte(pw)
	err = me.newETH(bpw)
	if err != nil {
		return
	}
	err = me.newPGP(me.kpETH.Pub, bpw)
	if err != nil {
		return nil, err
	}
	me.unlockAndPGPServiceInsert()
	return
}

func NewFastReadETHAddrAndId(cfg *Config, name string) (*Account, error) {
	me := &Account{cfg: cfg}
	p := filepath.Join(me.cfg.StorageDir, filepath.Base(name))
	b64File, err := os.OpenFile(p, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	fcontent, err := ioutil.ReadAll(b64File)
	if err != nil {
		return nil, err
	}
	b64File.Close()
	_, err = me.parseImport(fcontent)
	return me, err
}

func NewAccountFromDisk(cfg *Config, name, pw string) (*Account, error) {
	me := &Account{cfg: cfg}
	return me, me.Unlock(name, pw)
}

func (me *Account) Unlock(name string, pw string) error {
	p := filepath.Join(me.cfg.StorageDir, filepath.Base(name))
	b64File, err := os.OpenFile(p, os.O_RDONLY, 0600)
	if err != nil {
		return err
	}
	fcontent, err := ioutil.ReadAll(b64File)
	if err != nil {
		return err
	}
	b64File.Close()
	ks, err := me.parseImport(fcontent)
	if err != nil {
		return err
	}

	err = me.unlock(ks, []byte(pw))
	if len(pw) > 0 {
		return err
	}
	return nil
}

func (me *Account) VerifyCredentials(name string, pw string) error {
	p := filepath.Join(me.cfg.StorageDir, filepath.Base(name))
	b64File, err := os.OpenFile(p, os.O_RDONLY, 0600)
	if err != nil {
		return err
	}
	fcontent, err := ioutil.ReadAll(b64File)
	if err != nil {
		return err
	}
	b64File.Close()
	ks, err := me.parseImport(fcontent)
	if err != nil {
		return err
	}
	ksBta, err := json.Marshal(ks.ETHKeystore())
	if err != nil {
		return err
	}
	return me.verifyCredentials(ksBta, []byte(pw))
}

func NewAccountImport(cfg *Config, r io.Reader) (*Account, error) {
	me := &Account{cfg: cfg}
	bts, err := ioutil.ReadAll(r)
	_, err = me.parseImport(bts)
	if err != nil {
		return nil, err
	}
	if err = me.ensureStoreDir(); err != nil {
		return nil, err
	}
	fw, err := os.OpenFile(filepath.Join(me.cfg.StorageDir, me.GetFileName()), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	_, err = fw.Write(bts)
	if err != nil {
		return nil, err
	}
	err = fw.Close()
	return me, err
}

func NewAccountImportETHAndPGP(cfg *Config, pw, content []byte) (*Account, error) {
	me := &Account{cfg: cfg}
	ks, err := me.parseImport(content)
	if err != nil {
		return nil, err
	}
	return me, me.unlock(ks, pw)
}

func NewAccountImportETHPrivCreatePGP(cfg *Config, ethPriv, pw string) (*Account, error) {
	me := &Account{cfg: cfg}
	me.kpETH.PrivPw = []byte(pw)
	err := me.importETHPriv(ethPriv)
	if err != nil {
		return nil, err
	}
	err = me.newPGP(me.kpETH.Pub, me.kpETH.PrivPw)
	if err != nil {
		return nil, err
	}
	me.unlockAndPGPServiceInsert()
	return me, nil
}

func NewAccountImportETHPrivAndPGP(cfg *Config, ethPriv, pw, pgpPriv, pgpPw string) (*Account, error) {
	me := &Account{cfg: cfg}
	me.kpETH.PrivPw = []byte(pw)
	err := me.importETHPriv(ethPriv)
	if err != nil {
		return nil, err
	}
	err = me.importPGP(pgpPw, pgpPriv)
	if err != nil {
		return nil, err
	}
	me.unlockAndPGPServiceInsert()
	return me, nil
}

func (me *Account) GetName() string {
	return me.name
}

func (me *Account) Id() string {
	if me.key != nil && len(me.key.Id) > 0 {
		return string(me.key.Id)
	}
	return me.GetETHAddress()
}

func (me *Account) GetFileName() string {
	if me.cfg != nil {
		return me.GetETHAddress() + me.cfg.FileSuffix
	}
	return ""
}

func (me *Account) GetFilePath() string {
	return filepath.Join(me.cfg.StorageDir, me.GetFileName())
}

func (me *Account) SetName(n string) {
	me.name = n
}

func (me *Account) GetETHAddress() string {
	return string(me.kpETH.Pub)
}

func (me *Account) GetETHCommonAddress() *common.Address {
	if me.ethAddr == nil {
		if len(me.kpETH.Pub) > 0 {
			addr := common.HexToAddress(me.GetETHAddress())
			me.ethAddr = &addr
		}
	}
	return me.ethAddr
}

func (me *Account) GetETHPriv() string {
	return string(me.kpETH.Priv)
}

func (me *Account) GetPGPPublicKey() string {
	return string(me.kpPGP.Pub)
}

func (me *Account) GetPGPPrivateKey() []byte {
	return me.kpPGP.Priv
}

func (me *Account) GetPGPPrivatePw() []byte {
	return me.kpPGP.PrivPw
}

func (me *Account) unlock(ks *ProxeusKeystore, pw []byte) (err error) {
	defer func() {
		if err == nil {
			me.unlockAndPGPServiceInsert()
		}
	}()
	pws := string(pw)
	var unlockedKey *keystore.Key

	ethKs, err := json.Marshal(ks.ETHKeystore())
	if err != nil {
		return
	}
	unlockedKey, err = keystore.DecryptKey(ethKs, pws)
	if err != nil {
		return
	}
	me.kpETH.PrivPw = pw
	me.key = unlockedKey
	me.impETH(unlockedKey.PrivateKey)
	if len(ks.PrivateKey) > 0 {
		err = me.importPGP(pws, ks.PrivateKey)
		return
	}
	err = me.newPGP(me.kpETH.Pub, pw)
	return
}

func (me *Account) verifyCredentials(ks, pw []byte) (err error) {
	pws := string(pw)
	_, err = keystore.DecryptKey(ks, pws)
	return err
}

func (me *Account) parseImport(content []byte) (pks *ProxeusKeystore, err error) {
	pks = &ProxeusKeystore{}
	err = pks.Load(content)
	if err != nil {
		return nil, err
	}
	me.name = pks.Name
	me.readETHPubFromKeystore(pks)
	me.readPGPPublicFromKeystore(pks)
	return pks, err
}

func (me *Account) readETHPubFromKeystore(ks *ProxeusKeystore) error {
	if me.key == nil {
		me.key = &keystore.Key{}
	}
	if len(ks.ETHKeystore().ID) > 0 {
		me.key.Id = []byte(uuid.Parse(ks.ETHKeystore().ID))
	}
	addrStr := ks.AddressWithPrefix()
	if len(addrStr) == 42 {
		me.kpETH.Pub = []byte(strings.ToLower(addrStr))
		me.key.Address = common.BytesToAddress(me.kpETH.Pub)
	}
	return os.ErrNotExist
}

func (me *Account) readPGPPublicFromKeystore(ks *ProxeusKeystore) {
	me.kpPGP.Pub = []byte(ks.PublicKey)
	me.kpPGP.PrivPw = []byte(ks.PgpPw)

	if len(me.kpPGP.Pub) == 0 {
		if ks.PgpKeys != nil {
			for k, v := range ks.PgpKeys {
				if len(k) > 0 {
					if string(me.kpETH.Pub) == strings.ToLower(k) {
						if v != nil {
							if ppgm, ok := v.(map[string]interface{}); ok {
								if pk, ok := ppgm["publicKey"].(string); ok {
									me.kpPGP.Pub = []byte(pk)
								}
							}
						}
					}
				}
			}
		}
	}
}

func (me *Account) SignWithETH(msg []byte) ([]byte, error) {
	if !me.unlocked {
		return nil, ErrAccountLocked
	}
	var (
		challenge []byte
		err       error
	)

	ecdsaPriv, err := crypto.HexToECDSA(string(me.kpETH.Priv))
	if err != nil {
		return nil, err
	}

	challenge, err = me.decodeHex(msg)
	if err != nil {
		return nil, err
	}

	challenge = me.ethSignPrependMsgAndHash(challenge)

	signature, err := crypto.Sign(challenge, ecdsaPriv)
	if err != nil {
		return nil, err
	}
	buf := new(bytes.Buffer)
	buf.Write(signature[:64])
	buf.WriteByte(byte(signature[64]) + 27) // Yes add 27, weird Ethereum quirk
	return []byte("0x" + hex.EncodeToString(buf.Bytes())), err
}

func (me *Account) VerifyWithETH(msg, sig []byte) (bool, error) {
	var (
		challenge []byte
		signature []byte
		err       error
	)
	if len(sig) == 0 {
		return false, os.ErrInvalid
	}
	signature, err = me.decodeHex(sig)
	if err != nil {
		return false, err
	}
	if len(sig) != 65 {
		return false, os.ErrInvalid
	}
	// need to subtract 27 from V, reason here: https://github.com/ethereum/wiki/wiki/JavaScript-API#returns-45
	signature[64] -= 27

	challenge, err = me.decodeHex(msg)
	if err != nil {
		return false, err
	}

	challenge = me.ethSignPrependMsgAndHash(challenge)
	// get the public key from the signature and challenge hash
	pubKeyECDSA, err := crypto.SigToPub(challenge, signature)

	// if any of the decodes raised errors, return it
	if err != nil {
		return false, err
	}
	return bytes.Equal([]byte(crypto.PubkeyToAddress(*pubKeyECDSA).Hex()), me.kpETH.Pub), err
}

func (me *Account) ethSignPrependMsgAndHash(challenge []byte) []byte {
	// pre-pend the eth_sign RPC prefix https://github.com/ethereum/wiki/wiki/JSON-RPC#eth_sign
	challenge = append([]byte("\x19Ethereum Signed Message:\n"+strconv.Itoa(len(challenge))), challenge...)
	return crypto.Keccak256(challenge)
}

func (me *Account) decodeHex(msg []byte) (res []byte, err error) {
	if bytes.HasPrefix(msg, []byte("0x")) {
		res, err = hex.DecodeString(string(msg[2:]))
	} else {
		res, err = hex.DecodeString(string(msg))
	}
	return
}

var ErrAccountLocked = errors.New("account locked")

func (me *Account) SignWithPGP(msg []byte) ([]byte, error) {
	if !me.unlocked {
		return nil, ErrAccountLocked
	}
	sig, err := pgp.Sign(msg, me.kpPGP.PrivPw, [][]byte{me.kpPGP.Priv})
	if err != nil {
		return nil, err
	}
	return sig, nil
}

func (me *Account) VerifyWithPGP(msg, sig []byte) (bool, error) {
	return pgp.Verify(msg, sig, [][]byte{me.kpPGP.Pub})
}

func (me *Account) importETHPriv(p string) error {
	if strings.HasPrefix(p, "0x") {
		p = p[2:]
	}
	privateKey, err := crypto.HexToECDSA(p)
	if err != nil {
		return err
	}
	me.impETH(privateKey)
	return nil
}

func (me *Account) newETH(pw []byte) error {
	me.kpETH.PrivPw = pw
	k, err := crypto.GenerateKey()
	me.impETH(k)
	return err
}

func (me *Account) impETH(key *ecdsa.PrivateKey) {
	me.kpETH.Pub = []byte(strings.ToLower(crypto.PubkeyToAddress(key.PublicKey).Hex()))
	me.kpETH.Priv = []byte(hex.EncodeToString(key.D.Bytes()))
}

func (me *Account) newPGP(key, pw []byte) error {
	stryKey := string(key)
	p, err := pgp.Create(stryKey, stryKey, 4096, 0)
	if err != nil {
		return err
	}
	me.kpPGP.Pub = p["public"]
	me.kpPGP.Priv = p["private"]
	me.kpPGP.PrivPw = pw
	return err
}

func (me *Account) importPGP(pw, priv string) error {
	me.kpPGP.Priv = []byte(priv)
	if len(me.kpPGP.PrivPw) == 0 {
		me.kpPGP.PrivPw = []byte(pw)
	}
	strKey := string(me.kpETH.Pub)
	m, err := pgp.ReadIdentity([][]byte{me.kpPGP.Priv})
	if err != nil {
		return err
	}
	if len(m) > 0 { // same identity
		if strings.ToLower(m[0]["name"]) == strings.ToLower(strKey) {
			me.kpPGP.Pub, err = pgp.ReadPublicKey([]byte(pw), me.kpPGP.Priv)
			if err != nil {
				return err
			}
			log.Printf("import PGP with same identity %s\n", strKey)
			return nil
		}
	}
	// other identity
	kpair, err := pgp.WriteIdentity([]byte(pw), me.kpPGP.Priv, strKey, "", strKey)
	if err != nil {
		return err
	}
	log.Printf("import PGP with new identity %s\n", strKey)
	me.kpPGP.Pub = kpair["public"]   //PGP with new identity
	me.kpPGP.Priv = kpair["private"] //PGP with new identity
	return err
}

func (me *Account) unlockAndPGPServiceInsert() {
	me.unlocked = true
	if me.cfg != nil && me.cfg.pgpPublicKeyAddListener != nil {
		me.cfg.pgpPublicKeyAddListener(me)
	}
}

func (me *Account) IsUnlocked() bool {
	return me.unlocked
}
func (me *Account) IsLocked() bool {
	return !me.unlocked
}

func (me *Account) ExportEncrypted(w io.Writer) error {
	log.Println(string(me.kpETH.Pub), string(me.kpETH.Priv))
	ecdsaPriv, err := crypto.HexToECDSA(string(me.kpETH.Priv))
	if err != nil {
		return err
	}
	if me.key == nil {
		me.key = &keystore.Key{PrivateKey: ecdsaPriv, Address: crypto.PubkeyToAddress(ecdsaPriv.PublicKey)}
	}
	if me.key.PrivateKey == nil {
		me.key.Address = crypto.PubkeyToAddress(ecdsaPriv.PublicKey)
		me.key.PrivateKey = ecdsaPriv
	}
	if len(me.key.Id) == 0 {
		me.key.Id = uuid.NewRandom()
	}

	ksBts, err := keystore.EncryptKey(me.key, string(me.kpETH.PrivPw), keystore.StandardScryptN, keystore.StandardScryptP)
	if err != nil {
		return err
	}

	keystores := &ProxeusKeystore{}
	keystores.Name = me.name
	keystores.PrivateKey = string(me.kpPGP.Priv)
	keystores.PublicKey = string(me.kpPGP.Pub)
	if !bytes.Equal(me.kpPGP.PrivPw, me.kpETH.PrivPw) {
		keystores.PgpPw = string(me.kpPGP.PrivPw)
	}
	err = keystores.SetETHKeystore(ksBts)
	if err != nil {
		return err
	}

	bts, err := json.Marshal(keystores)
	if err != nil {
		return err
	}
	uEnc := b64.StdEncoding.EncodeToString(bts)
	_, err = w.Write([]byte(uEnc))
	return err
}

func (me *Account) Export(pw []byte) (map[string]interface{}, error) {
	if bytes.Equal(me.kpETH.PrivPw, pw) {
		return map[string]interface{}{
			"name": me.name,
			"eth": map[string]interface{}{
				"pub":  string(me.kpETH.Pub),
				"priv": string(me.kpETH.Priv),
			},
			"pgp": map[string]interface{}{
				"pub":  string(me.kpPGP.Pub),
				"priv": string(me.kpPGP.Priv),
			},
		}, nil
	}
	return nil, os.ErrPermission
}

func (me *Account) Store() error {
	return me.store()
}

func (me *Account) store() (err error) {
	if len(me.kpETH.PrivPw) > 0 && len(me.kpPGP.Pub) > 0 && len(me.kpPGP.Priv) > 0 && len(me.kpETH.Pub) > 0 && len(me.kpETH.Priv) > 0 {

		//Encrypt unencrypted Private Key (always with Ethereum PW)
		me.kpPGP.Priv, err = pgp.EncryptPrivateKeys(string(me.kpETH.PrivPw), me.kpPGP.Priv)
		if err != nil {
			return err
		}

		if err = me.ensureStoreDir(); err != nil {
			return
		}

		buf := new(bytes.Buffer)
		err := me.ExportEncrypted(buf)
		if err != nil {
			log.Printf("[account][store] ExportEncrypted error: %s", err.Error())
			return err
		} else {
			var fw *os.File
			fw, err = os.OpenFile(me.GetFilePath(), os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
			if err != nil {
				log.Printf("[account][store] OpenFile error: %s", err.Error())
				return err
			}
			fw.Write(buf.Bytes())
			err = fw.Close()
		}
	}
	if err != nil {
		log.Printf("[account][store] stored: file %s, err: %s", me.GetFilePath(), err.Error())
	} else {
		log.Printf("[account][store] stored: file: %s", me.GetFilePath())
	}
	return
}

func (me *Account) ensureStoreDir() (err error) {
	_, err = os.Stat(me.cfg.StorageDir)
	if os.IsNotExist(err) {
		err = os.MkdirAll(me.cfg.StorageDir, 0750)
	}
	return
}

func (me *Account) Close() error {
	me.ethAddr = nil
	me.unlocked = false
	return nil
}

func (me *Account) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{"name": me.name, "address": string(me.kpETH.Pub), "pgpPublicKey": string(me.kpPGP.Pub)})
}

func (me *AccFile) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{"path": me.path, "name": me.acc.name, "address": string(me.acc.kpETH.Pub), "pgpPublicKey": string(me.acc.kpPGP.Pub)})
}
