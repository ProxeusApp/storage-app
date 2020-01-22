package storage

import (
	"errors"

	"github.com/boltdb/bolt"
)

var (
	ErrNotFound = errors.New("value not found")
)

var DatabaseDir string

var db *bolt.DB
var bucketName = []byte("items")

func OpenDB() error {
	var err error
	db, err = bolt.Open(DatabaseDir, 0644, nil)
	if err != nil {
		return err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func CloseDB() {
	db.Close()
}

func GetPublicKey(ethereumAddress string) (publicKey string, err error) {
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		pk := bucket.Get([]byte(ethereumAddress))
		if pk == nil {
			return ErrNotFound
		}
		publicKey = string(pk)
		return nil
	})
	return publicKey, err
}

func SetPublicKey(ethereumAddress string, publicKey string) error {
	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(bucketName)
		return bucket.Put([]byte(ethereumAddress), []byte(publicKey))
	})
}
