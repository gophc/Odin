package common

// RangeValue represents a range [lo, hi].
type RangeValue struct {
	Lo i64
	Hi i64
}

// RangeCache caches index ranges for fast lookup.
type RangeCache struct {
	allocator Allocator
	entries   []RangeValue
}

// RangeCacheMake creates a new RangeCache.
func RangeCacheMake(a Allocator) RangeCache {
	return RangeCache{
		allocator: a,
		entries:   make([]RangeValue, 0, 64),
	}
}

// RangeCacheDestroy frees the range cache.
func RangeCacheDestroy(c *RangeCache) {
	c.entries = nil
}

// RangeCacheAddIndex adds a single index to the cache.
func RangeCacheAddIndex(c *RangeCache, index i64) bool {
	return RangeCacheAddRange(c, index, index)
}

// RangeCacheAddRange adds a range [lo, hi] to the cache.
func RangeCacheAddRange(c *RangeCache, lo, hi i64) bool {
	// Try to merge with existing ranges.
	// Find insertion point via binary search.
	left, right := 0, len(c.entries)
	for left < right {
		mid := (left + right) / 2
		if c.entries[mid].Hi+1 < lo {
			left = mid + 1
		} else {
			right = mid
		}
	}

	if left < len(c.entries) && c.entries[left].Lo <= hi+1 {
		// Merge.
		if lo < c.entries[left].Lo {
			c.entries[left].Lo = lo
		}
		if hi > c.entries[left].Hi {
			c.entries[left].Hi = hi
		}
		// Merge subsequent overlapping ranges.
		for left+1 < len(c.entries) && c.entries[left].Hi+1 >= c.entries[left+1].Lo {
			if c.entries[left+1].Hi > c.entries[left].Hi {
				c.entries[left].Hi = c.entries[left+1].Hi
			}
			copy(c.entries[left+1:], c.entries[left+2:])
			c.entries = c.entries[:len(c.entries)-1]
		}
		return true
	}

	// Insert new range.
	c.entries = append(c.entries, RangeValue{})
	copy(c.entries[left+1:], c.entries[left:])
	c.entries[left] = RangeValue{Lo: lo, Hi: hi}
	return true
}
