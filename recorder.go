package sqlcache

import (
	"database/sql/driver"
	"io"

	"github.com/prashanthpai/sqlcache/cache"
)

func newRowsRecorder(setter func(item *cache.Item), rows driver.Rows, maxRows int, skipEmptyResultset bool) *rowsRecorder {
	return &rowsRecorder{
		item:    new(cache.Item),
		setter:  setter,
		maxRows: maxRows,
		dr:      rows,
		//
		skipEmptyResultset: skipEmptyResultset,
	}
}

type rowsRecorder struct {
	item       *cache.Item
	setter     func(item *cache.Item)
	gotErr     bool
	gotEOF     bool
	maxRowsHit bool
	maxRows    int
	dr         driver.Rows
	//
	skipEmptyResultset bool
}

func (r *rowsRecorder) Columns() []string {
	r.snapshotCols()
	return r.item.Cols
}

// snapshotCols copies column names before the driver Close. pgdriver returns
// rowDescription.names from a sync.Pool; that slice must not be stored as-is.
func (r *rowsRecorder) snapshotCols() {
	if r.item.Cols != nil {
		return
	}
	if cols := r.dr.Columns(); len(cols) > 0 {
		r.item.Cols = cloneStrings(cols)
	}
}

func (r *rowsRecorder) Close() error {
	r.snapshotCols()
	if err := r.dr.Close(); err != nil {
		r.gotErr = true
		return err
	}

	// cache only if we've reached EOF without any errors
	// and without hitting max rows limit
	if r.gotEOF && !r.gotErr && !r.maxRowsHit {
		// if skipEmpty is set, then we don't cache if there are no rows
		if !(r.skipEmptyResultset && r.isEmptyResultset()) {
			r.setter(r.item)
		}
	}
	return nil
}

func (r *rowsRecorder) isEmptyResultset() bool {
	return isItemEmpty(r.item)
}

// isItemEmpty reports whether a cache.Item represents an empty resultset:
// either zero rows, or a single row whose every column value is nil. The
// latter covers outer-join / aggregate queries that always return one row of
// NULLs when nothing matched. Used on both the write path (whether to cache)
// and the read path (whether to serve from cache).
func isItemEmpty(item *cache.Item) bool {
	if len(item.Rows) == 0 {
		return true
	}
	if len(item.Rows) > 1 {
		return false
	}
	// there is only 1 row. check if all column values are nil
	row := item.Rows[0]
	for idx := range item.Cols {
		if idx >= len(row) || row[idx] != nil {
			return false
		}
	}
	return true
}

func (r *rowsRecorder) Next(dest []driver.Value) error {
	err := r.dr.Next(dest)
	if err != nil {
		if err == io.EOF {
			r.gotEOF = true
		} else {
			r.gotErr = true
		}
	}

	if r.gotEOF || r.gotErr || r.maxRowsHit {
		return err
	}

	if len(r.item.Rows) == r.maxRows {
		r.maxRowsHit = true
		return err
	}

	r.item.Rows = append(r.item.Rows, cloneDriverValues(dest))

	return err
}
