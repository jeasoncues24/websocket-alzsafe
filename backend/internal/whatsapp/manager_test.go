package whatsapp

import (
	"sync"
	"testing"

	"go.mau.fi/whatsmeow"
)

func TestManagerSetGetDelete(t *testing.T) {
	m := NewManager()

	var c *whatsmeow.Client
	m.Set("20123456789", c)

	got, ok := m.Get("20123456789")
	if !ok {
		t.Fatalf("expected key to exist")
	}
	if got != c {
		t.Fatalf("expected same client reference")
	}

	if !m.Exists("20123456789") {
		t.Fatalf("expected Exists to return true")
	}

	m.Delete("20123456789")
	if m.Exists("20123456789") {
		t.Fatalf("expected key to be deleted")
	}
}

func TestManagerNormalizesAccountID(t *testing.T) {
	m := NewManager()

	var c *whatsmeow.Client
	m.Set(" 20123456789 ", c)

	if _, ok := m.Get("20123456789"); !ok {
		t.Fatalf("expected normalized key to be found")
	}

	if m.Count() != 1 {
		t.Fatalf("expected count to be 1, got %d", m.Count())
	}

	m.Set("+20123456789", c)
	if m.Count() != 1 {
		t.Fatalf("expected plus and non-plus to map same account, got %d", m.Count())
	}
	if _, ok := m.Get("+20123456789"); !ok {
		t.Fatalf("expected get with plus prefix to be normalized")
	}
}

func TestManagerListKeys(t *testing.T) {
	m := NewManager()

	m.Set("1", nil)
	m.Set("2", nil)

	keys := m.ListKeys()
	if len(keys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(keys))
	}
}

func TestManagerConcurrentAccess(t *testing.T) {
	m := NewManager()
	wg := sync.WaitGroup{}

	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			key := "acc"
			if n%2 == 0 {
				key = "acc-even"
			}
			m.Set(key, nil)
			m.Exists(key)
			m.Get(key)
		}(i)
	}

	wg.Wait()

	if m.Count() == 0 {
		t.Fatalf("expected at least one key after concurrent writes")
	}
}
