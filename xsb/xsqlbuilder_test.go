package xsb_test

import (
	"testing"
	"time"

	"github.com/seefs001/xox/xsb"
	"github.com/stretchr/testify/assert"
)

func TestSimpleQueries(t *testing.T) {
	t.Run("SimpleSelect", func(t *testing.T) {
		builder := xsb.New[any]().
			Table("users").
			Columns("id", "name", "email").
			Where("age > ?", 18)

		query, args := builder.Build()
		assert.Equal(t, "SELECT id, name, email FROM users WHERE age > ?", query)
		assert.Equal(t, []interface{}{18}, args)
	})

	t.Run("SimpleInsert", func(t *testing.T) {
		builder := xsb.New[any]().
			Table("users").
			Columns("name", "email").
			Values("John Doe", "john@example.com")

		query, args := builder.BuildInsert()
		assert.Equal(t, "INSERT INTO users (name, email) VALUES (?, ?)", query)
		assert.Equal(t, []interface{}{"John Doe", "john@example.com"}, args)
	})

	t.Run("SimpleUpdate", func(t *testing.T) {
		builder := xsb.New[any]().
			Table("users").
			Set("name", "Jane Doe").
			Where("id = ?", 1)

		query, args := builder.BuildUpdate()
		assert.Equal(t, "UPDATE users SET name = ? WHERE id = ?", query)
		assert.Equal(t, []interface{}{"Jane Doe", 1}, args)
	})

	t.Run("SimpleDelete", func(t *testing.T) {
		builder := xsb.New[any]().
			Table("users").
			Where("id = ?", 1)

		query, args := builder.BuildDelete()
		assert.Equal(t, "DELETE FROM users WHERE id = ?", query)
		assert.Equal(t, []interface{}{1}, args)
	})
}

func TestComplexQueries(t *testing.T) {
	t.Run("JoinWithSubquery", func(t *testing.T) {
		subquery := xsb.New[any]().
			Table("orders").
			Columns("user_id", "COUNT(*) as order_count").
			GroupBy("user_id")

		builder := xsb.New[any]().
			Table("users u").
			Columns("u.id", "u.name", "o.order_count").
			LeftJoin(subquery.BuildSQL(), "o ON u.id = o.user_id").
			Where("u.active = ?", true).
			OrderBy("o.order_count DESC").
			Limit(10)

		query, args := builder.Build()
		expectedQuery := "SELECT u.id, u.name, o.order_count FROM users u LEFT JOIN (SELECT user_id, COUNT(*) as order_count FROM orders GROUP BY user_id) AS o ON u.id = o.user_id WHERE u.active = ? ORDER BY o.order_count DESC LIMIT 10"
		assert.Equal(t, expectedQuery, query)
		assert.Equal(t, []interface{}{true}, args)
	})

	t.Run("ComplexInsertWithCTE", func(t *testing.T) {
		cte := xsb.New[any]().
			Table("recent_orders").
			Columns("user_id", "total_amount").
			Where("order_date > ?", "2023-01-01")

		builder := xsb.New[any]().
			CTE("recent_orders_cte", cte).
			Table("user_stats").
			Columns("user_id", "order_count", "total_spent")

		subquery := xsb.New[any]().
			Table("recent_orders_cte").
			Columns("user_id", "COUNT(*)", "SUM(total_amount)").
			GroupBy("user_id")

		builder.Values(subquery.BuildSQL())

		query, args := builder.BuildInsert()
		expectedQuery := "WITH recent_orders_cte AS (SELECT user_id, total_amount FROM recent_orders WHERE order_date > ?) INSERT INTO user_stats (user_id, order_count, total_spent) SELECT user_id, COUNT(*), SUM(total_amount) FROM recent_orders_cte GROUP BY user_id"
		assert.Equal(t, expectedQuery, query)
		assert.Equal(t, []interface{}{"2023-01-01"}, args)
	})
}

