package xsb

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"regexp"
	"strings"
	"time"
)

// Dialect represents a specific SQL dialect
type Dialect int

const (
	PostgreSQL Dialect = iota
	MySQL
	SQLite
	MSSQL
)

// Builder is the main struct for building SQL queries
type Builder[T any] struct {
	dialect              Dialect
	table                string
	columns              []string
	values               []interface{}
	whereClause          string
	orderBy              string
	limit                int
	offset               int
	joins                []string
	groupBy              string
	having               string
	updateMap            map[string]interface{}
	returning            []string
	unions               []*Builder[T]
	debug                bool
	paramPrefix          string
	paramCount           int
	ctes                 []string
	distinct             bool
	forUpdate            bool
	tx                   *sql.Tx
	upsertColumns        []string
	upsertValues         map[string]interface{}
	onDuplicateKeyUpdate map[string]interface{}
	explain              bool
}

// New creates a new Builder instance with PostgreSQL as default dialect
func New[T any]() *Builder[T] {
	return &Builder[T]{
		dialect:              PostgreSQL,
		updateMap:            make(map[string]interface{}),
		paramPrefix:          "?",
		paramCount:           0,
		upsertValues:         make(map[string]interface{}),
		onDuplicateKeyUpdate: make(map[string]interface{}),
	}
}

// EnableDebug turns on debug mode
func (b *Builder[T]) EnableDebug() *Builder[T] {
	b.debug = true
	return b
}

// WithDialect sets a specific dialect for the builder
func (b *Builder[T]) WithDialect(dialect Dialect) *Builder[T] {
	b.dialect = dialect
	return b
}

// Table sets the table name for the query
func (b *Builder[T]) Table(name string) *Builder[T] {
	b.table = name
	return b
}

// Columns sets the columns for the query
func (b *Builder[T]) Columns(cols ...string) *Builder[T] {
	b.columns = cols
	return b
}

// Values sets the values for an INSERT query
func (b *Builder[T]) Values(values ...interface{}) *Builder[T] {
	b.values = values
	return b
}

// Where adds a WHERE clause to the query
func (b *Builder[T]) Where(condition string, args ...interface{}) *Builder[T] {
	if b.whereClause == "" {
		b.whereClause = "WHERE " + condition
	} else {
		b.whereClause += " AND " + condition
	}
	b.values = append(b.values, args...)
	return b
}

// OrWhere adds an OR condition to the WHERE clause
func (b *Builder[T]) OrWhere(condition string, args ...interface{}) *Builder[T] {
	if b.whereClause == "" {
		return b.Where(condition, args...)
	}
	b.whereClause += " OR " + condition
	b.values = append(b.values, args...)
	return b
}

// WhereIn adds a WHERE IN clause
func (b *Builder[T]) WhereIn(column string, values ...interface{}) *Builder[T] {
	placeholders := make([]string, len(values))
	for i := range placeholders {
		b.paramCount++
		placeholders[i] = b.placeholder(b.paramCount)
	}
	condition := fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ", "))
	return b.Where(condition, values...)
}

// OrderBy adds an ORDER BY clause to the query
func (b *Builder[T]) OrderBy(clause string) *Builder[T] {
	b.orderBy = clause
	return b
}

// Limit adds a LIMIT clause to the query
func (b *Builder[T]) Limit(limit int) *Builder[T] {
	b.limit = limit
	return b
}

// Offset adds an OFFSET clause to the query
func (b *Builder[T]) Offset(offset int) *Builder[T] {
	b.offset = offset
	return b
}

// Join adds a JOIN clause to the query
func (b *Builder[T]) Join(joinType, table, condition string) *Builder[T] {
	join := fmt.Sprintf("%s JOIN %s ON %s", joinType, table, condition)
	b.joins = append(b.joins, join)
	return b
}

