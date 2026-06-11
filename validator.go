package sqlcache

import (
	"context"
	"database/sql/driver"
)

// ResultValidator inspects a cached resultset before it is served to the
// caller. It receives the cached column names and rows (raw driver.Values, in
// the same order the query selected them). Returning false causes sqlcache to
// discard the cached entry, re-read from the DB, and re-cache the fresh result
// — self-healing a poisoned or structurally-invalid entry.
//
// The validator runs on every cache hit for the query it is attached to, so
// keep it cheap (field presence / shape checks, not heavy work).
type ResultValidator func(cols []string, rows [][]driver.Value) bool

type validatorCtxKey struct{}

// WithResultValidator returns a copy of ctx carrying a per-query
// ResultValidator. Pass the returned context to db.QueryContext /
// stmt.QueryContext and sqlcache will invoke fn on any cache hit produced by
// that call. Because fn is a closure, it can capture the query's parameters
// (workspace id, slug, etc.) directly and assert the cached row actually
// matches what was requested.
//
// Only the query carrying cache directives is affected; calls without a
// validator behave exactly as before.
func WithResultValidator(ctx context.Context, fn ResultValidator) context.Context {
	if fn == nil {
		return ctx
	}
	return context.WithValue(ctx, validatorCtxKey{}, fn)
}

// validatorFromContext extracts a ResultValidator previously attached with
// WithResultValidator. It returns nil when none is set.
func validatorFromContext(ctx context.Context) ResultValidator {
	fn, _ := ctx.Value(validatorCtxKey{}).(ResultValidator)
	return fn
}
