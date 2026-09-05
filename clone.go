package sqlcache

import (
	"database/sql/driver"
	"strings"

	"github.com/prashanthpai/sqlcache/cache"
)

// cloneStrings copies src and the backing bytes of each name.
// Drivers such as bun/pgdriver return a pooled []string whose array (and, via
// unsafe.String, whose bytes) are reused by later queries.
func cloneStrings(src []string) []string {
	if src == nil {
		return nil
	}
	dst := make([]string, len(src))
	for i, s := range src {
		dst[i] = strings.Clone(s)
	}
	return dst
}

func cloneDriverValues(src []driver.Value) []driver.Value {
	if src == nil {
		return nil
	}
	dst := make([]driver.Value, len(src))
	for i, v := range src {
		dst[i] = cloneDriverValue(v)
	}
	return dst
}

func cloneDriverValue(v driver.Value) driver.Value {
	b, ok := v.([]byte)
	if !ok {
		return v
	}
	if b == nil {
		return []byte(nil)
	}
	cp := make([]byte, len(b))
	copy(cp, b)
	return cp
}

func cloneCacheItem(item *cache.Item) *cache.Item {
	if item == nil {
		return nil
	}
	out := &cache.Item{
		Cols: cloneStrings(item.Cols),
		Rows: make([][]driver.Value, len(item.Rows)),
	}
	for i, row := range item.Rows {
		out.Rows[i] = cloneDriverValues(row)
	}
	return out
}
