package pgpService

import (
	"bytes"
	"crypto/rsa"
	"io"
	"testing"

	"github.com/ProxeusApp/pgp"
	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

func TestPrivateKeyEncryption(t *testing.T) {

	p, err := pgp.Create("test", "test@test.com", 4096, 0)
	if err != nil {
		t.Fatal("Couldn't create new PGP Keys", err)
	}

	privateKey := p["private"]

	entitylist, err := openpgp.ReadArmoredKeyRing(bytes.NewBuffer(privateKey))
	if err != nil {
		t.Fatal("Couldn't read PrivateKey", err)
	}
	entity := entitylist[0]

	beforeEncryption := entity.PrivateKey.PrivateKey.(*rsa.PrivateKey)
	rsaN := beforeEncryption.N
	rsaE := beforeEncryption.E
	rsaD := beforeEncryption.D
	rsaP0 := beforeEncryption.Primes[0]
	rsaP1 := beforeEncryption.Primes[1]
	rsaQinv := beforeEncryption.Precomputed.Qinv

	passphrase := "This is a Test Passphrase!"
	encryptedBytes, err := EncryptPrivateKey(passphrase, privateKey)
	if err != nil {
		t.Fatal("Couldn't encrypt PrivateKey", err)
	}

	newEntitylist, err := openpgp.ReadArmoredKeyRing(bytes.NewBuffer(encryptedBytes))
	if err != nil {
		t.Fatal("Couldn't read encrypted PrivateKey", err)
	}
	newEntity := newEntitylist[0]

	newEntity.PrivateKey.Decrypt([]byte(passphrase))
	if err != nil {
		t.Fatal("Couldn't decrypt PrivateKey", err)
	}

	afterEncryption := newEntity.PrivateKey.PrivateKey.(*rsa.PrivateKey)
	rsaNrestored := afterEncryption.N
	rsaErestored := afterEncryption.E
	rsaDrestored := afterEncryption.D
	rsaP0restored := afterEncryption.Primes[0]
	rsaP1restored := afterEncryption.Primes[1]
	rsaQinvrestored := afterEncryption.Precomputed.Qinv

	if rsaN.Uint64() != rsaNrestored.Uint64() {
		t.Fatal("N parameter mismatch:", rsaN, rsaNrestored)
	}
	if rsaE != rsaErestored {
		t.Fatal("E parameter mismatch:", rsaE, rsaErestored)
	}
	if rsaD.Uint64() != rsaDrestored.Uint64() {
		t.Fatal("D parameter mismatch:", rsaD, rsaDrestored)
	}
	if rsaP0.Uint64() != rsaP0restored.Uint64() {
		t.Fatal("Prime 0 parameter mismatch:", rsaP0, rsaP0restored)
	}
	if rsaP1.Uint64() != rsaP1restored.Uint64() {
		t.Fatal("Prime 1 parameter mismatch:", rsaP1, rsaP1restored)
	}
	if rsaQinv.Uint64() != rsaQinvrestored.Uint64() {
		t.Fatal("Qinv parameter mismatch:", rsaQinv, rsaQinvrestored)
	}

	afterEncryption.Validate()
	if err != nil {
		t.Fatal("PrivateKey validation failed: ", err)
	}
}

//copy of Method from account.go that will go into PGP.go that handles interaction with EncryptablePrivateKey
func EncryptPrivateKey(passphrase string, priv []byte) ([]byte, error) {
	entitylist, err := openpgp.ReadArmoredKeyRing(bytes.NewBuffer(priv))
	if err != nil {
		return nil, err
	}
	entity := entitylist[0]

	if entity.PrivateKey.Encrypted {
		return priv, nil
	}

	buf := new(bytes.Buffer)
	ar, err := armor.Encode(buf, openpgp.PrivateKeyType, nil)
	if err != nil {
		return nil, err
	}

	encryptPriv := func(p *packet.PrivateKey, w *io.WriteCloser) (err error) {
		privencryption := new(EncryptablePrivateKey)
		privencryption.NewEncryptablePrivateKey(p)
		err = privencryption.Encrypt([]byte(passphrase))
		if err != nil {
			return err
		}
		err = privencryption.SerializePrivate(*w)
		if err != nil {
			return err
		}
		return nil
	}

	err = encryptPriv(entity.PrivateKey, &ar)
	if err != nil {
		return nil, err
	}

	for _, ident := range entity.Identities {
		err = ident.UserId.Serialize(ar)
		if err != nil {
			return nil, err
		}
		err = ident.SelfSignature.SignUserId(ident.UserId.Id, entity.PrimaryKey, entity.PrivateKey, nil)
		if err != nil {
			return nil, err
		}
		err = ident.SelfSignature.Serialize(ar)
		if err != nil {
			return nil, err
		}
	}

	for _, subkey := range entity.Subkeys {
		err = encryptPriv(subkey.PrivateKey, &ar)
		if err != nil {
			return nil, err
		}
		err = subkey.Sig.SignKey(subkey.PublicKey, entity.PrivateKey, nil)
		if err != nil {
			return nil, err
		}
		err = subkey.Sig.Serialize(ar)
		if err != nil {
			return nil, err
		}
	}
	ar.Close()
	return buf.Bytes(), nil
}
