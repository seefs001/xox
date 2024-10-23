package xsb

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/seefs001/xox/xlog"
)

// Dialect represents a specific SQL dialect
type Dialect int

const (
	PostgreSQL Dialect = iota
	MySQL
	SQLite
	MSSQL
)

// RawExpr represents a raw SQL expression.
type RawExpr struct {
	Expr string
	args []interface{}
}

// UpdateClause represents a column-value pair for updates.
type UpdateClause struct {
	Column string
	Value  interface{}
}

// CTE represents a Common Table Expression with its SQL and args.
type CTE struct {
	Name  string
	Query string
	Args  []interface{}
}

// Builder is the main struct for building SQL queries
type Builder struct {
	dialect              Dialect
	table                string
	columns              []string
	values               []interface{}
	whereClause          string
	whereArgs            []interface{}
	orderBy              string
	limit                int
	offset               int
	joins                []string
	groupBy              string
	having               string
	havingArgs           []interface{}
	updateClauses        []UpdateClause
	returning            []string
	unions               []*Builder
	unionAll             bool
	debug                bool
	paramCount           int
	ctes                 []CTE
	distinct             bool
	forUpdate            bool
	tx                   *sql.Tx
	upsertColumns        []string
	upsertValues         []UpdateClause
	onDuplicateKeyUpdate []UpdateClause
	explain              bool
	ctx                  context.Context
	err                  error
	logSQL               bool
	allowEmptyWhere      bool
}

// New creates a new Builder instance with PostgreSQL as default dialect
func New() *Builder {
	return &Builder{
		dialect:              PostgreSQL,
		columns:              make([]string, 0),
		values:               make([]interface{}, 0),
		whereArgs:            make([]interface{}, 0),
		havingArgs:           make([]interface{}, 0),
		updateClauses:        make([]UpdateClause, 0),
		upsertValues:         make([]UpdateClause, 0),
		onDuplicateKeyUpdate: make([]UpdateClause, 0),
		joins:                make([]string, 0),
		ctes:                 make([]CTE, 0),
		unions:               make([]*Builder, 0),
		upsertColumns:        make([]string, 0),
		returning:            make([]string, 0),
		ctx:                  context.Background(),
	}
}

// EnableDebug turns on debug mode
func (b *Builder) EnableDebug() *Builder {
	b.debug = true
	return b
}

// WithDialect sets a specific dialect for the builder
func (b *Builder) WithDialect(dialect Dialect) *Builder {
	b.dialect = dialect
	return b
}

// Table sets the table name for the query
func (b *Builder) Table(name string) *Builder {
	b.table = name
	return b
}

// Columns sets the columns for the query
func (b *Builder) Columns(cols ...string) *Builder {
	b.columns = cols
	return b
}

// Values sets the values for an INSERT query
func (b *Builder) Values(values ...interface{}) *Builder {
	b.values = values
	return b
}

// Where adds a WHERE clause to the query
func (b *Builder) Where(condition string, args ...interface{}) *Builder {
	if b.whereClause == "" {
		b.whereClause = condition
	} else {
		b.whereClause += " AND " + condition
	}
	b.whereArgs = append(b.whereArgs, args...)
	return b
}

// OrWhere adds an OR condition to the WHERE clause
func (b *Builder) OrWhere(condition string, args ...interface{}) *Builder {
	if b.whereClause == "" {
		return b.Where(condition, args...)
	}
	b.whereClause += " OR " + condition
	b.whereArgs = append(b.whereArgs, args...)
	return b
}

// WhereIn adds a WHERE IN clause
func (b *Builder) WhereIn(column string, values ...interface{}) *Builder {
	placeholders := make([]string, len(values))
	for i := range placeholders {
		placeholders[i] = b.placeholder()
	}
	condition := fmt.Sprintf("%s IN (%s)", column, strings.Join(placeholders, ", "))
	return b.Where(condition, values...)
}

// OrderBy adds an ORDER BY clause to the query
func (b *Builder) OrderBy(clause string) *Builder {
	b.orderBy = clause
	return b
}

// Limit adds a LIMIT clause to the query
func (b *Builder) Limit(limit int) *Builder {
	b.limit = limit
	return b
}

// Offset adds an OFFSET clause to the query
func (b *Builder) Offset(offset int) *Builder {
	b.offset = offset
	return b
}

