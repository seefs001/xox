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
- Advanced features like CTEs, UPSERTs, and recursive queries
- Pagination support
- Increment and decrement operations

## Installation

```bash
go get github.com/seefs001/xox/xsb
```

## Usage

### Basic Query Building

```go
import "github.com/seefs001/xox/xsb"

builder := xsb.New().
    WithDialect(xsb.MySQL).
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
builder := xsb.New().
    WithDialect(xsb.MySQL).
    Table("users").
    Columns("name", "email").
    Values("John Doe", "john@example.com")

query, args := builder.BuildInsert()
// query: INSERT INTO users (name, email) VALUES (?, ?)
// args: ["John Doe", "john@example.com"]
```

### Update Query

```go
builder := xsb.New().
    WithDialect(xsb.MySQL).
    Table("users").
    Set("name", "Jane Doe").
    Where("id = ?", 1)

query, args := builder.BuildUpdate()
// query: UPDATE users SET name = ? WHERE id = ?
// args: ["Jane Doe", 1]
```

### Delete Query

```go
builder := xsb.New().
    WithDialect(xsb.MySQL).
    Table("users").
    Where("id = ?", 1)

query, args := builder.BuildDelete()
// query: DELETE FROM users WHERE id = ?
// args: [1]
```

### Joins

```go
builder := xsb.New().
    WithDialect(xsb.MySQL).
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
subquery := xsb.New().
    WithDialect(xsb.MySQL).
    Table("orders").
    Columns("user_id", "COUNT(*) as order_count").
    GroupBy("user_id")

builder := xsb.New().
    WithDialect(xsb.MySQL).
    Table("users u").
    Columns("u.id", "u.name", "o.order_count").
    LeftJoin(xsb.New().Subquery(subquery, "o"), "u.id = o.user_id")

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

builder := xsb.New().
    WithDialect(xsb.MySQL).
    WithTransaction(tx).
    Table("users").
    Set("status", "active").
    Where("id = ?", 1)

_, err = builder.Exec()
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
builder := xsb.New().
    WithDialect(xsb.MySQL).
    Table("users").
    Columns("id", "name").
    Where("id = ?", 1).
    Explain()

query, args := builder.Build()
// query: EXPLAIN SELECT id, name FROM users WHERE id = ?
// args: [1]
```

## Advanced Features

### Upsert (PostgreSQL)

```go
builder := xsb.New().
    WithDialect(xsb.PostgreSQL).
    Table("users").
    Columns("id", "name", "email").
    Values(1, "John Doe", "john@example.com").
    Upsert([]string{"id"}, []xsb.UpdateClause{
        {Column: "name", Value: "John Doe Updated"},
        {Column: "email", Value: "john_updated@example.com"},
    })

query, args := builder.BuildInsert()
// query: INSERT INTO users (id, name, email) VALUES ($1, $2, $3) ON CONFLICT (id) DO UPDATE SET name = $4, email = $5 RETURNING id
// args: [1, "John Doe", "john@example.com", "John Doe Updated", "john_updated@example.com"]
```

### On Duplicate Key Update (MySQL)

```go
builder := xsb.New().
    WithDialect(xsb.MySQL).
    Table("users").
    Columns("id", "name", "email").
    Values(1, "John Doe", "john@example.com").
    OnDuplicateKeyUpdate([]xsb.UpdateClause{
        {Column: "name", Value: "John Doe Updated"},
        {Column: "email", Value: "john_updated@example.com"},
    })

query, args := builder.BuildInsert()
// query: INSERT INTO users (id, name, email) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE name = ?, email = ?
// args: [1, "John Doe", "john@example.com", "John Doe Updated", "john_updated@example.com"]
```

### Recursive CTE

```go
cte := xsb.New().
    WithDialect(xsb.PostgreSQL).
    Table("employees").
    Columns("id", "name", "manager_id").
    Union(xsb.New().
        WithDialect(xsb.PostgreSQL).
        Table("employees e").
        InnerJoin("cte c", "e.manager_id = c.id").
        Columns("e.id", "e.name", "e.manager_id"))

builder := xsb.New().
    WithDialect(xsb.PostgreSQL).
    WithRecursive("cte", cte).
    Table("cte").
    Columns("id", "name", "manager_id").
    Where("id = ?", 1)

query, args := builder.Build()
// query: WITH RECURSIVE cte AS (SELECT id, name, manager_id FROM employees UNION SELECT e.id, e.name, e.manager_id FROM employees e INNER JOIN cte c ON e.manager_id = c.id) SELECT id, name, manager_id FROM cte WHERE id = ?
// args: [1]
```

