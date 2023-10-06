package pokecache

import (
	"fmt"
	"testing"
	"time"
)

func TestAddGet(t *testing.T) {
	const interval = 5 * time.Second
	cases := []struct {
		key string
		val []byte
	}{
		{
			key: "https://example.com",
			val: []byte("testdata"),
		},
		{
			key: "https://example.com/path",
			val: []byte("moretestdata"),
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("Test case %v", i), func(t *testing.T) {
			cache := NewCache(interval)
			cache.Add(c.key, c.val)
			val, ok := cache.Get(c.key)
			if !ok {
				t.Errorf("expected to find key")
				return
			}
			if string(val) != string(c.val) {
				t.Errorf("expected to find value")
				return
			}
		})
	}
}

func TestReapLoop(t *testing.T) {
	const baseTime = 5 * time.Millisecond
	const waitTime = baseTime + 5*time.Millisecond
	cache := NewCache(baseTime)
	cache.Add("https://example.com", []byte("testdata"))

	_, ok := cache.Get("https://example.com")
	if !ok {
		t.Errorf("expected to find key")
		return
	}

	time.Sleep(waitTime)

	_, ok = cache.Get("https://example.com")
	if ok {
		t.Errorf("expected to not find key")
		return
	}
}

func TestDelete(t *testing.T) {
	const interval = 5 * time.Second
	cache := NewCache(interval)
	key := "https://example.com/delete"
	val := []byte("deletetestdata")

	cache.Add(key, val)
	cache.Delete(key)
	_, ok := cache.Get(key)
	if ok {
		t.Errorf("expected not to find key after deletion")
	}
}

func TestMultipleExpiry(t *testing.T) {
	const baseTime = 5 * time.Millisecond
	const waitTime = baseTime + 5*time.Millisecond
	cache := NewCache(baseTime)

	keys := []string{
		"https://example.com/1",
		"https://example.com/2",
		"https://example.com/3",
	}

	for _, key := range keys {
		cache.Add(key, []byte("testdata"))
	}

	for _, key := range keys {
		_, ok := cache.Get(key)
		if !ok {
			t.Errorf("expected to find key: %s", key)
		}
	}

	time.Sleep(waitTime)

	for _, key := range keys {
		_, ok := cache.Get(key)
		if ok {
			t.Errorf("expected key to expire: %s", key)
		}
	}
}
