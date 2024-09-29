# XSB - Extensible SQL Builder

XSB is a powerful and flexible SQL query builder for Go, designed to simplify the process of constructing complex SQL queries programmatically. It supports multiple dialects and provides a fluent interface for building SELECT, INSERT, UPDATE, and DELETE queries.

## Features

- Support for multiple SQL dialects (PostgreSQL, MySQL, SQLite, MSSQL)
- Fluent interface for building queries
- Type-safe query construction
- Support for subqueries, joins, and complex conditions
- Transaction support
- Query debugging and explanation
- Sanitization of user inputs

## Installation

```bash
go get github.com/seefs001/xox/xsb
```

## Usage

### Basic Query Building

```go
import "github.com/seefs001/xox/xsb"

builder := xsb.New[any]().
    Table("users").
    Columns("id", "name", "email").
    Where("age > ?", 18).
    OrderBy("name ASC").
    Limit(10)

query, args := builder.Build()
// query: SELECT id, name, email FROM users WHERE age > ? ORDER BY name ASC LIMIT 10
// args: [18]
```

### Insert Query

```go
builder := xsb.New[any]().
    Table("users").
    Columns("name", "email").
    Values("John Doe", "john@example.com")

query, args := builder.BuildInsert()
// query: INSERT INTO users (name, email) VALUES (?, ?)
// args: ["John Doe", "john@example.com"]
```

### Update Query

```go
builder := xsb.New[any]().
    Table("users").
    Set("name", "Jane Doe").
    Where("id = ?", 1)

query, args := builder.BuildUpdate()
// query: UPDATE users SET name = ? WHERE id = ?
// args: ["Jane Doe", 1]
```

### Delete Query

```go
builder := xsb.New[any]().
    Table("users").
    Where("id = ?", 1)

query, args := builder.BuildDelete()
// query: DELETE FROM users WHERE id = ?
// args: [1]
```

### Joins

```go
builder := xsb.New[any]().
    Table("users u").
    Columns("u.id", "u.name", "o.order_id").
    InnerJoin("orders o", "u.id = o.user_id").
    Where("u.active = ?", true)

query, args := builder.Build()
// query: SELECT u.id, u.name, o.order_id FROM users u INNER JOIN orders o ON u.id = o.user_id WHERE u.active = ?
// args: [true]
```

### Subqueries

```go
subquery := xsb.New[any]().
    Table("orders").
    Columns("user_id", "COUNT(*) as order_count").
    GroupBy("user_id")

builder := xsb.New[any]().
    Table("users u").
    Columns("u.id", "u.name", "o.order_count").
    LeftJoin(subquery.BuildSQL(), "o ON u.id = o.user_id")

query, args := builder.Build()
// query: SELECT u.id, u.name, o.order_count FROM users u LEFT JOIN (SELECT user_id, COUNT(*) as order_count FROM orders GROUP BY user_id) AS o ON u.id = o.user_id
// args: []
```

### Transaction Support

```go
tx, err := db.Begin()
if err != nil {
    // Handle error
}

builder := xsb.New[any]().
    WithTransaction(tx).
    Table("users").
    Set("status", "active").
    Where("id = ?", 1)

_, err = builder.Exec(db)
if err != nil {
    tx.Rollback()
    // Handle error
}

err = tx.Commit()
if err != nil {
    // Handle error
}
```

### Debugging

```go
builder := xsb.New[any]().
    EnableDebug().
    Table("users").
    Columns("id", "name").
    Where("id = ?", 1)

debugInfo := builder.Debug()
fmt.Println(debugInfo)
// Output: Query Type: SELECT
//         Query: SELECT id, name FROM users WHERE id = ?
//         Args: [1]
```

## API Reference

### New[T any]() *Builder[T]
Creates a new Builder instance with PostgreSQL as the default dialect.

### WithDialect(dialect Dialect) *Builder[T]
Sets a specific dialect for the builder.

### Table(name string) *Builder[T]
Sets the table name for the query.

### Columns(cols ...string) *Builder[T]
Sets the columns for the query.

### Values(values ...interface{}) *Builder[T]
Sets the values for an INSERT query.

### Where(condition string, args ...interface{}) *Builder[T]
Adds a WHERE clause to the query.

### OrWhere(condition string, args ...interface{}) *Builder[T]
Adds an OR condition to the WHERE clause.

### WhereIn(column string, values ...interface{}) *Builder[T]
Adds a WHERE IN clause.

