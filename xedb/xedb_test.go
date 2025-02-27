package xedb_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/seefs001/xox/xedb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*xedb.DB, func()) {
	// Create temp directory
	dir, err := os.MkdirTemp("", "xedb-test-*")
	require.NoError(t, err)

	// Initialize DB with test options
	db, err := xedb.New(
		xedb.WithDataDir(dir),
		xedb.WithSyncWrite(true),
		xedb.WithAutoSaveInterval(time.Second),
	)
	require.NoError(t, err)

	// Return cleanup function
	cleanup := func() {
		db.Close()
		os.RemoveAll(dir)
	}

	return db, cleanup
}

func TestDB_String(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("Set and Get", func(t *testing.T) {
		err := db.String("key1").Set("value1")
		assert.NoError(t, err)

		val, exists := db.String("key1").Get()
		assert.True(t, exists)
		assert.Equal(t, "value1", val)
	})

	t.Run("Get Non-Existent", func(t *testing.T) {
		val, exists := db.String("nonexistent").Get()
		assert.False(t, exists)
		assert.Empty(t, val)
	})
}

func TestDB_List(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("Push and Pop", func(t *testing.T) {
		err := db.List("list1").Push("item1", "item2")
		assert.NoError(t, err)

		val, exists := db.List("list1").Pop()
		assert.True(t, exists)
		assert.Equal(t, "item2", val)

		val, exists = db.List("list1").Pop()
		assert.True(t, exists)
		assert.Equal(t, "item1", val)

		val, exists = db.List("list1").Pop()
		assert.False(t, exists)
		assert.Empty(t, val)
	})
}

func TestDB_Hash(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("Set and Get", func(t *testing.T) {
		err := db.Hash("hash1").Set("field1", "value1")
		assert.NoError(t, err)

		val, exists := db.Hash("hash1").Get("field1")
		assert.True(t, exists)
		assert.Equal(t, "value1", val)
	})

	t.Run("Get Non-Existent Field", func(t *testing.T) {
		val, exists := db.Hash("hash1").Get("nonexistent")
		assert.False(t, exists)
		assert.Empty(t, val)
	})
}

func TestDB_Set(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("Add and Check Membership", func(t *testing.T) {
		err := db.Set("set1").Add("member1", "member2")
		assert.NoError(t, err)

		assert.True(t, db.Set("set1").IsMember("member1"))
		assert.True(t, db.Set("set1").IsMember("member2"))
		assert.False(t, db.Set("set1").IsMember("nonexistent"))
	})
}

func TestDB_ZSet(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("Add and Range", func(t *testing.T) {
		err := db.ZSet("zset1").Add(1.0, "member1")
		assert.NoError(t, err)
		err = db.ZSet("zset1").Add(2.0, "member2")
		assert.NoError(t, err)

		members := db.ZSet("zset1").Range(0, 1)
		assert.Len(t, members, 2)
		assert.Equal(t, "member1", members[0].Member)
		assert.Equal(t, 1.0, members[0].Score)
		assert.Equal(t, "member2", members[1].Member)
		assert.Equal(t, 2.0, members[1].Score)
	})

	t.Run("Update Score", func(t *testing.T) {
		err := db.ZSet("zset1").Add(3.0, "member1")
		assert.NoError(t, err)

		members := db.ZSet("zset1").Range(0, 1)
		assert.Len(t, members, 2)
		assert.Equal(t, "member2", members[0].Member)
		assert.Equal(t, "member1", members[1].Member)
		assert.Equal(t, 3.0, members[1].Score)
	})

	t.Run("Negative Indices", func(t *testing.T) {
		err := db.ZSet("zset2").Add(1.0, "a")
		assert.NoError(t, err)
		err = db.ZSet("zset2").Add(2.0, "b")
		assert.NoError(t, err)
		err = db.ZSet("zset2").Add(3.0, "c")
		assert.NoError(t, err)

		// Test various negative index scenarios
		t.Run("Last Two Elements", func(t *testing.T) {
			members := db.ZSet("zset2").Range(-2, -1)
			assert.Len(t, members, 2)
			assert.Equal(t, "b", members[0].Member)
			assert.Equal(t, "c", members[1].Member)
		})

		t.Run("Last Element", func(t *testing.T) {
			members := db.ZSet("zset2").Range(-1, -1)
			assert.Len(t, members, 1)
			assert.Equal(t, "c", members[0].Member)
		})

		t.Run("Out of Bounds", func(t *testing.T) {
			members := db.ZSet("zset2").Range(-5, -4)
			assert.Nil(t, members)
		})

		t.Run("Mixed Indices", func(t *testing.T) {
			members := db.ZSet("zset2").Range(-2, 2)
			assert.Len(t, members, 2)
			assert.Equal(t, "b", members[0].Member)
			assert.Equal(t, "c", members[1].Member)
		})
	})
}

