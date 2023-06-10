package lru

import (
	"github.com/danielgatis/go-freetype/freetype"
)

type LRUCacheEntry struct {
	value freetype.Metrics
	newer *LRUCacheEntry
	older *LRUCacheEntry
}

type LRUCache struct {
	entries map[rune]*LRUCacheEntry
	oldest  *LRUCacheEntry
	newest  *LRUCacheEntry
}

func (e *LRUCacheEntry) unplug() {
	older := e.older
	newer := e.newer

	if older != nil {
		older = newer
	}

	if newer != nil {
		newer.older = nil
	}
}

func (cache *LRUCache) get(key rune) *freetype.Metrics {
	e := cache.entries[key]

	if e == nil || cache.newest == e {
		return nil
	}

	e.unplug()

	e.newer = nil
	if cache.newest != nil {
		e.older = cache.newest
		cache.newest.newer = e
		cache.newest = e
	} else {
		e.older = nil
		cache.newest = e
	}

	return &e.value
}

func (cache *LRUCache) store(key rune, metrics freetype.Metrics) {
	old := cache.entries[key]

	e := &LRUCacheEntry{
		value: metrics,
		older: cache.newest,
	}

	if old != nil {
		old.unplug()
	}

	cache.entries[key] = e

	if cache.newest != nil {
		cache.newest.newer = e
	}
	cache.newest = e
}
