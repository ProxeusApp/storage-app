package embdb

type (
	MemoryDB struct{}
)

func OpenDummyDB() *MemoryDB {
	db := &MemoryDB{}
	return db
}
func (me *MemoryDB) Put(key []byte, val []byte) error {
	return nil
}

func (me *MemoryDB) Get(key []byte) (resBytes []byte, err error) {
	return
}

func (me *MemoryDB) All() (keys [][]byte, err error) {
	return
}

func (me *MemoryDB) FilterKeySuffix(keySuffix []byte) (keys [][]byte, err error) {
	return
}

func (me *MemoryDB) AllWithValues() (keys [][]byte, vals [][]byte, err error) {
	return
}

func (me *MemoryDB) Del(key []byte) error {
	return nil
}

func (me *MemoryDB) Close() {
	return
}
