package vecd

import (
	"image"
	"sync"
)

// DoubleBuffer is a buffer that allows swapping two images.
type DoubleBuffer struct {
	mu   sync.Mutex
	ours *image.RGBA
}

var none = image.NewRGBA(image.Rect(0, 0, 1, 1))

// Swap swaps the contents of the given image with the contents of the buffer
// and returns the previous contents of the buffer.
func (db *DoubleBuffer) Swap(other *image.RGBA) *image.RGBA {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.ours == nil || db.ours.Bounds() != other.Bounds() {
		db.ours = image.NewRGBA(other.Bounds())
	}

	*db.ours, *other = *other, *db.ours
	return other
}

// Acquire calls the given function with the current contents of the buffer.
func (db *DoubleBuffer) Acquire(f func(*image.RGBA)) {
	db.mu.Lock()
	defer db.mu.Unlock()

	if db.ours == nil {
		f(none)
	} else {
		f(db.ours)
	}
}
