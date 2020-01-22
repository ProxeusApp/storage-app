package embdb

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/boltdb/bolt"
)

type DB struct {
	failSafeLock sync.Mutex
	db           *bolt.DB
	dbPath       string
	dbName       string
}

func Open(dbPath, dbName string) (*DB, error) {
	db := &DB{dbPath: dbPath, dbName: dbName}
	err := db.openDB()
	if err != nil {
		return nil, err
	}
	return db, nil
}

func (me *DB) openDB() (err error) {
	me.failSafeLock.Lock()
	defer me.failSafeLock.Unlock()
	err = me.ensure(me.dbPath)
	if err != nil {
		return err
	}
	me.db, err = bolt.Open(filepath.Join(me.dbPath, me.dbName), 0600, &bolt.Options{Timeout: 60 * time.Second, ReadOnly: false})
	return
}

func (me *DB) ensure(p string) error {
	var err error
	_, err = os.Stat(p)
	if os.IsNotExist(err) {
		err = os.MkdirAll(p, 0750)
		if err != nil {
			return err
		}
	}
	return nil
}

func (me *DB) failSafeCheck() bool {
	p := filepath.Join(me.dbPath, me.dbName)
	_, err := os.Stat(p)
	if os.IsNotExist(err) {
		return true
	}
	return false
}

func (me *DB) Put(key []byte, val []byte) error {
	me.failSafeLock.Lock()
	err := me.db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("toAdd"))
		if err != nil {
			return err
		}
		err = b.Put(key, val)
		return err
	})
	me.failSafeLock.Unlock()
	if err != nil || me.failSafeCheck() {
		me.openDB()
		return os.ErrInvalid
	}
	return err
}

func (me *DB) Get(key []byte) (resBytes []byte, err error) {
	me.failSafeLock.Lock()
	err = me.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("toAdd"))
		if b != nil {
			v := b.Get(key)
			resBytes = make([]byte, len(v))
			copy(resBytes, v)
		}
		return nil
	})
	me.failSafeLock.Unlock()
	if err != nil || me.failSafeCheck() {
		me.openDB()
		err = os.ErrInvalid
	}
	return
}

func (me *DB) All() (keys [][]byte, err error) {
	keys = [][]byte{}
	me.failSafeLock.Lock()
	err = me.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("toAdd"))
		if b != nil {
			return b.ForEach(func(k, v []byte) error {
				newK := make([]byte, len(k))
				copy(newK, k)
				keys = append(keys, newK)
				return nil
			})
		}
		return nil
	})
	me.failSafeLock.Unlock()
	if err != nil || me.failSafeCheck() {
		me.openDB()
		err = os.ErrInvalid
	}
	return
}

func (me *DB) FilterKeyPrefix(keyPrefix []byte) (keys [][]byte, err error) {
	keys = [][]byte{}
	me.failSafeLock.Lock()
	err = me.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("toAdd"))
		if b != nil {
			return b.ForEach(func(k, v []byte) error {
				if bytes.HasPrefix(bytes.ToLower(k), bytes.ToLower(keyPrefix)) {
					newK := make([]byte, len(k))
					copy(newK, k)
					keys = append(keys, newK)
				}
				return nil
			})
		}
		return nil
	})
	me.failSafeLock.Unlock()
	if err != nil || me.failSafeCheck() {
		me.openDB()
		err = os.ErrInvalid
	}
	return
}

func (me *DB) FilterKeySuffix(keySuffix []byte) (keys [][]byte, err error) {
	keys = [][]byte{}
	me.failSafeLock.Lock()
	err = me.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("toAdd"))
		if b != nil {
			return b.ForEach(func(k, v []byte) error {
				if bytes.HasSuffix(bytes.ToLower(k), bytes.ToLower(keySuffix)) {
					newK := make([]byte, len(k))
					copy(newK, k)
					keys = append(keys, newK)
				}
				return nil
			})
		}
		return nil
	})
	me.failSafeLock.Unlock()
	if err != nil || me.failSafeCheck() {
		me.openDB()
		err = os.ErrInvalid
	}
	return
}

func (me *DB) AllWithValues() (keys [][]byte, vals [][]byte, err error) {
	keys = [][]byte{}
	vals = [][]byte{}
	me.failSafeLock.Lock()
	err = me.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("toAdd"))
		if b != nil {
			return b.ForEach(func(k, v []byte) error {
				newK := make([]byte, len(k))
				copy(newK, k)
				keys = append(keys, newK)
				newV := make([]byte, len(v))
				copy(newV, v)
				vals = append(vals, newV)
				return nil
			})
		}
		return nil
	})
	me.failSafeLock.Unlock()
	if err != nil || me.failSafeCheck() {
		me.openDB()
		err = os.ErrInvalid
	}
	return
}

func (me *DB) Del(key []byte) error {
	me.failSafeLock.Lock()
	err := me.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte("toAdd"))
		if b != nil {
			return b.Delete(key)
		}
		return nil
	})
	me.failSafeLock.Unlock()
	if err != nil || me.failSafeCheck() {
		me.openDB()
		return os.ErrInvalid
	}
	return nil
}

func (me *DB) Close() {
	me.failSafeLock.Lock()
	defer me.failSafeLock.Unlock()
	if me.db != nil {
		me.db.Close()
	}
}