func TestAdvancedFeatures(t *testing.T) {
	t.Run("Upsert", func(t *testing.T) {
		builder := xsb.New[any]().
			Table("users").
			Columns("id", "name", "email").
			Values(1, "John Doe", "john@example.com").
			Upsert([]string{"id"}, map[string]interface{}{"name": "John Doe Updated", "email": "john_updated@example.com"})

		query, args := builder.BuildInsert()
		expectedQuery := "INSERT INTO users (id, name, email) VALUES (?, ?, ?) ON CONFLICT (id) DO UPDATE SET name = ?, email = ?"
		assert.Equal(t, expectedQuery, query)
		assert.Equal(t, []interface{}{1, "John Doe", "john@example.com", "John Doe Updated", "john_updated@example.com"}, args)
	})

	t.Run("WithRecursive", func(t *testing.T) {
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
		expectedQuery := "WITH RECURSIVE cte AS (SELECT id, name, manager_id FROM employees UNION ALL SELECT e.id, e.name, e.manager_id FROM employees e INNER JOIN cte c ON e.manager_id = c.id) SELECT id, name, manager_id FROM cte WHERE id = ?"
		assert.Equal(t, expectedQuery, query)
		assert.Equal(t, []interface{}{1}, args)
	})

	t.Run("Lock", func(t *testing.T) {
		builder := xsb.New[any]().
			Table("users").
			Columns("id", "name").
			Where("id = ?", 1).
			Lock()

		query, args := builder.Build()
		expectedQuery := "SELECT id, name FROM users WHERE id = ? FOR UPDATE"
		assert.Equal(t, expectedQuery, query)
		assert.Equal(t, []interface{}{1}, args)
	})

	t.Run("Explain", func(t *testing.T) {
		builder := xsb.New[any]().
			Table("users").
			Columns("id", "name").
			Where("id = ?", 1).
			Explain()

		query, args := builder.Build()
		expectedQuery := "EXPLAIN SELECT id, name FROM users WHERE id = ?"
		assert.Equal(t, expectedQuery, query)
		assert.Equal(t, []interface{}{1}, args)
	})
}

func TestDialectSpecificFeatures(t *testing.T) {
	t.Run("MySQLOnDuplicateKeyUpdate", func(t *testing.T) {
		builder := xsb.New[any]().
			WithDialect(xsb.MySQL).
			Table("users").
			Columns("id", "name", "email").
			Values(1, "John Doe", "john@example.com").
			OnDuplicateKeyUpdate(map[string]interface{}{
				"name":  "John Doe Updated",
				"email": "john_updated@example.com",
			})

		query, args := builder.BuildInsert()
		expectedQuery := "INSERT INTO users (id, name, email) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE name = ?, email = ?"
		assert.Equal(t, expectedQuery, query)
		assert.Equal(t, []interface{}{1, "John Doe", "john@example.com", "John Doe Updated", "john_updated@example.com"}, args)
	})

	t.Run("SQLiteWithoutLock", func(t *testing.T) {
		builder := xsb.New[any]().
			WithDialect(xsb.SQLite).
			Table("users").
			Columns("id", "name").
			Where("id = ?", 1).
			Lock()

		query, args := builder.Build()
		expectedQuery := "SELECT id, name FROM users WHERE id = ?"
		assert.Equal(t, expectedQuery, query)
		assert.Equal(t, []interface{}{1}, args)
	})

	t.Run("MSSQLWithLock", func(t *testing.T) {
		builder := xsb.New[any]().
			WithDialect(xsb.MSSQL).
			Table("users").
			Columns("id", "name").
			Where("id = ?", 1).
			Lock()

		query, args := builder.Build()
		expectedQuery := "SELECT id, name FROM users WITH (UPDLOCK, ROWLOCK) WHERE id = ?"
		assert.Equal(t, expectedQuery, query)
		assert.Equal(t, []interface{}{1}, args)
	})
}

