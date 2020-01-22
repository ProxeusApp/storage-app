package notification

import (
	"encoding/json"
	"os"
	"sort"
	"time"

	"github.com/pborman/uuid"

	"git.proxeus.com/core/central/dapp/core/embdb"
)

type (
	Manager struct {
		notificationDB *embdb.DB
		currentAccount string
	}

	/**
	TODO(ave) ->>>> pls note mmal hope it helps.. if not remove it
	search key [N_O_T_E]

	Notification simple

	/----------------------------------------/
	/---Title-----------------------------X--/
	/----------------------------------------/
	/---Description--------------------------/
	/----------------------------------------/


	Notification complex

	/----------------------------------------/
	/---Title-----------------------------X--/
	/----------------------------------------/
	/---Description--------------------------/
	/--------------------------[Action]------/

	handled by the frontend according to the notification type

	*/
	Notification struct {
		ID        string                 `json:"id"`
		Type      string                 `json:"type"`
		Timestamp uint64                 `json:"timestamp"`
		Unread    bool                   `json:"unread"`
		Pending   bool                   `json:"pending"`
		Data      map[string]interface{} `json:"data"`
		Dismissed bool                   `json:"dismissed"`
	}
	TimestampSorter []*Notification
)

const (
	NotificationDBName = "notification"
)

func (a TimestampSorter) Len() int           { return len(a) }
func (a TimestampSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a TimestampSorter) Less(i, j int) bool { return a[i].Timestamp < a[j].Timestamp }

func New(storageDir, currentAccount string) (*Manager, error) {
	var err error
	if currentAccount == "" {
		return nil, os.ErrInvalid
	}
	me := &Manager{currentAccount: currentAccount}
	me.notificationDB, err = embdb.Open(storageDir, NotificationDBName)
	if err != nil {
		return nil, err
	}
	return me, nil
}

/*
when finding tx check for txType as we may have multiple notifications with same tx-id
e.g.: tx_register and signing_request when adding own account as signer when uploading file
*/
func (me *Manager) AddOrUpdateAndAppendEventData(myType string, txHash string, notificationData map[string]interface{}, eventData interface{}) (*Notification, error) {
	n, err := me.FindByTxHashAndType(txHash, myType)
	if err == nil && n != nil {
		me.Append(eventData, n.Data)
		n, err = me.UpdateData(n.ID, n.Data)
	} else {
		me.Append(eventData, notificationData)
		n, err = me.Add(myType, notificationData)
	}
	return n, err
}

func (me *Manager) AddOrUpdate(myType string, filter map[string]string, data interface{}) (*Notification, error) {
	notification, err := me.FilterFirstMatch(myType, filter)
	if notification == nil && err == nil {
		return me.Add(myType, data)
	} else if notification != nil {
		return me.UpdateData(notification.ID, data)
	}

	return notification, err
}

func (me *Manager) Add(myType string, data interface{}) (*Notification, error) {
	id := uuid.NewRandom().String()
	pending := false
	if myType == "signing_request" || myType == "workflow_request" {
		pending = true
	}
	m := map[string]interface{}{"id": id, "type": myType, "data": data, "timestamp": time.Now().AddDate(0, 0, 0).Unix(), "unread": true, "pending": pending}
	bts, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	k := me.key(id)
	err = me.notificationDB.Put(k, bts)
	if err != nil {
		return nil, err
	}
	return me.get(k)
}

func (me *Manager) UpdateData(id string, data interface{}) (*Notification, error) {
	n, err := me.Get(id)
	if err != nil {
		return nil, err
	}
	dat, ok := data.(map[string]interface{})
	if ok {
		n.Data = dat
		me.Put(n.ID, *n)
	}
	return n, err
}

func (me *Manager) Get(id string) (*Notification, error) {
	return me.get(me.key(id))
}

func (me *Manager) Put(id string, notification Notification) error {
	return me.put(me.key(id), notification)
}

