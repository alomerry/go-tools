package apollo

import (
	"encoding/json"
	"fmt"
	"sync"
	"testing"

	"github.com/apolloconfig/agollo/v4/storage"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDynamic_LoadStore verifies the basic atomic snapshot semantics: Load
// returns the last Stored value, and a previously Loaded snapshot is not
// mutated when a new value is Stored.
func TestDynamic_LoadStore(t *testing.T) {
	d := &Dynamic[testCfg]{}
	assert.Nil(t, d.Load(), "Load before first Store returns nil")

	d.Store(&testCfg{Port: 8080})
	snap := d.Load()
	require.NotNil(t, snap)
	assert.Equal(t, 8080, snap.Port)

	// Store a new value; the old snapshot must remain unchanged.
	d.Store(&testCfg{Port: 9090})
	assert.Equal(t, 8080, snap.Port, "old snapshot is immutable")

	newSnap := d.Load()
	require.NotNil(t, newSnap)
	assert.Equal(t, 9090, newSnap.Port)
}

// TestListener_OnChange_HotReload simulates a full dynamic hot-reload cycle:
// register a callback via TryWatchKey, fire OnChange, and assert the Dynamic
// snapshot updated to the new value while the old snapshot stayed immutable.
func TestListener_OnChange_HotReload(t *testing.T) {
	l := newApolloListener()

	type cfg struct {
		Name string `json:"name"`
	}

	d := &Dynamic[cfg]{}
	d.Store(&cfg{Name: "initial"})

	// Register the dynamic key callback.
	registered := l.TryWatchKey("my-key,dynamic", func(newVal string) {
		var nc cfg
		if err := json.Unmarshal([]byte(newVal), &nc); err != nil {
			return // fail-open: keep old snapshot
		}
		d.Store(&nc)
	})
	assert.True(t, registered, "dynamic key must be registered")

	oldSnap := d.Load()
	require.NotNil(t, oldSnap)
	assert.Equal(t, "initial", oldSnap.Name)

	// Simulate an apollo OnChange event.
	l.OnChange(&storage.ChangeEvent{
		Changes: map[string]*storage.ConfigChange{
			"my-key": {NewValue: `{"name":"updated"}`},
		},
	})

	newSnap := d.Load()
	require.NotNil(t, newSnap)
	assert.Equal(t, "updated", newSnap.Name, "Load returns the new snapshot after OnChange")
	assert.Equal(t, "initial", oldSnap.Name, "old snapshot is immutable")
}

// TestListener_OnChange_ParseFailureFailOpen verifies that a malformed JSON
// payload during OnChange does NOT replace the old snapshot (fail-open to
// last-known-good).
func TestListener_OnChange_ParseFailureFailOpen(t *testing.T) {
	l := newApolloListener()

	type cfg struct {
		Name string `json:"name"`
	}

	d := &Dynamic[cfg]{}
	d.Store(&cfg{Name: "good"})

	l.TryWatchKey("fail-open-key,dynamic", func(newVal string) {
		var nc cfg
		if err := json.Unmarshal([]byte(newVal), &nc); err != nil {
			return // keep old snapshot
		}
		d.Store(&nc)
	})

	// Fire OnChange with malformed JSON.
	l.OnChange(&storage.ChangeEvent{
		Changes: map[string]*storage.ConfigChange{
			"fail-open-key": {NewValue: `{not-json}`},
		},
	})

	snap := d.Load()
	require.NotNil(t, snap)
	assert.Equal(t, "good", snap.Name, "malformed payload must not replace the old snapshot")
}

// TestListener_NonDynamicKeyNotRegistered verifies that a key without the
// ",dynamic" suffix is not registered for OnChange callbacks.
func TestListener_NonDynamicKeyNotRegistered(t *testing.T) {
	l := newApolloListener()
	registered := l.TryWatchKey("plain-key", func(newVal string) {})
	assert.False(t, registered, "non-dynamic key must not be registered")
}

// TestDynamic_ConcurrentLoadStore verifies race-free concurrent access: multiple
// goroutines Load while another Stores. Run with -race to detect data races.
func TestDynamic_ConcurrentLoadStore(t *testing.T) {
	d := &Dynamic[testCfg]{}
	d.Store(&testCfg{Port: 1})

	var wg sync.WaitGroup
	wg.Add(2)

	// Writer goroutine
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			d.Store(&testCfg{Port: i})
		}
	}()

	// Reader goroutine
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			if snap := d.Load(); snap != nil {
				_ = snap.Port // safe read of immutable snapshot
			}
		}
	}()

	wg.Wait()
}

// TestDynamic_ConcurrentLoadAndOnChange verifies the core scenario of this
// refactor: multiple readers Load snapshots while OnChange Stores new ones
// concurrently. Run with -race to detect data races. Each Load must return a
// self-consistent snapshot (Port is never a half-written value).
func TestDynamic_ConcurrentLoadAndOnChange(t *testing.T) {
	l := newApolloListener()

	d := &Dynamic[testCfg]{}
	d.Store(&testCfg{Port: 0})

	l.TryWatchKey("concurrent-key,dynamic", func(newVal string) {
		var nc testCfg
		if err := json.Unmarshal([]byte(newVal), &nc); err != nil {
			return
		}
		d.Store(&nc)
	})

	var wg sync.WaitGroup
	wg.Add(2)

	// Writer: simulate OnChange events
	go func() {
		defer wg.Done()
		for i := 1; i <= 500; i++ {
			l.OnChange(&storage.ChangeEvent{
				Changes: map[string]*storage.ConfigChange{
					"concurrent-key": {NewValue: fmt.Sprintf(`{"port":%d}`, i)},
				},
			})
		}
	}()

	// Reader: concurrent Load + field read
	go func() {
		defer wg.Done()
		for i := 0; i < 1000; i++ {
			if snap := d.Load(); snap != nil {
				_ = snap.Port // safe read of immutable snapshot
			}
		}
	}()

	wg.Wait()

	// Final state must be the last Stored value.
	snap := d.Load()
	require.NotNil(t, snap)
	assert.Equal(t, 500, snap.Port)
}

type testCfg struct {
	Port int `json:"port"`
}
