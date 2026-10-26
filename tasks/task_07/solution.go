package main

import (
	"container/list"
	"sync"
)

type LRU[K comparable, V any] interface {
	Get(key K) (value V, ok bool)
	Set(key K, value V)
}
type entry[K comparable, V any] struct {
	key   K
	value V
}
type LRUCache[K comparable, V any] struct {
	capacity int
	mu       sync.Mutex
	ll       list.List
	items    map[K]*list.Element
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	return &LRUCache[K, V]{
		capacity: capacity,
		items:    make(map[K]*list.Element, capacity),
	}
}
func (c *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	element, ok := c.items[key]
	if !ok {
		return
	}
	c.ll.MoveToFront(element)
	value = element.Value.(entry[K, V]).value
	return
}
func (c *LRUCache[K, V]) Set(key K, value V) {
	if c.capacity <= 0 {
		return
	}
	_, is_exist := c.items[key]
	count := len(c.items)
	if is_exist {
		// just update value in map
		c.items[key].Value = entry[K, V]{key, value}
		return
	}

	if count >= c.capacity {
		// remove old from map, from list
		el_to_remove := c.ll.Back()
		delete(c.items, el_to_remove.Value.(entry[K, V]).key)
		c.ll.Remove(el_to_remove)
	}

	// insert new
	c.ll.PushFront(entry[K, V]{key, value})
	c.items[key] = c.ll.Front()
}