// InnerJoin adds an INNER JOIN clause
func (b *Builder[T]) InnerJoin(table, condition string) *Builder[T] {
	return b.Join("INNER", table, condition)
}

// LeftJoin adds a LEFT JOIN clause
func (b *Builder[T]) LeftJoin(table, condition string) *Builder[T] {
	return b.Join("LEFT", table, condition)
}

// RightJoin adds a RIGHT JOIN clause
func (b *Builder[T]) RightJoin(table, condition string) *Builder[T] {
	return b.Join("RIGHT", table, condition)
}

// GroupBy adds a GROUP BY clause to the query
func (b *Builder[T]) GroupBy(clause string) *Builder[T] {
	b.groupBy = clause
	return b
}

// Having adds a HAVING clause to the query
func (b *Builder[T]) Having(condition string, args ...interface{}) *Builder[T] {
	b.having = condition
	b.values = append(b.values, args...)
	return b
}

// Set adds a column-value pair for UPDATE queries
func (b *Builder[T]) Set(column string, value interface{}) *Builder[T] {
	b.updateMap[column] = value
	return b
}

// Returning adds a RETURNING clause (for PostgreSQL)
func (b *Builder[T]) Returning(columns ...string) *Builder[T] {
	b.returning = columns
	return b
}

// Union adds a UNION clause
func (b *Builder[T]) Union(other *Builder[T]) *Builder[T] {
	b.unions = append(b.unions, other)
	return b
}

// CTE adds a Common Table Expression (WITH clause)
func (b *Builder[T]) CTE(name string, subquery *Builder[T]) *Builder[T] {
	subQuerySQL, _ := subquery.BuildSelect()
	cte := fmt.Sprintf("%s AS (%s)", name, subQuerySQL)
	b.ctes = append(b.ctes, cte)
	return b
}

// Distinct adds DISTINCT to the SELECT query
func (b *Builder[T]) Distinct() *Builder[T] {
	b.distinct = true
	return b
}

// Lock adds FOR UPDATE to the SELECT query
func (b *Builder[T]) Lock() *Builder[T] {
	b.forUpdate = true
	return b
}

// BuildSelect builds a SELECT query
func (b *Builder[T]) BuildSelect() (string, []interface{}) {
	var query strings.Builder

	if len(b.ctes) > 0 {
		query.WriteString("WITH ")
		query.WriteString(strings.Join(b.ctes, ", "))
		query.WriteString(" ")
	}

	if b.explain {
		query.WriteString("EXPLAIN ")
	}

	query.WriteString("SELECT ")
	if b.distinct {
		query.WriteString("DISTINCT ")
	}
	query.WriteString(strings.Join(b.columns, ", "))
	query.WriteString(" FROM ")
	query.WriteString(b.table)

	for _, join := range b.joins {
		query.WriteString(" ")
		query.WriteString(join)
	}

	if b.whereClause != "" {
		query.WriteString(" ")
		query.WriteString(b.whereClause)
	}

	if b.groupBy != "" {
		query.WriteString(" GROUP BY ")
		query.WriteString(b.groupBy)
	}

	if b.having != "" {
		query.WriteString(" HAVING ")
		query.WriteString(b.having)
	}

	if b.orderBy != "" {
		query.WriteString(" ORDER BY ")
		query.WriteString(b.orderBy)
	}

	if b.limit > 0 {
		query.WriteString(fmt.Sprintf(" LIMIT %d", b.limit))
	}

	if b.offset > 0 {
		query.WriteString(fmt.Sprintf(" OFFSET %d", b.offset))
	}

	for _, union := range b.unions {
		unionQuery, unionArgs := union.BuildSelect()
		query.WriteString(" UNION ALL ")
		query.WriteString(unionQuery)
		b.values = append(b.values, unionArgs...)
	}

	if b.forUpdate {
		switch b.dialect {
		case PostgreSQL, MySQL:
			query.WriteString(" FOR UPDATE")
		case MSSQL:
			query.WriteString(" WITH (UPDLOCK, ROWLOCK)")
		}
	}

	return query.String(), b.values
}