### Increment and Decrement

```go
builder := xsb.New().
    WithDialect(xsb.MySQL).
    Table("products").
    Increment("stock", 5).
    Decrement("price", 2).
    Where("id = ?", 1)

query, args := builder.BuildUpdate()
// query: UPDATE products SET stock = stock + 5, price = price - 2 WHERE id = ?
// args: [1]
```

### Pagination

```go
builder := xsb.New().
    WithDialect(xsb.MySQL).
    Table("users").
    Columns("id", "name").
    OrderBy("id ASC").
    Paginate(2, 10)

query, args := builder.BuildSelect()
// query: SELECT id, name FROM users ORDER BY id ASC LIMIT 10 OFFSET 10
// args: []
```

## API Reference

### New() *Builder
Creates a new Builder instance with MySQL as the default dialect.

### WithDialect(dialect Dialect) *Builder
Sets a specific dialect for the builder.

### Table(name string) *Builder
Sets the table name for the query.

### Columns(cols ...string) *Builder
Sets the columns for the query.

### Values(values ...interface{}) *Builder
Sets the values for an INSERT query.

### Where(condition string, args ...interface{}) *Builder
Adds a WHERE clause to the query.

### OrWhere(condition string, args ...interface{}) *Builder
Adds an OR condition to the WHERE clause.

### WhereIn(column string, values ...interface{}) *Builder
Adds a WHERE IN clause.

### OrderBy(clause string) *Builder
Adds an ORDER BY clause to the query.

### Limit(limit int) *Builder
Adds a LIMIT clause to the query.

### Offset(offset int) *Builder
Adds an OFFSET clause to the query.

### Join(joinType, table, condition string) *Builder
Adds a JOIN clause to the query.

### InnerJoin(table, condition string) *Builder
Adds an INNER JOIN clause.

### LeftJoin(table, condition string) *Builder
Adds a LEFT JOIN clause.

### RightJoin(table, condition string) *Builder
Adds a RIGHT JOIN clause.

### GroupBy(clause string) *Builder
Adds a GROUP BY clause to the query.

### Having(condition string, args ...interface{}) *Builder
Adds a HAVING clause to the query.

### Set(column string, value interface{}) *Builder
Adds a column-value pair for UPDATE queries.

### Union(other *Builder) *Builder
Adds a UNION clause.

### CTE(name string, subquery *Builder) *Builder
Adds a Common Table Expression (WITH clause).

### Lock() *Builder
Adds FOR UPDATE to the SELECT query (for supported dialects).

### Build() (string, []interface{})
Builds the final query string and returns it along with the arguments.

### BuildInsert() (string, []interface{})
Builds an INSERT query.

### BuildUpdate() (string, []interface{})
Builds an UPDATE query.

### BuildDelete() (string, []interface{})
Builds a DELETE query.

### Exec() (sql.Result, error)
Executes the query and returns the result.

### QueryRow() *sql.Row
Executes the query and returns a single row.

### Query() (*sql.Rows, error)
Executes the query and returns multiple rows.

### WithTransaction(tx *sql.Tx) *Builder
Wraps the builder with a transaction.

### Explain() *Builder
Adds EXPLAIN to the query for debugging purposes.

### Sanitize(input string) string
Removes any potentially harmful SQL from the input.

### Increment(column string, amount int) *Builder
Increments a column's value.

### Decrement(column string, amount int) *Builder
Decrements a column's value.

### Paginate(page, perPage int) *Builder
Adds LIMIT and OFFSET for pagination.

### WhereExists(subquery *Builder) *Builder
Adds a WHERE EXISTS subquery.

### WhereNotExists(subquery *Builder) *Builder
Adds a WHERE NOT EXISTS subquery.

### WithLock(lockType string) *Builder
Adds a locking clause based on the dialect.

### InsertGetId(db *sql.DB) (int64, error)
Performs an INSERT and returns the last inserted ID.

## Best Practices

1. Always use placeholders for values in WHERE clauses to prevent SQL injection.
2. Use transactions for operations that require multiple queries to be executed atomically.
3. Use the Explain() method to debug and optimize your queries.
4. Sanitize user inputs using the Sanitize() function before using them in your queries.
5. Use the appropriate dialect for your database to ensure compatibility.
6. Take advantage of subqueries and CTEs for complex queries.
7. Use pagination for large result sets to improve performance.
8. Use the WithLock() method when you need to ensure data consistency in concurrent scenarios.

## Contributing

Contributions to XSB are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.