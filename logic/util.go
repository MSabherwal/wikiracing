package logic

import "sync"

func NewConcurrentMap() *ConcurrentMap {
	return &ConcurrentMap{
		data:    make(map[string]string),
		RWMutex: &sync.RWMutex{},
	}
}

type ConcurrentMap struct {
	data map[string]string
	*sync.RWMutex
}

func (cm *ConcurrentMap) Get(in string) (string, bool) {
	cm.RLock()
	defer cm.RUnlock()
	val, exists := cm.data[in]
	return val, exists
}

func (cm *ConcurrentMap) Exists(in string) bool {
	cm.RLock()
	defer cm.RUnlock()
	_, exists := cm.data[in]
	return exists
}

func (cm *ConcurrentMap) Set(in, val string) {
	cm.Lock()
	defer cm.Unlock()
	cm.data[in] = val
	return
}

type ConcurrentSet struct {
	data map[string]struct{}
	*sync.RWMutex
}

func NewConcurrentSet() *ConcurrentSet {
	return &ConcurrentSet{
		data:    make(map[string]struct{}),
		RWMutex: &sync.RWMutex{},
	}
}

func (cs *ConcurrentSet) Exists(in string) bool {
	cs.RLock()
	defer cs.RUnlock()
	_, exists := cs.data[in]
	return exists
}

func (cs *ConcurrentSet) Add(in string) {
	cs.Lock()
	defer cs.Unlock()
	cs.data[in] = struct{}{}
}