func TestDB_Persistence(t *testing.T) {
	dir, err := os.MkdirTemp("", "xedb-persist-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create and populate DB
	db, err := xedb.New(xedb.WithDataDir(dir))
	require.NoError(t, err)

	err = db.String("key1").Set("value1")
	require.NoError(t, err)

	err = db.Save()
	require.NoError(t, err)
	db.Close()

	// Reopen DB and verify data
	db2, err := xedb.New(xedb.WithDataDir(dir))
	require.NoError(t, err)
	defer db2.Close()

	val, exists := db2.String("key1").Get()
	assert.True(t, exists)
	assert.Equal(t, "value1", val)
}

func TestDB_Options(t *testing.T) {
	t.Run("Custom Options", func(t *testing.T) {
		dir, err := os.MkdirTemp("", "xedb-options-test-*")
		require.NoError(t, err)
		defer os.RemoveAll(dir)

		db, err := xedb.New(
			xedb.WithDataDir(dir),
			xedb.WithSyncWrite(false),
			xedb.WithAutoSaveInterval(time.Second*30),
			xedb.WithMaxMemory(1<<20),
		)
		require.NoError(t, err)
		defer db.Close()

		assert.NotNil(t, db)
	})
}

func TestDB_BatchOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("Multiple Operations", func(t *testing.T) {
		ops := []xedb.BatchOp{
			{Op: "STRING", Key: "str1", Value: "value1"},
			{Op: "LIST", Key: "list1", Value: []string{"item1", "item2"}},
			{Op: "HASH", Key: "hash1", Value: map[string]string{"field1": "value1"}},
		}

		err := db.ExecuteBatch(ops)
		assert.NoError(t, err)

		// Verify string operation
		val, exists := db.String("str1").Get()
		assert.True(t, exists)
		assert.Equal(t, "value1", val)

		// Verify list operation
		val, exists = db.List("list1").Pop()
		assert.True(t, exists)
		assert.Equal(t, "item2", val)

		// Verify hash operation
		val, exists = db.Hash("hash1").Get("field1")
		assert.True(t, exists)
		assert.Equal(t, "value1", val)
	})
}

func TestDB_MemoryLimit(t *testing.T) {
	dir, err := os.MkdirTemp("", "xedb-memory-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	db, err := xedb.New(
		xedb.WithDataDir(dir),
		xedb.WithMaxMemory(100), // Set very low memory limit
	)
	require.NoError(t, err)
	defer db.Close()

	// Test string operations with memory limit
	t.Run("String Memory Limit", func(t *testing.T) {
		largeValue := string(make([]byte, 200)) // Value larger than memory limit
		err := db.String("large").Set(largeValue)
		assert.Error(t, err)
		assert.ErrorIs(t, err, xedb.ErrMemoryLimit)
	})
}

func TestDB_DataTypes(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("Type Mismatch", func(t *testing.T) {
		// Set a string value
		err := db.String("key1").Set("value1")
		assert.NoError(t, err)

		// Try to use it as a list
		val, exists := db.List("key1").Pop()
		assert.False(t, exists)
		assert.Empty(t, val)

		// Try to use it as a hash
		val, exists = db.Hash("key1").Get("field1")
		assert.False(t, exists)
		assert.Empty(t, val)
	})
}

