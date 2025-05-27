package patterns

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

// CRUDOperations defines generic CRUD operations for database entities
type CRUDOperations[T any] interface {
	Create(ctx context.Context, entity *T) error
	Get(ctx context.Context, id string) (*T, error)
	List(ctx context.Context, filters map[string]any) ([]*T, error)
	Update(ctx context.Context, id string, updates map[string]any) error
	Delete(ctx context.Context, id string) error
}

// BaseCRUD provides common CRUD functionality
type BaseCRUD[T any] struct {
	DB        *sql.DB
	TableName string
	IDField   string
}

// ScanFunc is a function that scans a row into an entity
type ScanFunc[T any] func(rows *sql.Rows) (*T, error)

// BuildQueryFilter builds WHERE clause from filters
func BuildQueryFilter(filters map[string]any, startArg int) (string, []any) {
	if len(filters) == 0 {
		return "", nil
	}
	
	var conditions []string
	var args []any
	argCount := startArg
	
	for key, value := range filters {
		switch v := value.(type) {
		case nil:
			conditions = append(conditions, fmt.Sprintf("%s IS NULL", key))
		case []string, []int, []int64:
			// Handle IN clauses
			rv := reflect.ValueOf(v)
			placeholders := make([]string, rv.Len())
			for i := 0; i < rv.Len(); i++ {
				argCount++
				placeholders[i] = fmt.Sprintf("$%d", argCount)
				args = append(args, rv.Index(i).Interface())
			}
			conditions = append(conditions, fmt.Sprintf("%s IN (%s)", key, strings.Join(placeholders, ",")))
		default:
			argCount++
			conditions = append(conditions, fmt.Sprintf("%s = $%d", key, argCount))
			args = append(args, value)
		}
	}
	
	return " WHERE " + strings.Join(conditions, " AND "), args
}

// Paginate adds pagination to a query
func Paginate(query string, page, pageSize int, args []any) (string, []any) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	
	offset := (page - 1) * pageSize
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, pageSize, offset)
	
	return query, args
}

// CountRows counts total rows for pagination
func CountRows(ctx context.Context, db *sql.DB, baseQuery string, args []any) (int, error) {
	// Extract the FROM clause
	fromIndex := strings.Index(strings.ToUpper(baseQuery), "FROM")
	if fromIndex == -1 {
		return 0, fmt.Errorf("invalid query: no FROM clause found")
	}
	
	// Remove ORDER BY, LIMIT, OFFSET if present
	countQuery := baseQuery[fromIndex:]
	if orderIndex := strings.LastIndex(strings.ToUpper(countQuery), "ORDER BY"); orderIndex != -1 {
		countQuery = countQuery[:orderIndex]
	}
	
	countQuery = "SELECT COUNT(*) " + countQuery
	
	var count int
	err := db.QueryRowContext(ctx, countQuery, args...).Scan(&count)
	return count, err
}

// WithTransaction executes a function within a database transaction
func WithTransaction(ctx context.Context, db *sql.DB, fn func(*sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	
	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to rollback: %v (original error: %w)", rbErr, err)
		}
		return err
	}
	
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	
	return nil
}

// NullableTime converts a *time.Time to sql.NullTime
func NullableTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

// NullableString converts a string to sql.NullString
func NullableString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// TimePointer converts sql.NullTime to *time.Time
func TimePointer(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}

// StringValue converts sql.NullString to string
func StringValue(ns sql.NullString) string {
	if !ns.Valid {
		return ""
	}
	return ns.String
}