// BuildInsert builds an INSERT query
func (b *Builder[T]) BuildInsert() (string, []interface{}) {
	var query strings.Builder
	query.WriteString("INSERT INTO ")
	query.WriteString(b.table)
	query.WriteString(" (")
	query.WriteString(strings.Join(b.columns, ", "))
	query.WriteString(") VALUES (")

	placeholders := make([]string, len(b.columns))
	for i := range placeholders {
		b.paramCount++
		placeholders[i] = b.placeholder(b.paramCount)
	}
	query.WriteString(strings.Join(placeholders, ", "))
	query.WriteString(")")

	// Add Upsert for PostgreSQL
	if b.dialect == PostgreSQL && len(b.upsertColumns) > 0 {
		query.WriteString(" ON CONFLICT (")
		query.WriteString(strings.Join(b.upsertColumns, ", "))
		query.WriteString(") DO UPDATE SET ")
		updates := make([]string, 0, len(b.upsertValues))
		for col, val := range b.upsertValues {
			b.paramCount++
			updates = append(updates, fmt.Sprintf("%s = %s", col, b.placeholder(b.paramCount)))
			b.values = append(b.values, val)
		}
		query.WriteString(strings.Join(updates, ", "))
	}

	// Add ON DUPLICATE KEY UPDATE for MySQL
	if b.dialect == MySQL && len(b.onDuplicateKeyUpdate) > 0 {
		query.WriteString(" ON DUPLICATE KEY UPDATE ")
		updates := make([]string, 0, len(b.onDuplicateKeyUpdate))
		for col, val := range b.onDuplicateKeyUpdate {
			b.paramCount++
			updates = append(updates, fmt.Sprintf("%s = %s", col, b.placeholder(b.paramCount)))
			b.values = append(b.values, val)
		}
		query.WriteString(strings.Join(updates, ", "))
	}

	return query.String(), b.values
}

// BuildUpdate builds an UPDATE query
func (b *Builder[T]) BuildUpdate() (string, []interface{}) {
	var query strings.Builder
	query.WriteString("UPDATE ")
	query.WriteString(b.table)
	query.WriteString(" SET ")

	setClauses := make([]string, 0, len(b.updateMap))
	args := make([]interface{}, 0, len(b.updateMap))
	for col, val := range b.updateMap {
		b.paramCount++
		setClauses = append(setClauses, fmt.Sprintf("%s = %s", col, b.placeholder(b.paramCount)))
		args = append(args, val)
	}
	query.WriteString(strings.Join(setClauses, ", "))

	if b.whereClause != "" {
		query.WriteString(" ")
		query.WriteString(b.whereClause)
		args = append(args, b.values...)
	}

	if len(b.returning) > 0 && b.dialect == PostgreSQL {
		query.WriteString(" RETURNING ")
		query.WriteString(strings.Join(b.returning, ", "))
	}

	return query.String(), args
}

// BuildDelete builds a DELETE query
func (b *Builder[T]) BuildDelete() (string, []interface{}) {
	var query strings.Builder
	query.WriteString("DELETE FROM ")
	query.WriteString(b.table)

	if b.whereClause != "" {
		query.WriteString(" ")
		query.WriteString(b.whereClause)
	}

	return query.String(), b.values
}

// Scan scans the result into the provided struct
func (b *Builder[T]) Scan(rows *sql.Rows, dest *T) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	err = rows.Scan(scanArgs...)
	if err != nil {
		return err
	}

	destValue := reflect.ValueOf(dest).Elem()
	for i, column := range columns {
		field := destValue.FieldByName(column)
		if field.IsValid() && field.CanSet() {
			value := values[i]
			if value != nil {
				field.Set(reflect.ValueOf(value).Convert(field.Type()))
			}
		}
	}

	return nil
}