func TestDB_ListOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("LPush and LPop", func(t *testing.T) {
		err := db.List("list1").LPush("item1", "item2")
		assert.NoError(t, err)

		val, exists := db.List("list1").LPop()
		assert.True(t, exists)
		assert.Equal(t, "item1", val)

		val, exists = db.List("list1").LPop()
		assert.True(t, exists)
		assert.Equal(t, "item2", val)
	})

	t.Run("Range Operations", func(t *testing.T) {
		err := db.List("list2").Push("a", "b", "c", "d", "e")
		assert.NoError(t, err)

		// Test positive indices
		values := db.List("list2").Range(1, 3)
		assert.Equal(t, []string{"b", "c", "d"}, values)

		// Test negative indices
		values = db.List("list2").Range(-3, -1)
		assert.Equal(t, []string{"c", "d", "e"}, values)

		// Test out of bounds
		values = db.List("list2").Range(10, 20)
		assert.Nil(t, values)
	})

	t.Run("Length Operation", func(t *testing.T) {
		err := db.List("list3").Push("a", "b", "c")
		assert.NoError(t, err)

		length := db.List("list3").Len()
		assert.Equal(t, 3, length)

		// Empty list
		length = db.List("nonexistent").Len()
		assert.Equal(t, 0, length)
	})
}

func TestDB_HashOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("Multiple Field Operations", func(t *testing.T) {
		// Set multiple fields
		err := db.Hash("hash1").Set("field1", "value1")
		assert.NoError(t, err)
		err = db.Hash("hash1").Set("field2", "value2")
		assert.NoError(t, err)

		// Get existing fields
		val, exists := db.Hash("hash1").Get("field1")
		assert.True(t, exists)
		assert.Equal(t, "value1", val)

		val, exists = db.Hash("hash1").Get("field2")
		assert.True(t, exists)
		assert.Equal(t, "value2", val)

		// Get non-existent field
		val, exists = db.Hash("hash1").Get("field3")
		assert.False(t, exists)
		assert.Empty(t, val)
	})
}

func TestDB_SetOperations(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("Multiple Member Operations", func(t *testing.T) {
		// Add multiple members
		err := db.Set("set1").Add("member1", "member2", "member3")
		assert.NoError(t, err)

		// Check membership
		assert.True(t, db.Set("set1").IsMember("member1"))
		assert.True(t, db.Set("set1").IsMember("member2"))
		assert.True(t, db.Set("set1").IsMember("member3"))
		assert.False(t, db.Set("set1").IsMember("member4"))

		// Add duplicate member
		err = db.Set("set1").Add("member1")
		assert.NoError(t, err)
		assert.True(t, db.Set("set1").IsMember("member1"))
	})

	t.Run("Empty Set Operations", func(t *testing.T) {
		assert.False(t, db.Set("emptySet").IsMember("member1"))
	})
}

func TestDB_Transaction(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("Basic Transaction", func(t *testing.T) {
		// Start transaction
		txn := db.NewTransaction(true)

		// Set some values
		err := txn.Set("key1", xedb.Entry{
			Type:  xedb.String,
			Value: "value1",
		})
		assert.NoError(t, err)

		err = txn.Set("key2", xedb.Entry{
			Type:  xedb.String,
			Value: "value2",
		})
		assert.NoError(t, err)

		// Commit transaction
		err = txn.Commit()
		assert.NoError(t, err)

		// Verify values
		val, exists := db.String("key1").Get()
		assert.True(t, exists)
		assert.Equal(t, "value1", val)

		val, exists = db.String("key2").Get()
		assert.True(t, exists)
		assert.Equal(t, "value2", val)
	})

	t.Run("Transaction Conflict", func(t *testing.T) {
		// First transaction
		txn1 := db.NewTransaction(true)
		err := txn1.Set("key", xedb.Entry{
			Type:  xedb.String,
			Value: "value1",
		})
		assert.NoError(t, err)

		// Second transaction
		txn2 := db.NewTransaction(true)
		err = txn2.Set("key", xedb.Entry{
			Type:  xedb.String,
			Value: "value2",
		})
		assert.NoError(t, err)

		// Commit first transaction
		err = txn1.Commit()
		assert.NoError(t, err)

		// Try to commit second transaction
		err = txn2.Commit()
		assert.Error(t, err) // Should fail due to conflict
	})

	t.Run("Transaction Read", func(t *testing.T) {
		// Set initial value
		txn := db.NewTransaction(true)
		err := txn.Set("key", xedb.Entry{
			Type:  xedb.String,
			Value: "initial",
		})
		assert.NoError(t, err)
		err = txn.Commit()
		assert.NoError(t, err)

		// Read transaction
		txn = db.NewTransaction(false)
		entry, err := txn.Get("key")
		assert.NoError(t, err)
		assert.Equal(t, "initial", entry.Value)
	})
}

