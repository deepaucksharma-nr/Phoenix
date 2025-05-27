package database

import (
	"database/sql"
	"database/sql/driver"
	"time"
)

// NullString represents a string that may be null.
// It's a wrapper around sql.NullString to avoid direct imports.
type NullString = sql.NullString

// NullTime represents a time.Time that may be null.
// It's a wrapper around sql.NullTime to avoid direct imports.
type NullTime = sql.NullTime

// NullInt64 represents an int64 that may be null.
// It's a wrapper around sql.NullInt64 to avoid direct imports.
type NullInt64 = sql.NullInt64

// NullFloat64 represents a float64 that may be null.
// It's a wrapper around sql.NullFloat64 to avoid direct imports.
type NullFloat64 = sql.NullFloat64

// NullBool represents a bool that may be null.
// It's a wrapper around sql.NullBool to avoid direct imports.
type NullBool = sql.NullBool

// Common errors
var (
	// ErrNoRows is returned by Scan when QueryRow doesn't return a row.
	ErrNoRows = sql.ErrNoRows
	
	// ErrTxDone is returned when performing operations on a transaction that has already been committed or rolled back.
	ErrTxDone = sql.ErrTxDone
)

// Scanner is an interface used by Scan.
type Scanner = sql.Scanner

// Driver is the interface that must be implemented by a database driver.
type Driver = driver.Driver

// Result summarizes an executed SQL command.
type Result = sql.Result

// Row is the result of calling QueryRow to select a single row.
type Row = sql.Row

// Rows is the result of a query. Its cursor starts before the first row
// of the result set. Use Next to advance from row to row.
type Rows = sql.Rows

// Tx is an in-progress database transaction.
type Tx = sql.Tx

// NewNullString creates a new NullString
func NewNullString(s string, valid bool) NullString {
	return NullString{
		String: s,
		Valid:  valid,
	}
}

// NewNullTime creates a new NullTime
func NewNullTime(t time.Time, valid bool) NullTime {
	return NullTime{
		Time:  t,
		Valid: valid,
	}
}

// NewNullInt64 creates a new NullInt64
func NewNullInt64(i int64, valid bool) NullInt64 {
	return NullInt64{
		Int64: i,
		Valid: valid,
	}
}

// StringOrEmpty returns the string value or empty string if null
func StringOrEmpty(ns NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

// TimeOrZero returns the time value or zero time if null
func TimeOrZero(nt NullTime) time.Time {
	if nt.Valid {
		return nt.Time
	}
	return time.Time{}
}

// Int64OrZero returns the int64 value or zero if null
func Int64OrZero(ni NullInt64) int64 {
	if ni.Valid {
		return ni.Int64
	}
	return 0
}