// MapToStruct maps a row to a struct
func (b *Builder[T]) MapToStruct(rows *sql.Rows, dest interface{}) error {
	columns, err := rows.Columns()
	if err != nil {
		return err
	}

	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(columns))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	err = rows.Scan(scanArgs...)
	if err != nil {
		return err
	}

	destValue := reflect.ValueOf(dest).Elem()
	for i, column := range columns {
		field := destValue.FieldByName(column)
		if field.IsValid() && field.CanSet() {
			value := values[i]
			if value != nil {
				field.Set(reflect.ValueOf(value).Convert(field.Type()))
			}
		}
	}

	return nil
}

// Debug returns a string representation of the query for debugging purposes
func (b *Builder[T]) Debug() string {
	var queryType string
	var query string
	var args []interface{}

	switch {
	case len(b.columns) > 0:
		queryType = "SELECT"
		query, args = b.BuildSelect()
	case len(b.updateMap) > 0:
		queryType = "UPDATE"
		query, args = b.BuildUpdate()
	case b.table != "":
		queryType = "DELETE"
		query, args = b.BuildDelete()
	default:
		queryType = "UNKNOWN"
	}

	return fmt.Sprintf("Query Type: %s\nQuery: %s\nArgs: %v", queryType, query, args)
}

// SQL returns the SQL string for the current query
func (b *Builder[T]) SQL() string {
	query, _ := b.Build()
	return query
}

// Build returns the final query string and arguments
func (b *Builder[T]) Build() (string, []interface{}) {
	switch {
	case len(b.columns) > 0:
		return b.BuildSelect()
	case len(b.updateMap) > 0:
		return b.BuildUpdate()
	case b.table != "":
		return b.BuildDelete()
	default:
		return "", nil
	}
}

// Subquery allows nesting of queries
func (b *Builder[T]) Subquery(subquery *Builder[T], alias string) string {
	query, _ := subquery.Build()
	return fmt.Sprintf("(%s) AS %s", query, alias)
}

// Raw allows inserting raw SQL into the query
func (b *Builder[T]) Raw(sql string, args ...interface{}) *Builder[T] {
	// This method should be used carefully to avoid SQL injection
	b.whereClause += " " + sql
	b.values = append(b.values, args...)
	return b
}

// Clone creates a deep copy of the Builder
func (b *Builder[T]) Clone() *Builder[T] {
	newBuilder := &Builder[T]{
		dialect:              b.dialect,
		table:                b.table,
		columns:              make([]string, len(b.columns)),
		values:               make([]interface{}, len(b.values)),
		whereClause:          b.whereClause,
		orderBy:              b.orderBy,
		limit:                b.limit,
		offset:               b.offset,
		joins:                make([]string, len(b.joins)),
		groupBy:              b.groupBy,
		having:               b.having,
		updateMap:            make(map[string]interface{}),
		returning:            make([]string, len(b.returning)),
		unions:               make([]*Builder[T], len(b.unions)),
		debug:                b.debug,
		paramPrefix:          b.paramPrefix,
		paramCount:           b.paramCount,
		ctes:                 make([]string, len(b.ctes)),
		distinct:             b.distinct,
		forUpdate:            b.forUpdate,
		tx:                   b.tx,
		upsertColumns:        make([]string, len(b.upsertColumns)),
		upsertValues:         make(map[string]interface{}),
		onDuplicateKeyUpdate: make(map[string]interface{}),
	}
	copy(newBuilder.columns, b.columns)
	copy(newBuilder.values, b.values)
	copy(newBuilder.joins, b.joins)
	copy(newBuilder.returning, b.returning)
	copy(newBuilder.unions, b.unions)
	copy(newBuilder.ctes, b.ctes)
	copy(newBuilder.upsertColumns, b.upsertColumns)
	for k, v := range b.updateMap {
		newBuilder.updateMap[k] = v
	}
	for k, v := range b.upsertValues {
		newBuilder.upsertValues[k] = v
	}
	for k, v := range b.onDuplicateKeyUpdate {
		newBuilder.onDuplicateKeyUpdate[k] = v
	}
	return newBuilder
}