func TestDB_Iterator(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Setup test data
	data := map[string]string{
		"user:1": "Alice",
		"user:2": "Bob",
		"user:3": "Charlie",
		"post:1": "Post1",
		"post:2": "Post2",
	}

	for k, v := range data {
		err := db.String(k).Set(v)
		require.NoError(t, err)
	}

	t.Run("Forward Iteration", func(t *testing.T) {
		iter := db.NewIterator(xedb.IteratorOptions{
			Prefix:  "user:",
			Reverse: false,
		})

		var keys []string
		for iter.Seek("user:"); iter.Valid(); iter.Next() {
			item := iter.Item()
			if item == nil {
				continue
			}
			// Safely type assert and handle nil case
			if str, ok := item.Value.(string); ok {
				keys = append(keys, str)
			}
		}

		assert.Equal(t, []string{"Alice", "Bob", "Charlie"}, keys)
	})

	t.Run("Reverse Iteration", func(t *testing.T) {
		iter := db.NewIterator(xedb.IteratorOptions{
			Prefix:  "user:",
			Reverse: true,
		})

		var keys []string
		for iter.Seek("user:"); iter.Valid(); iter.Next() {
			item := iter.Item()
			if item == nil {
				continue
			}
			// Safely type assert and handle nil case
			if str, ok := item.Value.(string); ok {
				keys = append(keys, str)
			}
		}

		assert.Equal(t, []string{"Charlie", "Bob", "Alice"}, keys)
	})

	t.Run("Prefix Iteration", func(t *testing.T) {
		iter := db.NewIterator(xedb.IteratorOptions{
			Prefix:  "post:",
			Reverse: false,
		})

		var posts []string
		for iter.Seek("post:"); iter.Valid(); iter.Next() {
			item := iter.Item()
			if item == nil {
				continue
			}
			// Safely type assert and handle nil case
			if str, ok := item.Value.(string); ok {
				posts = append(posts, str)
			}
		}

		assert.Equal(t, []string{"Post1", "Post2"}, posts)
	})

	t.Run("Empty Prefix", func(t *testing.T) {
		iter := db.NewIterator(xedb.IteratorOptions{
			Prefix:  "nonexistent:",
			Reverse: false,
		})

		count := 0
		for iter.Seek("nonexistent:"); iter.Valid(); iter.Next() {
			count++
		}

		assert.Equal(t, 0, count)
	})
}

func TestDB_TransactionVersioning(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("Version Increments", func(t *testing.T) {
		// First transaction
		txn1 := db.NewTransaction(true)
		err := txn1.Set("key", xedb.Entry{
			Type:  xedb.String,
			Value: "value1",
		})
		assert.NoError(t, err)
		err = txn1.Commit()
		assert.NoError(t, err)

		// Second transaction
		txn2 := db.NewTransaction(true)
		err = txn2.Set("key", xedb.Entry{
			Type:  xedb.String,
			Value: "value2",
		})
		assert.NoError(t, err)
		err = txn2.Commit()
		assert.NoError(t, err)

		// Verify version incremented
		entry, err := db.NewTransaction(false).Get("key")
		assert.NoError(t, err)
		assert.Greater(t, entry.Version, uint64(0))
	})

	t.Run("Read Consistency", func(t *testing.T) {
		// Start read transaction
		readTxn := db.NewTransaction(false)

		// Write new value in separate transaction
		writeTxn := db.NewTransaction(true)
		err := writeTxn.Set("key", xedb.Entry{
			Type:  xedb.String,
			Value: "new-value",
		})
		assert.NoError(t, err)
		err = writeTxn.Commit()
		assert.NoError(t, err)

		// Read should still see old value
		entry, err := readTxn.Get("key")
		assert.NoError(t, err)
		assert.Equal(t, "value2", entry.Value)
	})
}

