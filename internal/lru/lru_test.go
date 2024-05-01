package lru

import (
	"testing"
)

func TestSimpleAdd(t *testing.T) {
    toInsert := 1
    c := NewLRUCache[int](3)
    c.Store('r', toInsert)
    v := c.Get('r')

    if *v != toInsert {
        t.Fatalf("Failed to get value after insert")
    }
}

func TestEntryReplaced(t *testing.T) {
    toInsert := 1
    c := NewLRUCache[int](3)
    c.Store('r', toInsert)
    v := c.Get('r')

    if *v != toInsert {
        t.Fatalf("Failed to get value after insert")
    }

    c.Store('a', 2)
    c.Store('b', 3)
    c.Store('c', 4)

    v = c.Get('r')
    if v != nil && *v == toInsert {
        t.Fatal("Failed to replace value, when cache is filled")
    }
}

func TestEntryNotReplacedSinceUsed(t *testing.T) {
    toInsert := 1
    c := NewLRUCache[int](3)
    c.Store('r', toInsert)
    v := c.Get('r')

    if *v != toInsert {
        t.Fatalf("Failed to get value after insert")
    }

    c.Store('a', 2)
    c.Store('b', 3)
    c.Get('r') // use r, to bring it to the front - now 'a' should be least recently used
    c.Store('c', 4)

    v = c.Get('r')
    if v != nil && *v == toInsert {
        t.Fatal("Failed to replace value, when cache is filled")
    }

    v = c.Get('a')
    if v != nil && *v == 0 {
        t.Fatal("Failed to replace value, when cache is filled")
    }
}