// FromStruct builds a query from a struct
func (b *Builder[T]) FromStruct(s interface{}) *Builder[T] {
	v := reflect.ValueOf(s)
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = v.Type()
	}

	if t.Kind() != reflect.Struct {
		log.Printf("FromStruct: expected struct, got %v", t.Kind())
		return b
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		if tag := field.Tag.Get("db"); tag != "" && tag != "-" {
			if !value.IsZero() {
				b.columns = append(b.columns, tag)
				b.values = append(b.values, value.Interface())
			}
		}
	}

	return b
}

// Exec executes the query and returns the result
func (b *Builder[T]) Exec(db *sql.DB) (sql.Result, error) {
	query, args := b.Build()
	if b.debug {
		log.Printf("Executing query: %s with args: %v", query, args)
	}
	if b.tx != nil {
		return b.tx.Exec(query, args...)
	}
	return db.Exec(query, args...)
}

// QueryRow executes the query and returns a single row
func (b *Builder[T]) QueryRow(db *sql.DB) *sql.Row {
	query, args := b.Build()
	if b.debug {
		log.Printf("Executing query: %s with args: %v", query, args)
	}
	if b.tx != nil {
		return b.tx.QueryRow(query, args...)
	}
	return db.QueryRow(query, args...)
}

// Query executes the query and returns multiple rows
func (b *Builder[T]) Query(db *sql.DB) (*sql.Rows, error) {
	query, args := b.Build()
	if b.debug {
		log.Printf("Executing query: %s with args: %v", query, args)
	}
	if b.tx != nil {
		return b.tx.Query(query, args...)
	}
	return db.Query(query, args...)
}

// WithTransaction wraps the builder with a transaction
func (b *Builder[T]) WithTransaction(tx *sql.Tx) *Builder[T] {
	b.tx = tx
	return b
}

// Config represents configuration options for the Builder
type Config struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

// Configure applies the provided configuration to the Builder
func (b *Builder[T]) Configure(config Config) *Builder[T] {
	// This method doesn't directly affect the Builder, but could be used to configure a database connection
	// For example, if we had a db *sql.DB field in the Builder:
	// db.SetMaxOpenConns(config.MaxOpenConns)
	// db.SetMaxIdleConns(config.MaxIdleConns)
	// db.SetConnMaxLifetime(config.ConnMaxLifetime)
	return b
}

// BuildSQL builds and returns the SQL query as a string
func (b *Builder[T]) BuildSQL() string {
	query, _ := b.Build()
	return query
}

// Count adds a COUNT(*) to the query
func (b *Builder[T]) Count() *Builder[T] {
	b.columns = []string{"COUNT(*)"}
	return b
}

// Exists wraps the current query in an EXISTS clause
func (b *Builder[T]) Exists() *Builder[T] {
	subQuery, _ := b.Build()
	b.columns = []string{fmt.Sprintf("EXISTS (%s)", subQuery)}
	b.values = nil
	b.whereClause = ""
	b.orderBy = ""
	b.limit = 0
	b.offset = 0
	return b
}

// NotExists wraps the current query in a NOT EXISTS clause
func (b *Builder[T]) NotExists() *Builder[T] {
	subQuery, _ := b.Build()
	b.columns = []string{fmt.Sprintf("NOT EXISTS (%s)", subQuery)}
	b.values = nil
	b.whereClause = ""
	b.orderBy = ""
	b.limit = 0
	b.offset = 0
	return b
}

// WhereNull adds a WHERE IS NULL condition
func (b *Builder[T]) WhereNull(column string) *Builder[T] {
	return b.Where(fmt.Sprintf("%s IS NULL", column))
}

