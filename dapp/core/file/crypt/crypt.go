package crypt

import (
	"fmt"
	"log"
	"os"

	"git.proxeus.com/core/central/dapp/core/file/archive"

	"github.com/ProxeusApp/pgp"
)

func EncryptDirectory(dst string, src string, pgpPublicKeys [][]byte) error {
	archiveFilePath := fmt.Sprintf("%s_%s", src, "archived")
	archiveFile, err := os.OpenFile(archiveFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer archiveFile.Close() //close archiveFile in any case

	if err = archive.Tar(src, archiveFile); err != nil {
		return err
	}

	err = EncryptFile(dst, archiveFilePath, pgpPublicKeys)
	if err != nil {
		log.Printf("encryptDirectory: Failed to encrypt : %s, Error: %s", archiveFilePath, err.Error())
		return err
	}

	if err = archiveFile.Close(); err != nil { //close archiveFile before removing
		log.Printf("encryptDirectory: Failed to close archive directory: %s, Error: %s", src, err.Error())
	}
	if err = os.Remove(archiveFilePath); err != nil {
		log.Printf("encryptDirectory: Failed to remove archive directory: %s, Error: %s", src, err.Error())
	}

	if err = os.RemoveAll(src); err != nil {
		log.Printf("encryptDirectory: Failed to remove plain directory: %s, Error: %s", src, err.Error())
	}

	return err
}

func DecryptDirectory(dst string, src string, pw, pgpPrivateKey []byte) error {
	decryptedOutputFileSrc := fmt.Sprintf("%s_%s", dst, "decrypted")
	if err := DecryptFile(decryptedOutputFileSrc, src, pw, pgpPrivateKey); err != nil {
		return err
	}

	decryptedOutputFile, err := os.OpenFile(decryptedOutputFileSrc, os.O_RDONLY, 0660)
	if err != nil {
		return err
	}
	defer decryptedOutputFile.Close() //close decryptedOutputFile in any case

	if err = archive.Untar(dst, decryptedOutputFile); err != nil {
		return err
	}

	if err = os.Remove(src); err != nil {
		return err
	}

	decryptedOutputFile.Close() //close decryptedOutputFile before calling remove
	return os.Remove(decryptedOutputFileSrc)
}

func EncryptFile(dst string, src string, pgpPublicKeys [][]byte) error {
	srcf, err := os.Open(src)
	defer srcf.Close()
	if err != nil {
		return err
	}
	dstf, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	defer dstf.Close()
	if err != nil {
		return err
	}
	_, err = pgp.EncryptStream(srcf, dstf, pgpPublicKeys)

	return err
}

func DecryptFile(dst string, src string, pw, pgpPrivateKey []byte) error {
	srcf, err := os.OpenFile(src, os.O_RDONLY, 0600)
	defer srcf.Close()
	if err != nil {
		return err
	}
	dstf, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	defer dstf.Close()
	if err != nil {
		return err
	}
	_, err = pgp.DecryptStream(srcf, dstf, pw, pgpPrivateKey)
	return err
}

func Encrypt(msg []byte, pubKey [][]byte) ([]byte, error) {
	return pgp.Encrypt(msg, pubKey)
}