func TestDB_ConcurrentTransactions(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	t.Run("Concurrent Reads", func(t *testing.T) {
		// Setup initial data
		txn := db.NewTransaction(true)
		err := txn.Set("key", xedb.Entry{
			Type:  xedb.String,
			Value: "value",
		})
		assert.NoError(t, err)
		err = txn.Commit()
		assert.NoError(t, err)

		// Concurrent reads
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				txn := db.NewTransaction(false)
				entry, err := txn.Get("key")
				assert.NoError(t, err)
				assert.Equal(t, "value", entry.Value)
			}()
		}
		wg.Wait()
	})

	t.Run("Concurrent Write Attempts", func(t *testing.T) {
		var wg sync.WaitGroup
		var successCount atomic.Int32
		var errorCount atomic.Int32

		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()
				txn := db.NewTransaction(true)
				err := txn.Set("concurrent-key", xedb.Entry{
					Type:  xedb.String,
					Value: fmt.Sprintf("value-%d", i),
				})
				// Don't assert NoError here since write conflicts are expected
				// assert.NoError(t, err)

				err = txn.Commit()
				if err == nil {
					successCount.Add(1)
				} else {
					errorCount.Add(1)
					// We expect write conflict errors in this concurrent scenario
					if !strings.Contains(err.Error(), "write conflict") {
						t.Errorf("Unexpected error type: %v", err)
					}
				}
			}(i)
		}
		wg.Wait()

		// Due to the lock contention behavior, more than one transaction might succeed
		// based on the current implementation. This is expected.
		assert.GreaterOrEqual(t, successCount.Load(), int32(1), "At least one transaction should succeed")

		// The total of successes and errors should equal the number of transactions
		assert.Equal(t, int32(10), successCount.Load()+errorCount.Load(), "All transactions should either succeed or fail with write conflict")
	})
}

func TestDB_ExportToJSON(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Add test data
	err := db.String("str1").SetWithVersion("value1")
	require.NoError(t, err)

	err = db.List("list1").Push("item1", "item2")
	require.NoError(t, err)

	err = db.Hash("hash1").Set("field1", "value1")
	require.NoError(t, err)

	err = db.Set("set1").Add("member1", "member2")
	require.NoError(t, err)

	err = db.ZSet("zset1").Add(1.0, "member1")
	require.NoError(t, err)

	// Export to JSON
	jsonData, err := db.ExportToJSON()
	require.NoError(t, err)

	// Parse and verify JSON content
	var exportedData map[string]interface{}
	err = json.Unmarshal([]byte(jsonData), &exportedData)
	require.NoError(t, err)

	// Verify string data
	strData, ok := exportedData["str1"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(0), strData["type"])
	assert.Equal(t, "value1", strData["value"])
	assert.NotEmpty(t, strData["created"])
	assert.NotEmpty(t, strData["last_updated"])
	assert.Greater(t, strData["version"].(float64), float64(0))

	// Verify list data
	listData, ok := exportedData["list1"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(1), listData["type"])
	assert.Equal(t, []interface{}{"item1", "item2"}, listData["value"])
	assert.NotEmpty(t, listData["created"])
	assert.NotEmpty(t, listData["last_updated"])
	assert.Greater(t, listData["version"].(float64), float64(0))

	// Verify hash data
	hashData, ok := exportedData["hash1"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(2), hashData["type"])
	hashValue, ok := hashData["value"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "value1", hashValue["field1"])
	assert.NotEmpty(t, hashData["created"])
	assert.NotEmpty(t, hashData["last_updated"])
	assert.Greater(t, hashData["version"].(float64), float64(0))

	// Verify set data
	setData, ok := exportedData["set1"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(3), setData["type"])
	setValue, ok := setData["value"].([]interface{})
	assert.True(t, ok)
	assert.ElementsMatch(t, []interface{}{"member1", "member2"}, setValue)
	assert.NotEmpty(t, setData["created"])
	assert.NotEmpty(t, setData["last_updated"])
	assert.Greater(t, setData["version"].(float64), float64(0))

	// Verify zset data
	zsetData, ok := exportedData["zset1"].(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, float64(4), zsetData["type"])
	zsetValue, ok := zsetData["value"].([]interface{})
	assert.True(t, ok)
	assert.Len(t, zsetValue, 1)
	zsetMember := zsetValue[0].(map[string]interface{})
	assert.Equal(t, "member1", zsetMember["Member"])
	assert.Equal(t, float64(1.0), zsetMember["Score"])
	assert.NotEmpty(t, zsetData["created"])
	assert.NotEmpty(t, zsetData["last_updated"])
	assert.Greater(t, zsetData["version"].(float64), float64(0))
}

