package file

import (
	"encoding/json"
	"errors"
	"log"
	"strings"

	"git.proxeus.com/core/central/dapp/core/embdb"
)

type (
	handler struct {
		fileMetaDB *embdb.DB
	}

	FileMeta struct {
		FileHash     string
		FileName     string
		FileKind     int
		ContentName  string
		Uploaded     bool
		Hidden       bool
		Expiry       int64
		SpUrl        string
		GraceSeconds int
		Expired      bool
		HasThumbnail bool
	}
)

const (
	FileMetaDBName = "filemetadb"
	filesMetaName  = "filemeta"
)

func newFileMeta(userAccountDir string) (*handler, error) {
	me := new(handler)
	var err error
	me.fileMetaDB, err = embdb.Open(userAccountDir, FileMetaDBName)
	if err != nil {
		return nil, err
	}

	return me, nil
}

func (me *handler) Put(fileMeta *FileMeta) error {
	bts, err := json.Marshal(fileMeta)
	if err != nil {
		log.Println("[fileMetaHandler][Put] error while encoding file meta", err)
		return err
	}

	err = me.fileMetaDB.Put([]byte(strings.ToLower(fileMeta.FileHash)), bts)
	if err != nil {
		log.Println("[fileMetaHandler][Put] error while writing file meta", err)
		return err
	}

	log.Printf("[fileMetaHandler][Put] wrote FileMetaHandler %+v to fileMetaDB\n", fileMeta)
	return nil
}

var ErrFileMetaNotFound = errors.New("file meta not found")

func (me *handler) Get(fileHash string) (*FileMeta, error) {
	bts, err := me.fileMetaDB.Get([]byte(strings.ToLower(fileHash)))
	if err != nil {
		return nil, err
	}

	if len(bts) > 0 {
		fileMeta := FileMeta{}
		err := json.Unmarshal(bts, &fileMeta)
		if err != nil {
			log.Println("[fileMetaHandler][Get] deserialize FileMetaHandler error: ", err, string(bts))
			return nil, err
		}
		return &fileMeta, err
	}

	return nil, ErrFileMetaNotFound
}

func (me *handler) Del(fileHash string) error {
	return me.fileMetaDB.Del([]byte(strings.ToLower(fileHash)))
}

func (me *handler) close() {
	me.fileMetaDB.Close()
}