func TestUtilityMethods(t *testing.T) {
	t.Run("FromStruct", func(t *testing.T) {
		type User struct {
			ID    int    `db:"id"`
			Name  string `db:"name"`
			Email string `db:"email"`
		}

		user := User{ID: 1, Name: "John Doe", Email: "john@example.com"}
		builder := xsb.New[User]().
			Table("users").
			FromStruct(user)

		query, args := builder.BuildInsert()
		expectedQuery := "INSERT INTO users (id, name, email) VALUES (?, ?, ?)"
		assert.Equal(t, expectedQuery, query)
		assert.Equal(t, []interface{}{1, "John Doe", "john@example.com"}, args)
	})

	t.Run("Paginate", func(t *testing.T) {
		builder := xsb.New[any]().
			Table("users").
			Columns("id", "name").
			OrderBy("id ASC").
			Paginate(2, 10)

		query, args := builder.Build()
		expectedQuery := "SELECT id, name FROM users ORDER BY id ASC LIMIT 10 OFFSET 10"
		assert.Equal(t, expectedQuery, query)
		assert.Empty(t, args)
	})

	t.Run("WhereExists", func(t *testing.T) {
		subquery := xsb.New[any]().
			Table("orders").
			Columns("1").
			Where("orders.user_id = users.id")

		builder := xsb.New[any]().
			Table("users").
			Columns("id", "name").
			WhereExists(subquery)

		query, args := builder.Build()
		expectedQuery := "SELECT id, name FROM users WHERE EXISTS (SELECT 1 FROM orders WHERE orders.user_id = users.id)"
		assert.Equal(t, expectedQuery, query)
		assert.Empty(t, args)
	})

	t.Run("Sanitize", func(t *testing.T) {
		input := "SELECT * FROM users; DROP TABLE users; --"
		sanitized := xsb.Sanitize(input)
		expected := "SELECT * FROM users"
		assert.Equal(t, expected, sanitized)
	})
}

func TestTransactionAndExecution(t *testing.T) {
	// Note: These tests would typically use a mock database or a test database
	// For brevity, we'll just test the query building part

	t.Run("WithTransaction", func(t *testing.T) {
		builder := xsb.New[any]().
			Table("users").
			Columns("id", "name").
			Where("id = ?", 1)

		// In a real scenario, you'd start a transaction here
		// tx, _ := db.Begin()
		// builder = builder.WithTransaction(tx)

		query, args := builder.Build()
		expectedQuery := "SELECT id, name FROM users WHERE id = ?"
		assert.Equal(t, expectedQuery, query)
		assert.Equal(t, []interface{}{1}, args)

		// In a real scenario, you'd commit or rollback the transaction here
		// tx.Commit() or tx.Rollback()
	})

	t.Run("Configure", func(t *testing.T) {
		builder := xsb.New[any]().
			Configure(xsb.Config{
				MaxOpenConns:    10,
				MaxIdleConns:    5,
				ConnMaxLifetime: time.Hour,
			})

		// The Configure method doesn't affect the query building
		// It's used to set up the database connection pool
		// For testing purposes, we'll just check that the builder is still usable

		builder = builder.
			Table("users").
			Columns("id", "name").
			Where("id = ?", 1)

		query, args := builder.Build()
		expectedQuery := "SELECT id, name FROM users WHERE id = ?"
		assert.Equal(t, expectedQuery, query)
		assert.Equal(t, []interface{}{1}, args)
	})
}

func TestErrorHandling(t *testing.T) {
	t.Run("BuildWithoutTable", func(t *testing.T) {
		builder := xsb.New[any]().
			Columns("id", "name")

		query, args := builder.Build()
		assert.Empty(t, query)
		assert.Empty(t, args)
	})

	t.Run("UnsupportedDialectFeature", func(t *testing.T) {
		builder := xsb.New[any]().
			WithDialect(xsb.SQLite).
			Table("users").
			Columns("id", "name").
			Upsert([]string{"id"}, map[string]interface{}{"name": "John Doe"})

		query, args := builder.BuildInsert()
		// SQLite doesn't support UPSERT, so it should fall back to a regular INSERT
		expectedQuery := "INSERT INTO users (id, name) VALUES (?, ?)"
		assert.Equal(t, expectedQuery, query)
		assert.Empty(t, args)
	})
}