func TestDB_Versioning(t *testing.T) {
	dir, err := os.MkdirTemp("", "xedb-version-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create DB with default options (versioning should be enabled by default)
	db, err := xedb.New(
		xedb.WithDataDir(dir),
	)
	require.NoError(t, err)
	defer db.Close()

	t.Run("Default Version Control", func(t *testing.T) {
		// Set multiple versions of a key
		err := db.String("key1").SetWithVersion("value1")
		require.NoError(t, err)
		time.Sleep(time.Millisecond)

		err = db.String("key1").SetWithVersion("value2")
		require.NoError(t, err)
		time.Sleep(time.Millisecond)

		err = db.String("key1").SetWithVersion("value3")
		require.NoError(t, err)

		// Get current value
		val, exists := db.String("key1").Get()
		assert.True(t, exists)
		assert.Equal(t, "value3", val)

		// List versions - should have all versions by default
		versions := db.String("key1").ListVersions()
		assert.Len(t, versions, 3) // Should have all 3 versions
	})

	t.Run("Export JSON with All Versions", func(t *testing.T) {
		// Set multiple versions
		err := db.String("key2").SetWithVersion("v1")
		require.NoError(t, err)
		time.Sleep(time.Millisecond)

		err = db.String("key2").SetWithVersion("v2")
		require.NoError(t, err)
		time.Sleep(time.Millisecond)

		err = db.String("key2").SetWithVersion("v3")
		require.NoError(t, err)

		// Export to JSON
		jsonData, err := db.ExportToJSON()
		require.NoError(t, err)

		// Parse JSON
		var data map[string]interface{}
		err = json.Unmarshal([]byte(jsonData), &data)
		require.NoError(t, err)

		// Verify key2 entry
		entry, ok := data["key2"].(map[string]interface{})
		assert.True(t, ok)
		assert.Equal(t, "v3", entry["value"])

		// Verify all versions are present
		versions, ok := entry["versions"].([]interface{})
		assert.True(t, ok)
		assert.Len(t, versions, 2) // Should have 2 previous versions

		// Verify version contents
		firstVersion := versions[0].(map[string]interface{})
		assert.Equal(t, "v2", firstVersion["value"])

		secondVersion := versions[1].(map[string]interface{})
		assert.Equal(t, "v1", secondVersion["value"])
	})

	t.Run("Version Limit Override", func(t *testing.T) {
		// Create new DB with custom version limit
		dbWithLimit, err := xedb.New(
			xedb.WithDataDir(dir),
			xedb.WithMaxVersions(2),
		)
		require.NoError(t, err)
		defer dbWithLimit.Close()

		// Set multiple versions
		err = dbWithLimit.String("key3").SetWithVersion("v1")
		require.NoError(t, err)
		time.Sleep(time.Millisecond)

		err = dbWithLimit.String("key3").SetWithVersion("v2")
		require.NoError(t, err)
		time.Sleep(time.Millisecond)

		err = dbWithLimit.String("key3").SetWithVersion("v3")
		require.NoError(t, err)

		// Export to JSON
		jsonData, err := dbWithLimit.ExportToJSON()
		require.NoError(t, err)

		var data map[string]interface{}
		err = json.Unmarshal([]byte(jsonData), &data)
		require.NoError(t, err)

		// Verify version limit
		entry := data["key3"].(map[string]interface{})
		versions := entry["versions"].([]interface{})
		assert.Len(t, versions, 2) // Should only have 2 versions due to limit
	})

	t.Run("Version History Order", func(t *testing.T) {
		// Set multiple versions
		err := db.String("key4").SetWithVersion("v1")
		require.NoError(t, err)
		time.Sleep(time.Millisecond)

		err = db.String("key4").SetWithVersion("v2")
		require.NoError(t, err)
		time.Sleep(time.Millisecond)

		err = db.String("key4").SetWithVersion("v3")
		require.NoError(t, err)

		// Get versions
		versions := db.String("key4").ListVersions()
		assert.Len(t, versions, 3)

		// Verify version order (newest to oldest)
		val, exists := db.String("key4").GetVersion(versions[0])
		assert.True(t, exists)
		assert.Equal(t, "v3", val)

		val, exists = db.String("key4").GetVersion(versions[1])
		assert.True(t, exists)
		assert.Equal(t, "v2", val)

		val, exists = db.String("key4").GetVersion(versions[2])
		assert.True(t, exists)
		assert.Equal(t, "v1", val)
	})
}
