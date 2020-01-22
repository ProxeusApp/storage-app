package pgpService

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"io"
	"math/big"

	"golang.org/x/crypto/openpgp/elgamal"
	"golang.org/x/crypto/openpgp/errors"
	"golang.org/x/crypto/openpgp/packet"
	"golang.org/x/crypto/openpgp/s2k"
)

/*
  EncryptablePrivateKey provides
    func (pk *EncryptablePrivateKey) NewEncryptablePrivateKey(priv *packet.PrivateKey)
    func (pk *EncryptablePrivateKey) Encrypt(passphrase []byte) error
    func (pk *EncryptablePrivateKey) SerializePrivate(w io.Writer) error

*/

type EncryptablePrivateKey struct {
	packet.PrivateKey

	//privatekey encryption variables
	encryptedData []byte
	cipher        packet.CipherFunction
	s2k           func(out, in []byte)
	sha1Checksum  bool
	iv            []byte

	//encryption key derivation variables
	s2kmode  uint8       // only inverted+salted mode is used
	s2khash  crypto.Hash // Crypto.SHA256 is used
	s2kckc   uint8       // only sha1 checksum is used
	s2ksalt  []byte      // randomly generated
	s2kcount uint8       // as per s2kcountstd constant
}

//s2kmode constants
const (
	s2ksimple         uint8 = 0
	s2iterated        uint8 = 1
	s2kiteratedsalted uint8 = 3 // only inverted+salted mode is used
)

//s2kckc constants
const (
	s2knon      uint8 = 0
	s2ksha1     uint8 = 254 // only sha1 checksum is used
	s2kchecksum uint8 = 255
)

const (
	packetTypePrivateKey    uint8 = 5
	packetTypePublicKey     uint8 = 6
	packetTypePrivateSubkey uint8 = 7
)

const s2kcountstd uint32 = 65011712 // s2k iterations used
const s2kcountstd_octet uint8 = 255 // s2k iterations used

func (pk *EncryptablePrivateKey) NewEncryptablePrivateKey(priv *packet.PrivateKey) {
	pk.PrivateKey = *priv
}

//Encrypts private key with aes-128 CFB, based on an iterated/salted s2k key derivation
//of the supplied passphrase and uses SHA1 as checksum
func (pk *EncryptablePrivateKey) Encrypt(passphrase []byte) error {
	switch pk.PrivateKey.PrivateKey.(type) {
	case *rsa.PrivateKey, *dsa.PrivateKey, *ecdsa.PrivateKey, *elgamal.PrivateKey:
		privateKeyBuf := bytes.NewBuffer(nil)
		err := pk.serializePrivMPI(privateKeyBuf)
		if err != nil {
			return err
		}
		privateKeyBytes := privateKeyBuf.Bytes()

		//key derivation
		key := make([]byte, 16)
		pk.s2ksalt = make([]byte, 8)
		rand.Read(pk.s2ksalt)
		pk.s2k = func(out, in []byte) {
			s2k.Iterated(out, pk.s2khash.New(), in, pk.s2ksalt, int(s2kcountstd))
		}
		pk.s2khash = crypto.SHA256
		pk.s2k(key, passphrase)
		pk.s2kmode = s2kiteratedsalted
		pk.s2kcount = s2kcountstd_octet

		//encryption
		block, _ := aes.NewCipher(key)
		pk.iv = make([]byte, block.BlockSize())
		rand.Read(pk.iv)
		cfb := cipher.NewCFBEncrypter(block, pk.iv)
		h := sha1.New()
		h.Write(privateKeyBytes)
		sum := h.Sum(nil)
		privateKeyBytes = append(privateKeyBytes, sum...)
		pk.s2kckc = s2ksha1

		pk.encryptedData = make([]byte, len(privateKeyBytes))

		cfb.XORKeyStream(pk.encryptedData, privateKeyBytes)
		pk.Encrypted = true

		return err
	}
	return errors.UnsupportedError("no exportable private key found")
}

func (pk *EncryptablePrivateKey) SerializePrivate(w io.Writer) error {
	buf := bytes.NewBuffer(nil)

	if pk.Encrypted {
		pk.serializeSecretKeyPacket(buf)
	} else {
		return errors.UnsupportedError("only encrypted private keys supported")
	}

	ptype := packetTypePrivateKey
	contents := buf.Bytes()
	if pk.PrivateKey.PublicKey.IsSubkey {
		ptype = packetTypePrivateSubkey
	}
	err := serializeHeader(w, ptype, len(contents))
	if err != nil {
		return err
	}
	_, err = w.Write(contents)
	if err != nil {
		return err
	}
	return nil
}

func (pk *EncryptablePrivateKey) serializeSecretKeyPacket(w io.Writer) error {
	err := serializePublicKey(&pk.PrivateKey.PublicKey, w)
	if err != nil {
		return err
	}

	privateKeyBuf := bytes.NewBuffer(nil)
	encodedKeyBuf := bytes.NewBuffer(nil)

	//checksum sha1
	encodedKeyBuf.Write([]byte{uint8(pk.s2kckc)})

	//cipher aes-128
	pk.cipher = packet.CipherAES128
	encodedKeyBuf.Write([]byte{uint8(pk.cipher)})

	//s2k iterated/salted
	encodedKeyBuf.Write([]byte{pk.s2kmode})
	hashID, ok := s2k.HashToHashId(pk.s2khash)
	if !ok {
		return errors.UnsupportedError("no such hash")
	}
	encodedKeyBuf.Write([]byte{hashID})

	//s2k salt
	encodedKeyBuf.Write(pk.s2ksalt)

	//s2k iterations
	encodedKeyBuf.Write([]byte{pk.s2kcount})

	//encrypted privatekey MPIs
	privateKeyBuf.Write(pk.encryptedData)

	encodedKey := encodedKeyBuf.Bytes()
	privateKeyBytes := privateKeyBuf.Bytes()

	w.Write(encodedKey)
	w.Write(pk.iv)
	w.Write(privateKeyBytes)

	//sha1 hash checksum
	h := sha1.New()
	h.Write(privateKeyBytes)
	sum := h.Sum(nil)
	privateKeyBytes = append(privateKeyBytes, sum...)

	return nil
}

