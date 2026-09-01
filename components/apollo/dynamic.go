package apollo

import "sync/atomic"

// Dynamic wraps a configuration struct T behind an atomic pointer so that
// OnChange updates (from a ",dynamic" apollo key) and concurrent reads are
// race-free: the writer Stores a brand-new *T (the old snapshot stays
// immutable), and readers Load the current *T. The returned *T is a snapshot
// — its fields may be read freely without locks, but it must not be mutated.
type Dynamic[T any] struct {
	p atomic.Pointer[T]
}

// Load returns the current snapshot *T. It returns nil before the first
// successful fill (e.g. the initial apollo fetch failed and the caller chose
// not to store a default).
func (d *Dynamic[T]) Load() *T { return d.p.Load() }

// Store replaces the current snapshot atomically.
func (d *Dynamic[T]) Store(t *T) { d.p.Store(t) }
