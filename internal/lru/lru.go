package lru

import "github.com/benoitkugler/textlayout/fonts"

type Key = fonts.GID
type Entries[T any] map[Key]*LRUCacheEntry[T]

type LRUCacheEntry[T any] struct {
    key Key
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

func (cache *LRUCache[T]) Get(key Key) *T {
	e := cache.entries[key]

	if e == nil {
		return nil
	}

    // this introduces two invariants for the rest of the function
    // 1) the entry is not the newest
    // 2) there is more than a single entry in the cache (otherwise 1) would not hold)
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

func (cache *LRUCache[T]) Store(key Key, v T) {
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