### OrderBy(clause string) *Builder[T]
Adds an ORDER BY clause to the query.

### Limit(limit int) *Builder[T]
Adds a LIMIT clause to the query.

### Offset(offset int) *Builder[T]
Adds an OFFSET clause to the query.

### Join(joinType, table, condition string) *Builder[T]
Adds a JOIN clause to the query.

### InnerJoin(table, condition string) *Builder[T]
Adds an INNER JOIN clause.

### LeftJoin(table, condition string) *Builder[T]
Adds a LEFT JOIN clause.

### RightJoin(table, condition string) *Builder[T]
Adds a RIGHT JOIN clause.

### GroupBy(clause string) *Builder[T]
Adds a GROUP BY clause to the query.

### Having(condition string, args ...interface{}) *Builder[T]
Adds a HAVING clause to the query.

### Set(column string, value interface{}) *Builder[T]
Adds a column-value pair for UPDATE queries.

### Returning(columns ...string) *Builder[T]
Adds a RETURNING clause (for PostgreSQL).

### Union(other *Builder[T]) *Builder[T]
Adds a UNION clause.

### CTE(name string, subquery *Builder[T]) *Builder[T]
Adds a Common Table Expression (WITH clause).

### Distinct() *Builder[T]
Adds DISTINCT to the SELECT query.

### Lock() *Builder[T]
Adds FOR UPDATE to the SELECT query.

### Build() (string, []interface{})
Builds the final query string and returns it along with the arguments.

### BuildInsert() (string, []interface{})
Builds an INSERT query.

### BuildUpdate() (string, []interface{})
Builds an UPDATE query.

### BuildDelete() (string, []interface{})
Builds a DELETE query.

### Exec(db *sql.DB) (sql.Result, error)
Executes the query and returns the result.

### QueryRow(db *sql.DB) *sql.Row
Executes the query and returns a single row.

### Query(db *sql.DB) (*sql.Rows, error)
Executes the query and returns multiple rows.

### WithTransaction(tx *sql.Tx) *Builder[T]
Wraps the builder with a transaction.

### EnableDebug() *Builder[T]
Turns on debug mode.

### Debug() string
Returns a string representation of the query for debugging purposes.

### Sanitize(input string) string
Removes any potentially harmful SQL from the input.

## Advanced Features

### Upsert (PostgreSQL)

```go
builder := xsb.New[any]().
    WithDialect(xsb.PostgreSQL).
    Table("users").
    Columns("id", "name", "email").
    Values(1, "John Doe", "john@example.com").
    Upsert([]string{"id"}, map[string]interface{}{
        "name": "John Doe Updated",
        "email": "john_updated@example.com",
    })

query, args := builder.BuildInsert()
// query: INSERT INTO users (id, name, email) VALUES (?, ?, ?) ON CONFLICT (id) DO UPDATE SET name = ?, email = ?
// args: [1, "John Doe", "john@example.com", "John Doe Updated", "john_updated@example.com"]
```

### On Duplicate Key Update (MySQL)

```go
builder := xsb.New[any]().
    WithDialect(xsb.MySQL).
    Table("users").
    Columns("id", "name", "email").
    Values(1, "John Doe", "john@example.com").
    OnDuplicateKeyUpdate(map[string]interface{}{
        "name": "John Doe Updated",
        "email": "john_updated@example.com",
    })

query, args := builder.BuildInsert()
// query: INSERT INTO users (id, name, email) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE name = ?, email = ?
// args: [1, "John Doe", "john@example.com", "John Doe Updated", "john_updated@example.com"]
```

### Recursive CTE

```go
cte := xsb.New[any]().
    Table("employees").
    Columns("id", "name", "manager_id").
    UnionAll(xsb.New[any]().
        Table("employees e").
        InnerJoin("cte c", "e.manager_id = c.id").
        Columns("e.id", "e.name", "e.manager_id"))

builder := xsb.New[any]().
    WithRecursive("cte", cte).
    Table("cte").
    Columns("id", "name", "manager_id").
    Where("id = ?", 1)

query, args := builder.Build()
// query: WITH RECURSIVE cte AS (SELECT id, name, manager_id FROM employees UNION ALL SELECT e.id, e.name, e.manager_id FROM employees e INNER JOIN cte c ON e.manager_id = c.id) SELECT id, name, manager_id FROM cte WHERE id = ?
// args: [1]
```
