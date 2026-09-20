package sync

import (
	"slices"
	"sync"
	"testing"
)

func TestMapSetAndGet(t *testing.T) {
	m := NewMap[string, string]()

	m.Set("key1", "value1")
	val, ok := m.Get("key1")
	if !ok || val != "value1" {
		t.Errorf("Expected to get 'value1', got %v, %v", val, ok)
	}

	// Get of non-existent key returns zero value
	val, _ = m.Get("nonexistent")
	if val != "" {
		t.Errorf("Expected zero value for non-existent key, got %v", val)
	}
}

func TestMapOverwrite(t *testing.T) {
	m := NewMap[string, int]()

	m.Set("key", 1)
	m.Set("key", 2)

	val, _ := m.Get("key")
	if val != 2 {
		t.Errorf("Expected overwritten value 2, got %d", val)
	}

	if m.Len() != 1 {
		t.Errorf("Expected length 1 after overwrite, got %d", m.Len())
	}
}

func TestMapDelete(t *testing.T) {
	m := NewMap[string, int]()

	m.Set("a", 1)
	m.Set("b", 2)

	m.Delete("a")

	if m.Len() != 1 {
		t.Errorf("Expected length 1 after delete, got %d", m.Len())
	}

	val, _ := m.Get("a")
	if val != 0 {
		t.Errorf("Expected zero value after delete, got %v", val)
	}

	// Delete non-existent key should not panic
	m.Delete("nonexistent")
}

func TestMapLen(t *testing.T) {
	m := NewMap[int, int]()

	if m.Len() != 0 {
		t.Errorf("Expected length 0 for new map, got %d", m.Len())
	}

	m.Set(1, 10)
	m.Set(2, 20)
	m.Set(3, 30)

	if m.Len() != 3 {
		t.Errorf("Expected length 3, got %d", m.Len())
	}

	m.Delete(2)

	if m.Len() != 2 {
		t.Errorf("Expected length 2 after delete, got %d", m.Len())
	}
}

func TestMapKeys(t *testing.T) {
	m := NewMap[int, string]()

	m.Set(1, "a")
	m.Set(2, "b")
	m.Set(3, "c")

	keys := m.Keys()
	slices.Sort(keys)

	if len(keys) != 3 || keys[0] != 1 || keys[1] != 2 || keys[2] != 3 {
		t.Errorf("Expected keys [1 2 3], got %v", keys)
	}
}

func TestMapValues(t *testing.T) {
	m := NewMap[int, int]()

	m.Set(1, 10)
	m.Set(2, 20)
	m.Set(3, 30)

	values := m.Values()
	slices.Sort(values)

	if len(values) != 3 || values[0] != 10 || values[1] != 20 || values[2] != 30 {
		t.Errorf("Expected values [10 20 30], got %v", values)
	}
}

func TestMapConcurrentReadWrite(t *testing.T) {
	m := NewMap[int, int]()
	const numGoroutines = 100
	const numOperations = 1000

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// Writers
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := (id * numOperations) + j
				m.Set(key, key*2)
			}
		}(i)
	}

	// Readers
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := (id * numOperations) + j
				m.Get(key)
			}
		}(i)
	}

	wg.Wait()

	expectedCount := numGoroutines * numOperations
	if m.Len() != expectedCount {
		t.Errorf("Expected %d entries, got %d", expectedCount, m.Len())
	}
}

func TestMapConcurrentDelete(t *testing.T) {
	m := NewMap[int, int]()
	const numGoroutines = 50
	const numKeys = 1000

	// Populate map
	for i := 0; i < numKeys; i++ {
		m.Set(i, i)
	}

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 2)

	// Deleters
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numKeys; j++ {
				if j%numGoroutines == id {
					m.Delete(j)
				}
			}
		}(i)
	}

	// Readers during deletion
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < numKeys; j++ {
				m.Get(j)
			}
		}()
	}

	wg.Wait()

	if m.Len() != 0 {
		t.Errorf("Expected 0 entries after deletion, got %d", m.Len())
	}
}