func (pk *EncryptablePrivateKey) serializePrivMPI(w io.Writer) error {
	switch pk.PrivateKey.PubKeyAlgo {
	case packet.PubKeyAlgoRSA, packet.PubKeyAlgoRSAEncryptOnly, packet.PubKeyAlgoRSASignOnly:
		rsaPrivateKey := pk.PrivateKey.PrivateKey.(*rsa.PrivateKey)
		return writeMPIs(w, fromBig(rsaPrivateKey.D), fromBig(rsaPrivateKey.Primes[0]),
			fromBig(rsaPrivateKey.Primes[1]), fromBig(rsaPrivateKey.Precomputed.Qinv))
	case packet.PubKeyAlgoDSA:
		dsaPrivateKey := pk.PrivateKey.PrivateKey.(*dsa.PrivateKey)
		return writeMPIs(w, fromBig(dsaPrivateKey.X))
	case packet.PubKeyAlgoElGamal:
		elgamalPrivateKey := pk.PrivateKey.PrivateKey.(*elgamal.PrivateKey)
		return writeMPIs(w, fromBig(elgamalPrivateKey.X))
	case packet.PubKeyAlgoECDSA:
		ecdsaPrivateKey := pk.PrivateKey.PrivateKey.(*ecdsa.PrivateKey)
		return writeMPIs(w, fromBig(ecdsaPrivateKey.D))
	}
	return errors.InvalidArgumentError("unknown private key type")
}

//excat copy from crypto/openpgp/packet/Packet.go
func serializeHeader(w io.Writer, ptype uint8, length int) (err error) {
	var buf [6]byte
	var n int

	buf[0] = 0x80 | 0x40 | byte(ptype)
	if length < 192 {
		buf[1] = byte(length)
		n = 2
	} else if length < 8384 {
		length -= 192
		buf[1] = 192 + byte(length>>8)
		buf[2] = byte(length)
		n = 3
	} else {
		buf[1] = 255
		buf[2] = byte(length >> 24)
		buf[3] = byte(length >> 16)
		buf[4] = byte(length >> 8)
		buf[5] = byte(length)
		n = 6
	}

	_, err = w.Write(buf[:n])
	return
}

//copy from crypto/openpgp/packet/public_key.go with minimal changes to access publickey data
func serializePublicKey(pk *packet.PublicKey, w io.Writer) (err error) {
	var buf [6]byte
	buf[0] = 4
	t := uint32(pk.CreationTime.Unix())
	buf[1] = byte(t >> 24)
	buf[2] = byte(t >> 16)
	buf[3] = byte(t >> 8)
	buf[4] = byte(t)
	buf[5] = byte(pk.PubKeyAlgo)

	_, err = w.Write(buf[:])
	if err != nil {
		return
	}

	switch pk.PubKeyAlgo {
	case packet.PubKeyAlgoRSA, packet.PubKeyAlgoRSAEncryptOnly, packet.PubKeyAlgoRSASignOnly:
		rsaPublicKey := pk.PublicKey.(*rsa.PublicKey)
		return writeMPIs(w, fromBig(rsaPublicKey.N), fromBig(big.NewInt(int64(rsaPublicKey.E))))
	case packet.PubKeyAlgoDSA:
		dsaPublicKey := pk.PublicKey.(*dsa.PublicKey)
		return writeMPIs(w, fromBig(dsaPublicKey.P), fromBig(dsaPublicKey.Q),
			fromBig(dsaPublicKey.G), fromBig(dsaPublicKey.Y))
	case packet.PubKeyAlgoElGamal:
		elgamalPublicKey := pk.PublicKey.(*elgamal.PublicKey)
		return writeMPIs(w, fromBig(elgamalPublicKey.P), fromBig(elgamalPublicKey.G),
			fromBig(elgamalPublicKey.Y))
	}
	return errors.InvalidArgumentError("bad public-key algorithm")
}

//exact copy from crypto/openpgp/packet/public_key.go
type parsedMPI struct {
	bytes     []byte
	bitLength uint16
}

//exact copy from crypto/openpgp/packet/public_key.go
func fromBig(n *big.Int) parsedMPI {
	return parsedMPI{
		bytes:     n.Bytes(),
		bitLength: uint16(n.BitLen()),
	}
}

//exact copy from crypto/openpgp/packet/public_key.go
func writeMPIs(w io.Writer, mpis ...parsedMPI) (err error) {
	for _, mpi := range mpis {
		err = writeMPI(w, mpi.bitLength, mpi.bytes)
		if err != nil {
			return
		}
	}
	return
}

//exact copy from crypto/openpgp/packet/packet.go
func writeMPI(w io.Writer, bitLength uint16, mpiBytes []byte) (err error) {
	// Note that we can produce leading zeroes, in violation of RFC 4880 3.2.
	// Implementations seem to be tolerant of them, and stripping them would
	// make it complex to guarantee matching re-serialization.
	_, err = w.Write([]byte{byte(bitLength >> 8), byte(bitLength)})
	if err == nil {
		_, err = w.Write(mpiBytes)
	}
	return
}
