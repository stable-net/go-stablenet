// Copyright 2025 The go-wemix-wbft Authors
// This file is part of the go-wemix-wbft library.
//
// The go-wemix-wbft library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-wemix-wbft library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-wemix-wbft library. If not, see <http://www.gnu.org/licenses/>.

package wpoa

import (
	"container/list"
	"sync"
)

type LruCache struct {
	lock  *sync.RWMutex
	max   int  // max count
	count int  // current count
	fifo  bool // if true, element order is not updated on access
	lru   *list.List
	data  map[interface{}]interface{}
}

// NewLruCache creates LruCache
func NewLruCache(max int, fifo bool) *LruCache {
	return &LruCache{
		lock:  &sync.RWMutex{},
		max:   max,
		count: 0,
		fifo:  fifo,
		lru:   list.New(),
		data:  map[interface{}]interface{}{},
	}
}

// Count returns the current count of elements
func (c *LruCache) Count() int {
	return c.count
}

// Put adds a key-value pair
func (c *LruCache) Put(key, value interface{}) {
	c.lock.Lock()
	defer c.lock.Unlock()

	var e *list.Element
	_e, ok := c.data[key]
	if ok {
		e = _e.(*list.Element)
		e.Value = []interface{}{key, value}
		c.lru.MoveToFront(e)
	} else {
		if c.count >= c.max {
			e = c.lru.Back()
			delete(c.data, e.Value.([]interface{})[0])
			c.lru.Remove(e)
			c.count--
		}

		e = c.lru.PushFront([]interface{}{key, value})
		c.data[key] = e
		c.count++
	}
}

// Get returns a value with the given key if present, nil otherwise.
func (c *LruCache) Get(key interface{}) interface{} {
	c.lock.RLock()
	defer c.lock.RUnlock()

	_e, ok := c.data[key]
	if !ok {
		return nil
	} else {
		e := _e.(*list.Element)
		if !c.fifo {
			c.lru.MoveToFront(e)
		}
		return e.Value.([]interface{})[1]
	}
}

// Exists checks if a key exists.
func (c *LruCache) Exists(key interface{}) bool {
	c.lock.RLock()
	defer c.lock.RUnlock()
	_, ok := c.data[key]
	return ok
}

// Del deletes a key-value pair if present. It returns true iff it's present.
func (c *LruCache) Del(key interface{}) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	var e *list.Element
	_e, ok := c.data[key]
	if !ok {
		return false
	} else {
		e = _e.(*list.Element)
		delete(c.data, e.Value.([]interface{})[0])
		c.lru.Remove(e)
		c.count--
		return true
	}
}

// Clear resets the lru
func (c *LruCache) Clear() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.count = 0
	c.lru = list.New()
	c.data = map[interface{}]interface{}{}
}

// EOF
