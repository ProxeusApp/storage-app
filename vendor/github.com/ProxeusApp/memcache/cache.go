package cache

import (
	"errors"
	"os"
	"reflect"
	"sync"
	"time"
)

var (
	secondsAfter time.Duration = 1
	ErrNotExist                = errors.New("cache not exist")
)

type value struct {
	m      sync.RWMutex
	expiry time.Time
	access time.Time
	val    interface{}
}

type Cache struct {
	store               map[interface{}]*value
	Expiry              time.Duration
	cleanupTimer        *time.Timer
	cleanupLock         sync.Mutex
	cacheLock           sync.RWMutex
	defaultExtendExpiry bool
	OnExpired           func(key interface{}, val interface{})
}

func NewExtendExpiryOnGet(expiry time.Duration, extendExpiry bool) *Cache {
	c := &Cache{store: make(map[interface{}]*value), defaultExtendExpiry: extendExpiry}
	c.Expiry = expiry
	return c
}

func New(expiry time.Duration) *Cache {
	return NewExtendExpiryOnGet(expiry, false)
}

func (s *Cache) Get(key interface{}, ref interface{}) error {
	return s.GetAndExtendExpiry(key, ref, s.defaultExtendExpiry)
}

func (s *Cache) GetAndExtendExpiry(key interface{}, ref interface{}, extendExpiry bool) error {
	s.cacheLock.RLock()
	valueHolder := s.store[key]
	s.cacheLock.RUnlock()

	if valueHolder != nil {
		v := reflect.ValueOf(ref)
		if v.Kind() != reflect.Ptr || v.IsNil() {
			return os.ErrInvalid
		}

		//update last touch
		if extendExpiry {
			n := time.Now()
			valueHolder.m.Lock()
			valueHolder.access = n
			valueHolder.expiry = valueHolder.access.Add(s.Expiry)
			valueHolder.m.Unlock()
		}
		i := 0
		valueHolder.m.RLock()
		defer valueHolder.m.RUnlock()
		for v.Kind() != reflect.Struct && v.Kind() != reflect.Invalid && (!v.CanSet() || v.Type() != reflect.TypeOf(valueHolder.val)) {
			v = v.Elem()
			if i > 3 {
				break
			}
			i++
		}
		if !v.CanSet() || v.Kind() != reflect.TypeOf(valueHolder.val).Kind() {
			return os.ErrInvalid
		}
		v.Set(reflect.ValueOf(valueHolder.val))
		return nil
	}
	return ErrNotExist
}

func (s *Cache) Remove(key interface{}) bool {
	s.cacheLock.Lock()
	session := s.store[key]
	if session != nil {
		delete(s.store, key)
		s.cacheLock.Unlock()
		return true
	} else {
		s.cacheLock.Unlock()
		return false
	}
}

func (s *Cache) PutWithOtherExpiry(key interface{}, val interface{}, expiry time.Duration) {
	n := time.Now()
	exp := n.Add(expiry)
	session := &value{expiry: exp, access: n, val: val}
	s.cacheLock.Lock()
	s.store[key] = session
	s.cacheLock.Unlock()
	s.startCleanup(expiry + (secondsAfter * time.Second))
}

func (s *Cache) Put(key interface{}, val interface{}) {
	n := time.Now()
	expiry := n.Add(s.Expiry)
	s.cacheLock.Lock()
	v := s.store[key]
	if v == nil {
		v = &value{expiry: expiry, access: n, val: val}
		s.store[key] = v
		s.cacheLock.Unlock()
	} else {
		s.cacheLock.Unlock()
		v.m.Lock()
		v.access = n
		v.expiry = expiry
		v.val = val
		v.m.Unlock()
	}
	s.startCleanup(s.Expiry + (secondsAfter * time.Second))
}

func (s *Cache) cleanupScheduler() {
	//TODO improve
	n := time.Now()
	s.cleanupLock.Lock()
	minExpiry := n.Add(s.Expiry)
	expiredSessions := make(map[interface{}]*value)
	s.cacheLock.Lock()
	var key interface{}
	var val *value
	for key = range s.store {
		val = s.store[key]
		if n.After(val.expiry) {
			expiredSessions[key] = val
		} else if val.expiry.Before(minExpiry) {
			minExpiry = val.expiry
		}
	}
	for key = range expiredSessions {
		if s.OnExpired != nil {
			s.OnExpired(key, expiredSessions[key].val)
		}
		delete(s.store, key)
	}
	s.cacheLock.Unlock()
	for key = range expiredSessions {
		val = expiredSessions[key]
	}
	sessionsLength := len(s.store)
	if sessionsLength > 0 {
		nextRunIn := minExpiry.Sub(n) + (secondsAfter * time.Second)
		s.cleanupTimer = time.AfterFunc(nextRunIn, s.cleanupScheduler)
	} else {
		s.cleanupTimer = nil
	}
	s.cleanupLock.Unlock()
}

func (s *Cache) startCleanup(runAfter time.Duration) {
	if s.cleanupTimer == nil {
		s.cleanupLock.Lock()
		if s.cleanupTimer == nil {
			s.cleanupTimer = time.AfterFunc(runAfter, s.cleanupScheduler)
		}
		s.cleanupLock.Unlock()
	}
}

func (s *Cache) stopCleanup() {
	s.cleanupLock.Lock()
	if s.cleanupTimer != nil {
		s.cleanupTimer.Stop()
		s.cleanupTimer = nil
	}
	s.cleanupLock.Unlock()
}

func (s *Cache) Clean() {
	s.cacheLock.Lock()
	s.store = make(map[interface{}]*value)
	s.cacheLock.Unlock()
}

func (s *Cache) Close() {
	s.stopCleanup()
}