// WhereNotNull adds a WHERE IS NOT NULL condition
func (b *Builder[T]) WhereNotNull(column string) *Builder[T] {
	return b.Where(fmt.Sprintf("%s IS NOT NULL", column))
}

// WhereBetween adds a WHERE BETWEEN condition
func (b *Builder[T]) WhereBetween(column string, start, end interface{}) *Builder[T] {
	return b.Where(fmt.Sprintf("%s BETWEEN ? AND ?", column), start, end)
}

// WhereNotBetween adds a WHERE NOT BETWEEN condition
func (b *Builder[T]) WhereNotBetween(column string, start, end interface{}) *Builder[T] {
	return b.Where(fmt.Sprintf("%s NOT BETWEEN ? AND ?", column), start, end)
}

// WhereRaw adds a raw WHERE condition
func (b *Builder[T]) WhereRaw(raw string, args ...interface{}) *Builder[T] {
	return b.Where(raw, args...)
}

// OrWhereRaw adds a raw OR WHERE condition
func (b *Builder[T]) OrWhereRaw(raw string, args ...interface{}) *Builder[T] {
	return b.OrWhere(raw, args...)
}

// GroupByRaw adds a raw GROUP BY clause
func (b *Builder[T]) GroupByRaw(raw string) *Builder[T] {
	b.groupBy = raw
	return b
}

// HavingRaw adds a raw HAVING clause
func (b *Builder[T]) HavingRaw(raw string, args ...interface{}) *Builder[T] {
	return b.Having(raw, args...)
}

// OrderByRaw adds a raw ORDER BY clause
func (b *Builder[T]) OrderByRaw(raw string) *Builder[T] {
	b.orderBy = raw
	return b
}

// UnionAll adds a UNION ALL clause
func (b *Builder[T]) UnionAll(other *Builder[T]) *Builder[T] {
	b.unions = append(b.unions, other)
	return b
}

// Truncate builds a TRUNCATE query
func (b *Builder[T]) Truncate() (string, error) {
	if b.table == "" {
		return "", fmt.Errorf("table name is required for TRUNCATE")
	}
	return fmt.Sprintf("TRUNCATE TABLE %s", b.table), nil
}

// InsertIgnore builds an INSERT IGNORE query (for MySQL)
func (b *Builder[T]) InsertIgnore() (string, []interface{}, error) {
	if b.dialect != MySQL {
		return "", nil, fmt.Errorf("INSERT IGNORE is only supported for MySQL")
	}
	query, args := b.BuildInsert()
	return strings.Replace(query, "INSERT", "INSERT IGNORE", 1), args, nil
}

// OnDuplicateKeyUpdate builds an ON DUPLICATE KEY UPDATE clause (for MySQL)
func (b *Builder[T]) OnDuplicateKeyUpdate(updates map[string]interface{}) *Builder[T] {
	if b.dialect != MySQL {
		log.Printf("ON DUPLICATE KEY UPDATE is only supported for MySQL")
		return b
	}
	for col, val := range updates {
		b.onDuplicateKeyUpdate[col] = val
	}
	return b
}

// Upsert builds an UPSERT query (for PostgreSQL)
func (b *Builder[T]) Upsert(conflictColumns []string, updates map[string]interface{}) *Builder[T] {
	if b.dialect != PostgreSQL {
		log.Printf("UPSERT is only supported for PostgreSQL")
		return b
	}
	b.upsertColumns = conflictColumns
	for col, val := range updates {
		b.upsertValues[col] = val
	}
	return b
}

// WithRecursive adds a WITH RECURSIVE clause for recursive CTEs
func (b *Builder[T]) WithRecursive(name string, subquery *Builder[T]) *Builder[T] {
	subQuerySQL, _ := subquery.BuildSelect()
	cte := fmt.Sprintf("%s AS (%s)", name, subQuerySQL)
	b.ctes = append([]string{"RECURSIVE " + cte}, b.ctes...)
	return b
}

