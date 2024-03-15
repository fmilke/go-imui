package lru

import (
	"encoding/json"
)


type Entries[T any] map[rune]*LRUCacheEntry[T]

type LRUCacheEntry[T any] struct {
    key rune
	value T
	newer *LRUCacheEntry[T]
	older *LRUCacheEntry[T]
}

type LRUCache[T any] struct {
	entries Entries[T]
	oldest  *LRUCacheEntry[T]
	newest  *LRUCacheEntry[T]
    size    uint
    maxSize uint
}

func NewLRUCache[T any](s uint) *LRUCache[T] {
    return &LRUCache[T] {
        entries: make(Entries[T]),
        oldest: nil,
        newest: nil,
        size: 0,
        maxSize: s,
    }
}

// A helper functions, that removed
// and element from the linked list and
// reconnects open ends
// but the enrty itself still points
// to its siblings and needs correction
func (e *LRUCacheEntry[T]) unplug() {
	older := e.older
	newer := e.newer

	if older != nil {
		older = newer
	}

	if newer != nil {
		newer.older = older
	}
}

func (cache *LRUCache[T]) Get(key rune) *T {
	e := cache.entries[key]

	if e == nil {
        var r T
		return &r
	}

    if e == cache.newest {
        return &e.value
    }

	e.unplug()

	e.newer = nil
    e.older = cache.newest
    cache.newest.newer = e
    cache.newest = e

	return &e.value
}

func (cache *LRUCache[T]) Store(key rune, v T) {
    e := cache.entries[key]

    if e != nil {
        e.unplug()
    } else {
        e = &LRUCacheEntry[T] {
            value: v,
            older: cache.newest,
            key: key,
        }

        // increase counter 
        // or remove oldest one
        if cache.size < cache.maxSize {
            cache.size += 1
        } else if cache.oldest != nil {
            oldest := cache.oldest

            if oldest.newer != nil {
                cache.oldest = oldest.newer
                cache.oldest.newer = nil
            }

            delete(cache.entries, oldest.key)
        }

        cache.entries[key] = e
    }

    cache.newest = e

    if cache.oldest == nil {
        cache.oldest = e
    }
}