// Join adds a JOIN clause to the query
func (b *Builder) Join(joinType, table, condition string) *Builder {
	join := fmt.Sprintf("%s JOIN %s ON %s", joinType, table, condition)
	b.joins = append(b.joins, join)
	return b
}

// InnerJoin adds an INNER JOIN clause
func (b *Builder) InnerJoin(table, condition string) *Builder {
	return b.Join("INNER", table, condition)
}

// LeftJoin adds a LEFT JOIN clause
func (b *Builder) LeftJoin(table, condition string) *Builder {
	return b.Join("LEFT", table, condition)
}

// RightJoin adds a RIGHT JOIN clause
func (b *Builder) RightJoin(table, condition string) *Builder {
	return b.Join("RIGHT", table, condition)
}

// GroupBy adds a GROUP BY clause to the query
func (b *Builder) GroupBy(clause string) *Builder {
	b.groupBy = clause
	return b
}

// Having adds a HAVING clause to the query
func (b *Builder) Having(condition string, args ...interface{}) *Builder {
	b.having = condition
	b.havingArgs = append(b.havingArgs, args...)
	return b
}

// Set adds a column-value pair for UPDATE queries
func (b *Builder) Set(column string, value interface{}) *Builder {
	b.updateClauses = append(b.updateClauses, UpdateClause{Column: column, Value: value})
	return b
}

// Returning adds a RETURNING clause (for PostgreSQL)
func (b *Builder) Returning(columns ...string) *Builder {
	b.returning = columns
	return b
}

// Union adds a UNION clause
func (b *Builder) Union(other *Builder) *Builder {
	b.unions = append(b.unions, other)
	return b
}

// CTE adds a Common Table Expression (WITH clause)
func (b *Builder) CTE(name string, subquery *Builder) *Builder {
	subQuerySQL, subQueryArgs := subquery.BuildSelect()
	cte := CTE{
		Name:  name,
		Query: subQuerySQL,
		Args:  subQueryArgs,
	}
	b.ctes = append(b.ctes, cte)
	return b
}

// Distinct adds DISTINCT to the SELECT query
func (b *Builder) Distinct() *Builder {
	b.distinct = true
	return b
}

// Lock adds FOR UPDATE to the SELECT query
func (b *Builder) Lock() *Builder {
	b.forUpdate = true
	return b
}

