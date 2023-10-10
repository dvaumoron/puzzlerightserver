// Generated from model.crn - do not edit.

package model

import (
	"context"
	"database/sql"
	"strconv"
	"strings"
	"time"
)

const timeout = time.Duration(int64(10)) * time.Second

type ExecerContext interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

type RowQueryerContext interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

type QueryerContext interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

func varArgsFilter(query string, placeholder string, size int) string {
	startIndex, _ := strconv.Atoi(placeholder[1:])
	placeholders := make([]string, 0, size)
	size += startIndex
	for i := startIndex; i < size; i++ {
		placeholders = append(placeholders, "$"+strconv.Itoa(i))
	}
	return strings.ReplaceAll(query, placeholder, strings.Join(placeholders, ", "))
}

func anyConverter[S ~[]E, E any](args S) []any {
	results := make([]any, int64(0), len(args))
	for _, arg := range args {
		results = append(results, arg)
	}
	return results
}