func (me *Manager) List() (res []*Notification, err error) {
	var keys [][]byte
	keys, err = me.notificationDB.FilterKeySuffix([]byte(me.currentAccount))
	if err != nil {
		return nil, err
	}
	res = make([]*Notification, 0)
	for _, k := range keys {
		n, err := me.get(k)
		if err != nil || n.Dismissed {
			continue
		}
		res = append(res, n)
	}
	sort.Sort(TimestampSorter(res))
	return
}

func (me *Manager) Filter(myType string, filter map[string]string) ([]*Notification, error) {
	keys, err := me.notificationDB.FilterKeySuffix([]byte(me.currentAccount))
	if err != nil {
		return nil, err
	}
	res := make([]*Notification, 0)
	for _, k := range keys {
		n, err := me.get(k)
		if err != nil {
			continue
		}
		for fKey, fVal := range filter {
			if myType == "" || n.Type == myType {
				dataVal, ok := n.Data[fKey].(string)
				if !ok {
					continue
				}
				if dataVal == fVal {
					res = append(res, n)
				}
			}
		}
	}
	return res, nil
}

func (me *Manager) FilterFirstMatch(myType string, filter map[string]string) (*Notification, error) {
	res, err := me.Filter(myType, filter)
	if err == nil && len(res) > 0 {
		return res[0], nil
	}
	return nil, err
}

//Usually searching by tx-hash and type make sense as multiple events may have same tx
func (me *Manager) FindByTxHashAndType(txHash, txType string) (*Notification, error) {
	return me.FilterFirstMatch(txType, map[string]string{"txHash": txHash})
}

//Usually searching by file hash and type make sense as multiple events may have same file hash
func (me *Manager) FindByFileHashAndType(fileHash, txType string) (*Notification, error) {
	return me.FilterFirstMatch(txType, map[string]string{"hash": fileHash})
}

func (me *Manager) get(k []byte) (*Notification, error) {
	bts, err := me.notificationDB.Get(k)
	if err != nil {
		return nil, err
	}
	n := Notification{}
	err = json.Unmarshal(bts, &n)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (me *Manager) put(k []byte, n Notification) error {
	bts, err := json.Marshal(&n)
	if err != nil {
		return err
	}
	err = me.notificationDB.Put(k, bts)
	if err != nil {
		return err
	}
	return nil
}

func (me *Manager) remove(id string) error {
	return me.notificationDB.Del(me.key(id))
}

func (me *Manager) MarkAsDismissed(id string) error {
	n, err := me.get(me.key(id))
	if err != nil {
		return err
	}
	n.Dismissed = true
	if err = me.put(me.key(n.ID), *n); err != nil {
		return err
	}
	return nil
}

func (me *Manager) MarkPendingAs(id string, val bool) (*Notification, error) {
	n, err := me.get(me.key(id))
	if err != nil {
		return nil, err
	}
	n.Pending = val
	me.put(me.key(n.ID), *n)
	return n, nil
}

func (me *Manager) MarkUnreadAs(id string, val bool) (*Notification, error) {
	n, err := me.get(me.key(id))
	if err != nil {
		return nil, err
	}
	n.Unread = val
	me.put(me.key(n.ID), *n)
	return n, nil
}

func (me *Manager) MarkFileRemovedAs(id string, val bool) (*Notification, error) {
	n, err := me.get(me.key(id))
	if err != nil {
		return nil, err
	}
	n.Data["fileRemoved"] = val
	me.put(me.key(n.ID), *n)
	return n, nil
}

func (me *Manager) key(id string) []byte {
	return []byte(id + "_" + me.currentAccount)
}

func (me *Manager) Close() error {
	me.notificationDB.Close()
	return nil
}

func (me *Manager) Append(source interface{}, target map[string]interface{}) {
	dat, ok := source.(map[string]interface{})
	if ok {
		target["owner"] = dat["owner"]
		target["fileName"] = dat["filename"]
	}
}