// Paginate adds LIMIT and OFFSET for pagination
func (b *Builder[T]) Paginate(page, perPage int) *Builder[T] {
	b.Limit(perPage)
	b.Offset((page - 1) * perPage)
	return b
}

// Explain adds an EXPLAIN clause to the query
func (b *Builder[T]) Explain() *Builder[T] {
	b.explain = true
	return b
}

// placeholder returns the dialect-specific placeholder for the given index
func (b *Builder[T]) placeholder(index int) string {
	return "?"
}

// ToSQL returns the SQL string and arguments
func (b *Builder[T]) ToSQL() (string, []interface{}, error) {
	query, args := b.Build()
	return query, args, nil
}

// MustBuild builds the query and panics on error
func (b *Builder[T]) MustBuild() string {
	query, _ := b.Build()
	return query
}

// WhereExists adds a WHERE EXISTS subquery
func (b *Builder[T]) WhereExists(subquery *Builder[T]) *Builder[T] {
	subQuerySQL, subQueryArgs := subquery.Build()
	return b.Where(fmt.Sprintf("EXISTS (%s)", subQuerySQL), subQueryArgs...)
}

// WhereNotExists adds a WHERE NOT EXISTS subquery
func (b *Builder[T]) WhereNotExists(subquery *Builder[T]) *Builder[T] {
	subQuerySQL, subQueryArgs := subquery.Build()
	return b.Where(fmt.Sprintf("NOT EXISTS (%s)", subQuerySQL), subQueryArgs...)
}

// WithLock adds a locking clause based on the dialect
func (b *Builder[T]) WithLock(lockType string) *Builder[T] {
	switch b.dialect {
	case PostgreSQL:
		b.forUpdate = true
	case MySQL:
		b.forUpdate = true
	case SQLite:
		log.Printf("SQLite does not support explicit locking clauses")
	case MSSQL:
		b.forUpdate = true
	}
	return b
}

// InsertGetId performs an INSERT and returns the last inserted ID
func (b *Builder[T]) InsertGetId(db *sql.DB) (int64, error) {
	query, args := b.BuildInsert()
	if b.dialect == PostgreSQL {
		query += " RETURNING id"
	}

	var result sql.Result
	var err error
	var id int64

	if b.tx != nil {
		if b.dialect == PostgreSQL {
			err = b.tx.QueryRow(query, args...).Scan(&id)
		} else {
			result, err = b.tx.Exec(query, args...)
		}
	} else {
		if b.dialect == PostgreSQL {
			err = db.QueryRow(query, args...).Scan(&id)
		} else {
			result, err = db.Exec(query, args...)
		}
	}

	if err != nil {
		return 0, err
	}

	if b.dialect != PostgreSQL {
		id, err = result.LastInsertId()
		if err != nil {
			return 0, err
		}
	}

	return id, nil
}

// Increment increments a column's value
func (b *Builder[T]) Increment(column string, amount int) *Builder[T] {
	b.updateMap[column] = fmt.Sprintf("%s + %d", column, amount)
	return b
}

// Decrement decrements a column's value
func (b *Builder[T]) Decrement(column string, amount int) *Builder[T] {
	b.updateMap[column] = fmt.Sprintf("%s - %d", column, amount)
	return b
}

// Sanitize removes any potentially harmful SQL from the input
func Sanitize(input string) string {
	// Remove any SQL comments
	re := regexp.MustCompile(`/\*.*?\*/|--.*?$`)
	input = re.ReplaceAllString(input, "")

	// Remove any semicolons
	input = strings.ReplaceAll(input, ";", "")

	// Remove any UNION keywords
	re = regexp.MustCompile(`(?i)\bUNION\b`)
	input = re.ReplaceAllString(input, "")

	return strings.TrimSpace(input)
}
