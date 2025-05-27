package patterns

import (
	"fmt"
	"strings"
)

// QueryBuilder helps construct SQL queries dynamically
type QueryBuilder struct {
	selectCols []string
	from       string
	joins      []string
	where      []string
	groupBy    []string
	having     []string
	orderBy    []string
	limit      int
	offset     int
	args       []any
	argCount   int
}

// NewQueryBuilder creates a new query builder
func NewQueryBuilder() *QueryBuilder {
	return &QueryBuilder{
		selectCols: []string{},
		joins:      []string{},
		where:      []string{},
		groupBy:    []string{},
		having:     []string{},
		orderBy:    []string{},
		args:       []any{},
	}
}

// Select adds columns to SELECT clause
func (qb *QueryBuilder) Select(columns ...string) *QueryBuilder {
	qb.selectCols = append(qb.selectCols, columns...)
	return qb
}

// From sets the FROM clause
func (qb *QueryBuilder) From(table string) *QueryBuilder {
	qb.from = table
	return qb
}

// Join adds a JOIN clause
func (qb *QueryBuilder) Join(joinType, table, condition string) *QueryBuilder {
	qb.joins = append(qb.joins, fmt.Sprintf("%s JOIN %s ON %s", joinType, table, condition))
	return qb
}

// Where adds a WHERE condition
func (qb *QueryBuilder) Where(condition string, args ...any) *QueryBuilder {
	// Replace ? with positional parameters
	count := strings.Count(condition, "?")
	for i := 0; i < count; i++ {
		qb.argCount++
		condition = strings.Replace(condition, "?", fmt.Sprintf("$%d", qb.argCount), 1)
	}
	
	qb.where = append(qb.where, condition)
	qb.args = append(qb.args, args...)
	return qb
}

// WhereIn adds a WHERE IN condition
func (qb *QueryBuilder) WhereIn(column string, values []any) *QueryBuilder {
	if len(values) == 0 {
		return qb
	}
	
	placeholders := make([]string, len(values))
	for i, v := range values {
		qb.argCount++
		placeholders[i] = fmt.Sprintf("$%d", qb.argCount)
		qb.args = append(qb.args, v)
	}
	
	condition := fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ", "))
	qb.where = append(qb.where, condition)
	return qb
}

// GroupBy adds GROUP BY columns
func (qb *QueryBuilder) GroupBy(columns ...string) *QueryBuilder {
	qb.groupBy = append(qb.groupBy, columns...)
	return qb
}

// Having adds a HAVING condition
func (qb *QueryBuilder) Having(condition string, args ...any) *QueryBuilder {
	// Replace ? with positional parameters
	count := strings.Count(condition, "?")
	for i := 0; i < count; i++ {
		qb.argCount++
		condition = strings.Replace(condition, "?", fmt.Sprintf("$%d", qb.argCount), 1)
	}
	
	qb.having = append(qb.having, condition)
	qb.args = append(qb.args, args...)
	return qb
}

// OrderBy adds ORDER BY columns
func (qb *QueryBuilder) OrderBy(column string, direction string) *QueryBuilder {
	if direction == "" {
		direction = "ASC"
	}
	qb.orderBy = append(qb.orderBy, fmt.Sprintf("%s %s", column, direction))
	return qb
}

// Limit sets the LIMIT
func (qb *QueryBuilder) Limit(limit int) *QueryBuilder {
	qb.limit = limit
	return qb
}

// Offset sets the OFFSET
func (qb *QueryBuilder) Offset(offset int) *QueryBuilder {
	qb.offset = offset
	return qb
}

// Build constructs the final SQL query
func (qb *QueryBuilder) Build() (string, []any) {
	var parts []string
	
	// SELECT clause
	if len(qb.selectCols) > 0 {
		parts = append(parts, "SELECT "+strings.Join(qb.selectCols, ", "))
	} else {
		parts = append(parts, "SELECT *")
	}
	
	// FROM clause
	if qb.from != "" {
		parts = append(parts, "FROM "+qb.from)
	}
	
	// JOIN clauses
	if len(qb.joins) > 0 {
		parts = append(parts, strings.Join(qb.joins, " "))
	}
	
	// WHERE clause
	if len(qb.where) > 0 {
		parts = append(parts, "WHERE "+strings.Join(qb.where, " AND "))
	}
	
	// GROUP BY clause
	if len(qb.groupBy) > 0 {
		parts = append(parts, "GROUP BY "+strings.Join(qb.groupBy, ", "))
	}
	
	// HAVING clause
	if len(qb.having) > 0 {
		parts = append(parts, "HAVING "+strings.Join(qb.having, " AND "))
	}
	
	// ORDER BY clause
	if len(qb.orderBy) > 0 {
		parts = append(parts, "ORDER BY "+strings.Join(qb.orderBy, ", "))
	}
	
	query := strings.Join(parts, " ")
	
	// LIMIT and OFFSET
	if qb.limit > 0 {
		qb.argCount++
		query += fmt.Sprintf(" LIMIT $%d", qb.argCount)
		qb.args = append(qb.args, qb.limit)
	}
	
	if qb.offset > 0 {
		qb.argCount++
		query += fmt.Sprintf(" OFFSET $%d", qb.argCount)
		qb.args = append(qb.args, qb.offset)
	}
	
	return query, qb.args
}

// BuildCount builds a COUNT query
func (qb *QueryBuilder) BuildCount() (string, []any) {
	qbCount := &QueryBuilder{
		from:     qb.from,
		joins:    qb.joins,
		where:    qb.where,
		args:     qb.args,
		argCount: qb.argCount,
	}
	
	qbCount.selectCols = []string{"COUNT(*)"}
	query, args := qbCount.Build()
	
	// Remove ORDER BY, LIMIT, OFFSET from count query
	if idx := strings.Index(query, " ORDER BY"); idx != -1 {
		query = query[:idx]
	}
	
	return query, args
}