// BuildSelect builds a SELECT query
func (b *Builder) BuildSelect() (string, []interface{}) {
	var query strings.Builder
	var args []interface{}

	if len(b.ctes) > 0 {
		query.WriteString("WITH ")
		if strings.HasPrefix(b.ctes[0].Name, "RECURSIVE") {
			query.WriteString("RECURSIVE ")
			b.ctes[0].Name = strings.TrimPrefix(b.ctes[0].Name, "RECURSIVE ")
		}
		var cteStrings []string
		for _, cte := range b.ctes {
			cteStrings = append(cteStrings, fmt.Sprintf("%s AS (%s)", cte.Name, cte.Query))
			args = append(args, cte.Args...)
		}
		query.WriteString(strings.Join(cteStrings, ", "))
		query.WriteString(" ")
	}

	if b.explain {
		query.WriteString("EXPLAIN ")
	}

	query.WriteString("SELECT ")
	if b.distinct {
		query.WriteString("DISTINCT ")
	}

	if len(b.columns) > 0 {
		query.WriteString(strings.Join(b.columns, ", "))
	} else {
		query.WriteString("*")
	}

	query.WriteString(" FROM ")
	query.WriteString(b.table)

	// Insert lock hint after main table if MSSQL
	if b.forUpdate && b.dialect == MSSQL {
		query.WriteString(" WITH (UPDLOCK, ROWLOCK)")
	}

	for _, join := range b.joins {
		query.WriteString(" ")
		query.WriteString(join)
	}

	// Append locking hints for other dialects after joins
	if b.forUpdate && b.dialect != MSSQL {
		switch b.dialect {
		case PostgreSQL, MySQL:
			query.WriteString(" FOR UPDATE")
		}
	}

	if b.whereClause != "" {
		query.WriteString(" WHERE ")
		query.WriteString(b.whereClause)
		args = append(args, b.whereArgs...)
	}

	if b.groupBy != "" {
		query.WriteString(" GROUP BY ")
		query.WriteString(b.groupBy)
	}

	if b.having != "" {
		query.WriteString(" HAVING ")
		query.WriteString(b.having)
		args = append(args, b.havingArgs...)
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

	if len(b.unions) > 0 {
		for _, union := range b.unions {
			unionType := "UNION"
			if b.unionAll {
				unionType = "UNION ALL"
			}
			unionQuery, unionArgs := union.BuildSelect()
			query.WriteString(" ")
			query.WriteString(unionType)
			query.WriteString(" ")
			query.WriteString(unionQuery)
			args = append(args, unionArgs...)
		}
	}

	return query.String(), args
}

// BuildInsert builds an INSERT query
func (b *Builder) BuildInsert() (string, []interface{}) {
	var query strings.Builder
	var args []interface{}

	// Handle CTEs
	if len(b.ctes) > 0 {
		query.WriteString("WITH ")
		var cteStrings []string
		for _, cte := range b.ctes {
			cteStrings = append(cteStrings, fmt.Sprintf("%s AS (%s)", cte.Name, cte.Query))
			args = append(args, cte.Args...)
		}
		query.WriteString(strings.Join(cteStrings, ", "))
		query.WriteString(" ")
	}

	query.WriteString("INSERT INTO ")
	query.WriteString(b.table)
	query.WriteString(" (")

	query.WriteString(strings.Join(b.columns, ", "))
	query.WriteString(") ")

	if len(b.values) > 0 {
		if len(b.values) == 1 {
			if raw, ok := b.values[0].(RawExpr); ok {
				query.WriteString(raw.Expr)
				args = append(args, raw.args...)
			} else {
				query.WriteString("VALUES (")
				placeholders := make([]string, len(b.values))
				for i := range placeholders {
					placeholders[i] = b.placeholder()
				}
				query.WriteString(strings.Join(placeholders, ", "))
				query.WriteString(")")
				args = append(args, b.values...)
			}
		} else {
			query.WriteString("VALUES (")
			placeholders := make([]string, len(b.values))
			for i := range placeholders {
				placeholders[i] = b.placeholder()
			}
			query.WriteString(strings.Join(placeholders, ", "))
			query.WriteString(")")
			args = append(args, b.values...)
		}
	}

	// Handle ON DUPLICATE KEY UPDATE for MySQL
	if len(b.onDuplicateKeyUpdate) > 0 && b.dialect == MySQL {
		query.WriteString(" ON DUPLICATE KEY UPDATE ")

		// Iterate through the slice to preserve order
		updates := make([]string, 0, len(b.onDuplicateKeyUpdate))
		for _, update := range b.onDuplicateKeyUpdate {
			updates = append(updates, fmt.Sprintf("%s = %s", update.Column, b.placeholder()))
			args = append(args, update.Value)
		}

		query.WriteString(strings.Join(updates, ", "))
	} else if len(b.upsertColumns) > 0 && b.dialect == PostgreSQL {
		query.WriteString(" ON CONFLICT (")
		query.WriteString(strings.Join(b.upsertColumns, ", "))
		query.WriteString(") DO UPDATE SET ")
		updates := make([]string, 0, len(b.upsertValues))

		// Iterate through the slice to preserve order
		for _, update := range b.upsertValues {
			updates = append(updates, fmt.Sprintf("%s = %s", update.Column, b.placeholder()))
			args = append(args, update.Value)
		}

		query.WriteString(strings.Join(updates, ", "))
	}

	// **Always Append "RETURNING id" for PostgreSQL Inserts**
	if b.dialect == PostgreSQL {
		query.WriteString(" RETURNING id")
	}

	return query.String(), args
}

// BuildUpdate builds an UPDATE query
func (b *Builder) BuildUpdate() (string, []interface{}) {
	var query strings.Builder
	var args []interface{}

	query.WriteString("UPDATE ")
	query.WriteString(b.table)
	query.WriteString(" SET ")

	setClauses := make([]string, 0, len(b.updateClauses))
	for _, clause := range b.updateClauses {
		switch v := clause.Value.(type) {
		case RawExpr:
			setClauses = append(setClauses, fmt.Sprintf("%s = %s", clause.Column, v.Expr))
			args = append(args, v.args...)
		default:
			setClauses = append(setClauses, fmt.Sprintf("%s = %s", clause.Column, b.placeholder()))
			args = append(args, v)
		}
	}

	query.WriteString(strings.Join(setClauses, ", "))

	if b.whereClause != "" {
		query.WriteString(" WHERE ")
		query.WriteString(b.whereClause)
		args = append(args, b.whereArgs...)
	}

	if len(b.returning) > 0 && b.dialect == PostgreSQL {
		query.WriteString(" RETURNING ")
		query.WriteString(strings.Join(b.returning, ", "))
	}

	return query.String(), args
}

// BuildDelete builds a DELETE query
func (b *Builder) BuildDelete() (string, []interface{}) {
	var query strings.Builder
	var args []interface{}

	query.WriteString("DELETE FROM ")
	query.WriteString(b.table)

	if b.whereClause != "" {
		query.WriteString(" WHERE ")
		query.WriteString(b.whereClause)
		args = append(args, b.whereArgs...)
	}

	return query.String(), args
}

// Scan scans the result into the provided struct
func (b *Builder) Scan(rows *sql.Rows, dest interface{}) error {
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
func (b *Builder) MapToStruct(rows *sql.Rows, dest interface{}) error {
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
func (b *Builder) Debug() string {
	var queryType string
	var query string
	var args []interface{}

	switch {
	case len(b.columns) > 0:
		queryType = "SELECT"
		query, args = b.BuildSelect()
	case len(b.updateClauses) > 0:
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
func (b *Builder) SQL() string {
	query, _ := b.Build()
	return query
}

// Build returns the final query string and arguments
func (b *Builder) Build() (string, []interface{}) {
	if b.err != nil {
		return "", nil
	}

	if b.table == "" {
		b.err = fmt.Errorf("table name is required")
		return "", nil
	}

	switch {
	case len(b.columns) > 0:
		return b.BuildSelect()
	case len(b.updateClauses) > 0:
		return b.BuildUpdate()
	default:
		return b.BuildDelete()
	}
}

// Subquery allows nesting of queries
func (b *Builder) Subquery(subquery *Builder, alias string) string {
	query, _ := subquery.BuildSelect()
	return fmt.Sprintf("(%s) AS %s", query, alias)
}

// Raw allows inserting raw SQL into the query
func (b *Builder) Raw(sql string, args ...interface{}) *Builder {
	// This method should be used carefully to avoid SQL injection
	if b.whereClause == "" {
		b.whereClause = sql
	} else {
		b.whereClause += " " + sql
	}
	b.whereArgs = append(b.whereArgs, args...)
	return b
}

// Clone creates a deep copy of the Builder
func (b *Builder) Clone() *Builder {
	newBuilder := &Builder{
		dialect:              b.dialect,
		table:                b.table,
		columns:              make([]string, len(b.columns)),
		values:               make([]interface{}, len(b.values)),
		whereClause:          b.whereClause,
		whereArgs:            make([]interface{}, len(b.whereArgs)),
		orderBy:              b.orderBy,
		limit:                b.limit,
		offset:               b.offset,
		joins:                make([]string, len(b.joins)),
		groupBy:              b.groupBy,
		having:               b.having,
		havingArgs:           make([]interface{}, len(b.havingArgs)),
		updateClauses:        make([]UpdateClause, len(b.updateClauses)),
		returning:            make([]string, len(b.returning)),
		unions:               make([]*Builder, len(b.unions)),
		unionAll:             b.unionAll,
		debug:                b.debug,
		paramCount:           b.paramCount,
		ctes:                 make([]CTE, len(b.ctes)),
		distinct:             b.distinct,
		forUpdate:            b.forUpdate,
		tx:                   b.tx,
		upsertColumns:        make([]string, len(b.upsertColumns)),
		upsertValues:         make([]UpdateClause, len(b.upsertValues)),
		onDuplicateKeyUpdate: make([]UpdateClause, len(b.onDuplicateKeyUpdate)),
		explain:              b.explain,
		ctx:                  b.ctx,
		err:                  b.err,
		logSQL:               b.logSQL,
		allowEmptyWhere:      b.allowEmptyWhere,
	}
	copy(newBuilder.columns, b.columns)
	copy(newBuilder.values, b.values)
	copy(newBuilder.joins, b.joins)
	copy(newBuilder.returning, b.returning)
	copy(newBuilder.unions, b.unions)
	copy(newBuilder.ctes, b.ctes)
	copy(newBuilder.upsertColumns, b.upsertColumns)
	copy(newBuilder.upsertValues, b.upsertValues)
	copy(newBuilder.onDuplicateKeyUpdate, b.onDuplicateKeyUpdate)
	for i, clause := range b.updateClauses {
		newBuilder.updateClauses[i] = clause
	}
	return newBuilder
}

// FromStruct builds a query from a struct
func (b *Builder) FromStruct(s interface{}) *Builder {
	v := reflect.ValueOf(s)
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = v.Type()
	}

	if t.Kind() != reflect.Struct {
		xlog.Debugf("FromStruct: expected struct, got %v", t.Kind())
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
func (b *Builder) Exec(db *sql.DB) (sql.Result, error) {
	if b.err != nil {
		return nil, b.err
	}

	query, args := b.Build()
	if b.logSQL {
		xlog.Debugf("[SQL] %s %v", query, args)
	}

	if b.tx != nil {
		return b.tx.ExecContext(b.ctx, query, args...)
	}
	return db.ExecContext(b.ctx, query, args...)
}

// QueryRow executes the query and returns a single row
func (b *Builder) QueryRow(db *sql.DB) *sql.Row {
	query, args := b.Build()
	if b.logSQL {
		xlog.Debugf("[SQL] %s %v", query, args)
	}

	if b.tx != nil {
		return b.tx.QueryRowContext(b.ctx, query, args...)
	}
	return db.QueryRowContext(b.ctx, query, args...)
}

// Query executes the query and returns multiple rows
func (b *Builder) Query(db *sql.DB) (*sql.Rows, error) {
	if b.err != nil {
		return nil, b.err
	}

	query, args := b.Build()
	if b.logSQL {
		xlog.Debugf("[SQL] %s %v", query, args)
	}

	if b.tx != nil {
		return b.tx.QueryContext(b.ctx, query, args...)
	}
	return db.QueryContext(b.ctx, query, args...)
}

// WithTransaction wraps the builder with a transaction
func (b *Builder) WithTransaction(tx *sql.Tx) *Builder {
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
func (b *Builder) Configure(config Config) *Builder {
	// This method doesn't directly affect the Builder, but could be used to configure a database connection
	// For example, if we had a db *sql.DB field in the Builder:
	// db.SetMaxOpenConns(config.MaxOpenConns)
	// db.SetMaxIdleConns(config.MaxIdleConns)
	// db.SetConnMaxLifetime(config.ConnMaxLifetime)
	return b
}

// BuildSQL builds and returns the SQL query as a string
func (b *Builder) BuildSQL() string {
	query, _ := b.Build()
	return query
}

// Count adds a COUNT(*) to the query
func (b *Builder) Count() *Builder {
	b.columns = []string{"COUNT(*)"}
	return b
}

// Exists wraps the current query in an EXISTS clause
func (b *Builder) Exists() *Builder {
	subQuery, _ := b.Build()
	b.columns = []string{fmt.Sprintf("EXISTS (%s)", subQuery)}
	b.values = nil
	b.whereClause = ""
	b.whereArgs = nil
	b.orderBy = ""
	b.limit = 0
	b.offset = 0
	return b
}

// NotExists wraps the current query in a NOT EXISTS clause
func (b *Builder) NotExists() *Builder {
	subQuery, _ := b.Build()
	b.columns = []string{fmt.Sprintf("NOT EXISTS (%s)", subQuery)}
	b.values = nil
	b.whereClause = ""
	b.whereArgs = nil
	b.orderBy = ""
	b.limit = 0
	b.offset = 0
	return b
}

// WhereNull adds a WHERE IS NULL condition
func (b *Builder) WhereNull(column string) *Builder {
	return b.Where(fmt.Sprintf("%s IS NULL", column))
}

// WhereNotNull adds a WHERE IS NOT NULL condition
func (b *Builder) WhereNotNull(column string) *Builder {
	return b.Where(fmt.Sprintf("%s IS NOT NULL", column))
}

// WhereBetween adds a WHERE BETWEEN condition
func (b *Builder) WhereBetween(column string, start, end interface{}) *Builder {
	return b.Where(fmt.Sprintf("%s BETWEEN %s AND %s", column, b.placeholder(), b.placeholder()), start, end)
}

// WhereNotBetween adds a WHERE NOT BETWEEN condition
func (b *Builder) WhereNotBetween(column string, start, end interface{}) *Builder {
	return b.Where(fmt.Sprintf("%s NOT BETWEEN %s AND %s", column, b.placeholder(), b.placeholder()), start, end)
}

// WhereRaw adds a raw WHERE condition
func (b *Builder) WhereRaw(raw string, args ...interface{}) *Builder {
	return b.Where(raw, args...)
}

// OrWhereRaw adds a raw OR WHERE condition
func (b *Builder) OrWhereRaw(raw string, args ...interface{}) *Builder {
	return b.OrWhere(raw, args...)
}

// GroupByRaw adds a raw GROUP BY clause
func (b *Builder) GroupByRaw(raw string) *Builder {
	b.groupBy = raw
	return b
}

// HavingRaw adds a raw HAVING clause
func (b *Builder) HavingRaw(raw string, args ...interface{}) *Builder {
	return b.Having(raw, args...)
}

// OrderByRaw adds a raw ORDER BY clause
func (b *Builder) OrderByRaw(raw string) *Builder {
	b.orderBy = raw
	return b
}

// UnionAll adds a UNION ALL clause
func (b *Builder) UnionAll(other *Builder) *Builder {
	b.unions = append(b.unions, other)
	b.unionAll = true
	return b
}

// Truncate builds a TRUNCATE query
func (b *Builder) Truncate() (string, error) {
	if b.table == "" {
		return "", fmt.Errorf("table name is required for TRUNCATE")
	}
	return fmt.Sprintf("TRUNCATE TABLE %s", b.table), nil
}

// InsertIgnore builds an INSERT IGNORE query (for MySQL)
func (b *Builder) InsertIgnore() (string, []interface{}, error) {
	if b.dialect != MySQL {
		return "", nil, fmt.Errorf("INSERT IGNORE is only supported for MySQL")
	}
	query, args := b.BuildInsert()
	return strings.Replace(query, "INSERT", "INSERT IGNORE", 1), args, nil
}

// OnDuplicateKeyUpdate adds an ON DUPLICATE KEY UPDATE clause for MySQL
func (b *Builder) OnDuplicateKeyUpdate(updates []UpdateClause) *Builder {
	if b.dialect != MySQL {
		xlog.Debugf("ON DUPLICATE KEY UPDATE is only supported for MySQL")
		return b
	}
	b.onDuplicateKeyUpdate = append(b.onDuplicateKeyUpdate, updates...)
	return b
}

// Upsert adds an UPSERT clause for PostgreSQL
func (b *Builder) Upsert(conflictColumns []string, updates []UpdateClause) *Builder {
	if b.dialect != PostgreSQL {
		xlog.Debugf("UPSERT is only supported for PostgreSQL")
		return b
	}
	b.upsertColumns = conflictColumns
	b.upsertValues = append(b.upsertValues, updates...)
	return b
}

// WithRecursive adds a WITH RECURSIVE clause for recursive CTEs
func (b *Builder) WithRecursive(name string, subquery *Builder) *Builder {
	subQuerySQL, subQueryArgs := subquery.BuildSelect()
	cte := CTE{
		Name:  "RECURSIVE " + name,
		Query: subQuerySQL,
		Args:  subQueryArgs,
	}
	b.ctes = append([]CTE{cte}, b.ctes...)
	return b
}

// Paginate adds LIMIT and OFFSET for pagination
func (b *Builder) Paginate(page, perPage int) *Builder {
	b.Limit(perPage)
	b.Offset((page - 1) * perPage)
	return b
}

// Explain adds an EXPLAIN clause to the query
func (b *Builder) Explain() *Builder {
	b.explain = true
	return b
}

// placeholder returns the dialect-specific placeholder for the given index
func (b *Builder) placeholder() string {
	b.paramCount++
	switch b.dialect {
	case PostgreSQL:
		return fmt.Sprintf("$%d", b.paramCount)
	default:
		return "?"
	}
}

// ToSQL returns the SQL string and arguments
func (b *Builder) ToSQL() (string, []interface{}, error) {
	query, args := b.Build()
	return query, args, nil
}

// MustBuild builds the query and panics on error
func (b *Builder) MustBuild() string {
	query, _ := b.Build()
	return query
}

// WhereExists adds a WHERE EXISTS subquery
func (b *Builder) WhereExists(subquery *Builder) *Builder {
	subQuerySQL, subQueryArgs := subquery.BuildSelect()
	return b.Where("EXISTS ("+subQuerySQL+")", subQueryArgs...)
}

// WhereNotExists adds a WHERE NOT EXISTS subquery
func (b *Builder) WhereNotExists(subquery *Builder) *Builder {
	subQuerySQL, subQueryArgs := subquery.BuildSelect()
	return b.Where("NOT EXISTS ("+subQuerySQL+")", subQueryArgs...)
}

// WithLock adds a locking clause based on the dialect
func (b *Builder) WithLock(lockType string) *Builder {
	switch b.dialect {
	case PostgreSQL:
		b.forUpdate = true
	case MySQL:
		b.forUpdate = true
	case SQLite:
		xlog.Debugf("SQLite does not support explicit locking clauses")
	case MSSQL:
		b.forUpdate = true
	}
	return b
}

// InsertGetId performs an INSERT and returns the last inserted ID
func (b *Builder) InsertGetId(db *sql.DB) (int64, error) {
	query, args := b.BuildInsert()
	// "RETURNING id" is already appended in BuildInsert for PostgreSQL Upsert
	if b.dialect == PostgreSQL && len(b.values) != 1 {
		// Additional logic if needed
	}

	var result sql.Result
	var err error
	var id int64

	if b.tx != nil {
		if b.dialect == PostgreSQL && len(b.values) != 1 {
			err = b.tx.QueryRow(query, args...).Scan(&id)
		} else {
			result, err = b.tx.Exec(query, args...)
		}
	} else {
		if b.dialect == PostgreSQL && len(b.values) != 1 {
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
func (b *Builder) Increment(column string, amount int) *Builder {
	b.updateClauses = append(b.updateClauses, UpdateClause{
		Column: column,
		Value: RawExpr{
			Expr: fmt.Sprintf("%s + %d", column, amount),
		},
	})
	return b
}

// Decrement decrements a column's value
func (b *Builder) Decrement(column string, amount int) *Builder {
	b.updateClauses = append(b.updateClauses, UpdateClause{
		Column: column,
		Value: RawExpr{
			Expr: fmt.Sprintf("%s - %d", column, amount),
		},
	})
	return b
}

// Sanitize removes any potentially harmful SQL from the input
func Sanitize(input string) string {
	// Remove comments
	reComment := regexp.MustCompile(`(?s)/\*.*?\*/|--.*?$`)
	sanitized := reComment.ReplaceAllString(input, "")

	// Remove semicolons
	sanitized = strings.ReplaceAll(sanitized, ";", "")

	// Remove UNION keywords
	reUnion := regexp.MustCompile(`(?i)\bUNION\b`)
	sanitized = reUnion.ReplaceAllString(sanitized, "")

	// Trim spaces
	sanitized = strings.TrimSpace(sanitized)

	return sanitized
}

func (b *Builder) WithContext(ctx context.Context) *Builder {
	b.ctx = ctx
	return b
}

func (b *Builder) Error() error {
	return b.err
}

func (b *Builder) AllowEmptyWhere() *Builder {
	b.allowEmptyWhere = true
	return b
}

func (b *Builder) LogSQL() *Builder {
	b.logSQL = true
	return b
}

func (b *Builder) WhereMap(m map[string]interface{}) *Builder {
	for k, v := range m {
		b.Where(fmt.Sprintf("%s = %s", k, b.placeholder()), v)
	}
	return b
}

func (b *Builder) SetMap(m map[string]interface{}) *Builder {
	for k, v := range m {
		b.Set(k, v)
	}
	return b
}

func (b *Builder) InsertMap(m map[string]interface{}) *Builder {
	for k, v := range m {
		b.columns = append(b.columns, k)
		b.values = append(b.values, v)
	}
	return b
}

func (b *Builder) First(db *sql.DB) (*sql.Row, error) {
	b.Limit(1)
	query, args := b.BuildSelect()
	if b.logSQL {
		xlog.Debugf("[SQL] %s %v", query, args)
	}

	if b.tx != nil {
		return b.tx.QueryRowContext(b.ctx, query, args...), nil
	}
	return db.QueryRowContext(b.ctx, query, args...), nil
}

func (b *Builder) MustExec(db *sql.DB) sql.Result {
	result, err := b.Exec(db)
	if err != nil {
		panic(err)
	}
	return result
}
