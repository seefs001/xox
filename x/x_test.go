package x_test

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/url"
	"os"
	"reflect"
	"sort"
	"sync"
	"testing"
	"time"
	"unicode"

	"github.com/seefs001/xox/x"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMustFunctions(t *testing.T) {
	t.Run("Must0", func(t *testing.T) {
		x.Must0(nil) // Should not panic

		assert.Panics(t, func() {
			x.Must0(errors.New("test error"))
		})
	})

	t.Run("Must1", func(t *testing.T) {
		result := x.Must1(42, nil)
		assert.Equal(t, 42, result)

		assert.Panics(t, func() {
			x.Must1(0, errors.New("test error"))
		})
	})

	t.Run("Must2", func(t *testing.T) {
		v1, v2 := x.Must2(1, "two", nil)
		assert.Equal(t, 1, v1)
		assert.Equal(t, "two", v2)

		assert.Panics(t, func() {
			x.Must2(0, "", errors.New("test error"))
		})
	})

	t.Run("Must3", func(t *testing.T) {
		v1, v2, v3 := x.Must3(1, "two", true, nil)
		assert.Equal(t, 1, v1)
		assert.Equal(t, "two", v2)
		assert.True(t, v3)

		assert.Panics(t, func() {
			x.Must3(0, "", false, errors.New("test error"))
		})
	})

	t.Run("Must4", func(t *testing.T) {
		v1, v2, v3, v4 := x.Must4(1, "two", true, 4.0, nil)
		assert.Equal(t, 1, v1)
		assert.Equal(t, "two", v2)
		assert.True(t, v3)
		assert.Equal(t, 4.0, v4)

		assert.Panics(t, func() {
			x.Must4(0, "", false, 0.0, errors.New("test error"))
		})
	})
}

func TestIgnoreFunctions(t *testing.T) {
	t.Run("Ignore0", func(t *testing.T) {
		x.Ignore0(errors.New("ignored error"))
		// No assertion needed as Ignore0 doesn't return anything
	})

	t.Run("Ignore1", func(t *testing.T) {
		result := x.Ignore1(42, errors.New("ignored error"))
		assert.Equal(t, 42, result)
	})

	t.Run("Ignore2", func(t *testing.T) {
		v1, v2 := x.Ignore2(1, "two", errors.New("ignored error"))
		assert.Equal(t, 1, v1)
		assert.Equal(t, "two", v2)
	})

	t.Run("Ignore3", func(t *testing.T) {
		v1, v2, v3 := x.Ignore3(1, "two", true, errors.New("ignored error"))
		assert.Equal(t, 1, v1)
		assert.Equal(t, "two", v2)
		assert.True(t, v3)
	})

	t.Run("Ignore4", func(t *testing.T) {
		v1, v2, v3, v4 := x.Ignore4(1, "two", true, 4.0, errors.New("ignored error"))
		assert.Equal(t, 1, v1)
		assert.Equal(t, "two", v2)
		assert.True(t, v3)
		assert.Equal(t, 4.0, v4)
	})
}

func TestWhere(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	result := x.Where(numbers, func(n int) bool {
		return n%2 == 0
	})
	expected := []int{2, 4}
	assert.Equal(t, expected, result)
}

func TestSelect(t *testing.T) {
	numbers := []int{1, 2, 3}
	result := x.Select(numbers, func(n int) string {
		return string(rune('A' + n - 1))
	})
	expected := []string{"A", "B", "C"}
	assert.Equal(t, expected, result)
}

func TestAggregate(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	result := x.Aggregate(numbers, 0, func(acc, n int) int {
		return acc + n
	})
	assert.Equal(t, 15, result)
}

func TestForEach(t *testing.T) {
	numbers := []int{1, 2, 3}
	sum := 0
	x.ForEach(numbers, func(n int) {
		sum += n
	})
	assert.Equal(t, 6, sum)
}

func TestRange(t *testing.T) {
	result := x.Range(1, 5)
	expected := []int{1, 2, 3, 4, 5}
	assert.Equal(t, expected, result)
}

func TestCount(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	result := x.Count(numbers, func(n int) bool {
		return n%2 == 0
	})
	assert.Equal(t, 2, result)
}

func TestGroupBy(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	result := x.GroupBy(numbers, func(n int) string {
		if n%2 == 0 {
			return "even"
		}
		return "odd"
	})
	assert.Len(t, result, 2)
	assert.Equal(t, []int{2, 4}, result["even"])
	assert.Equal(t, []int{1, 3, 5}, result["odd"])
}

func TestFirst(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	result, ok := x.First(numbers, func(n int) bool {
		return n%2 == 0
	})
	assert.True(t, ok)
	assert.Equal(t, 2, result)

	_, ok = x.First(numbers, func(n int) bool {
		return n > 10
	})
	assert.False(t, ok)
}

func TestLast(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	result, ok := x.Last(numbers, func(n int) bool {
		return n%2 == 0
	})
	assert.True(t, ok)
	assert.Equal(t, 4, result)

	_, ok = x.Last(numbers, func(n int) bool {
		return n > 10
	})
	assert.False(t, ok)
}

func TestAny(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	result := x.Any(numbers, func(n int) bool {
		return n%2 == 0
	})
	assert.True(t, result)

	result = x.Any(numbers, func(n int) bool {
		return n > 10
	})
	assert.False(t, result)
}

func TestAll(t *testing.T) {
	numbers := []int{2, 4, 6, 8}
	result := x.All(numbers, func(n int) bool {
		return n%2 == 0
	})
	assert.True(t, result)

	numbers = append(numbers, 9)
	result = x.All(numbers, func(n int) bool {
		return n%2 == 0
	})
	assert.False(t, result)
}

func TestRandomString(t *testing.T) {
	t.Run("ModeAlphanumeric", func(t *testing.T) {
		result, err := x.RandomString(10, x.ModeAlphanumeric)
		require.NoError(t, err)
		assert.Len(t, result, 10)
		assert.True(t, isAlphanumeric(result))
	})

	t.Run("ModeAlpha", func(t *testing.T) {
		result, err := x.RandomString(10, x.ModeAlpha)
		require.NoError(t, err)
		assert.Len(t, result, 10)
		assert.True(t, isAlpha(result))
	})

	t.Run("ModeNumeric", func(t *testing.T) {
		result, err := x.RandomString(10, x.ModeNumeric)
		require.NoError(t, err)
		assert.Len(t, result, 10)
		assert.True(t, isNumeric(result))
	})

	t.Run("ModeLowercase", func(t *testing.T) {
		result, err := x.RandomString(10, x.ModeLowercase)
		require.NoError(t, err)
		assert.Len(t, result, 10)
		assert.True(t, isLowercase(result))
	})

	t.Run("ModeUppercase", func(t *testing.T) {
		result, err := x.RandomString(10, x.ModeUppercase)
		require.NoError(t, err)
		assert.Len(t, result, 10)
		assert.True(t, isUppercase(result))
	})

	t.Run("Invalid mode", func(t *testing.T) {
		_, err := x.RandomString(10, x.RandomStringMode(999))
		assert.Error(t, err)
	})

	t.Run("Zero length", func(t *testing.T) {
		_, err := x.RandomString(0, x.ModeAlphanumeric)
		assert.Error(t, err)
	})
}

func TestRandomStringCustom(t *testing.T) {
	t.Run("Valid charset", func(t *testing.T) {
		charset := "ABC123"
		result, err := x.RandomStringCustom(10, charset)
		require.NoError(t, err)
		assert.Len(t, result, 10)
		for _, char := range result {
			assert.Contains(t, charset, string(char))
		}
	})

	t.Run("Empty charset", func(t *testing.T) {
		_, err := x.RandomStringCustom(10, "")
		assert.Error(t, err)
	})

	t.Run("Zero length", func(t *testing.T) {
		_, err := x.RandomStringCustom(0, "ABC")
		assert.Error(t, err)
	})
}

func TestTuple(t *testing.T) {
	t.Run("NewTuple", func(t *testing.T) {
		tuple := x.NewTuple(1, "hello")
		assert.Equal(t, 1, tuple.First)
		assert.Equal(t, "hello", tuple.Second)
	})

	t.Run("Unpack", func(t *testing.T) {
		tuple := x.NewTuple(1, "hello")
		first, second := tuple.Unpack()
		assert.Equal(t, 1, first)
		assert.Equal(t, "hello", second)
	})

	t.Run("Swap", func(t *testing.T) {
		tuple := x.NewTuple(1, "hello")
		swapped := tuple.Swap()
		assert.Equal(t, "hello", swapped.First)
		assert.Equal(t, 1, swapped.Second)
	})

	t.Run("Map", func(t *testing.T) {
		tuple := x.NewTuple(1, 2)
		mapped := tuple.Map(func(a, b int) (int, int) {
			return a * 2, b * 3
		})
		assert.Equal(t, 2, mapped.First)
		assert.Equal(t, 6, mapped.Second)
	})
}

func TestTriple(t *testing.T) {
	t.Run("NewTriple", func(t *testing.T) {
		triple := x.NewTriple(1, "hello", true)
		assert.Equal(t, 1, triple.First)
		assert.Equal(t, "hello", triple.Second)
		assert.True(t, triple.Third)
	})

	t.Run("Unpack", func(t *testing.T) {
		triple := x.NewTriple(1, "hello", true)
		first, second, third := triple.Unpack()
		assert.Equal(t, 1, first)
		assert.Equal(t, "hello", second)
		assert.True(t, third)
	})

	t.Run("Rotate", func(t *testing.T) {
		triple := x.NewTriple(1, "hello", true)
		rotated := triple.Rotate()
		assert.Equal(t, "hello", rotated.First)
		assert.True(t, rotated.Second)
		assert.Equal(t, 1, rotated.Third)
	})

	t.Run("Map", func(t *testing.T) {
		triple := x.NewTriple(1, 2, 3)
		mapped := triple.Map(func(a, b, c int) (int, int, int) {
			return a * 2, b * 3, c * 4
		})
		assert.Equal(t, 2, mapped.First)
		assert.Equal(t, 6, mapped.Second)
		assert.Equal(t, 12, mapped.Third)
	})
}

func TestMap(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	result := x.Map(numbers, func(n int) int {
		return n * 2
	})
	expected := []int{2, 4, 6, 8, 10}
	assert.Equal(t, expected, result)
}

func TestFilter(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	result := x.Filter(numbers, func(n int) bool {
		return n%2 == 0
	})
	expected := []int{2, 4}
	assert.Equal(t, expected, result)
}

func TestReduce(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	result := x.Reduce(numbers, 0, func(acc, n int) int {
		return acc + n
	})
	assert.Equal(t, 15, result)
}

func TestContains(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	assert.True(t, x.Contains(numbers, 3))
	assert.False(t, x.Contains(numbers, 6))
}

func TestUnique(t *testing.T) {
	numbers := []int{1, 2, 2, 3, 3, 4, 5, 5}
	result := x.Unique(numbers)
	expected := []int{1, 2, 3, 4, 5}
	assert.Equal(t, expected, result)
}

func TestChunk(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5, 6, 7}
	result := x.Chunk(numbers, 3)
	expected := [][]int{{1, 2, 3}, {4, 5, 6}, {7}}
	assert.Equal(t, expected, result)
}

func TestFlatten(t *testing.T) {
	nestedSlice := [][]int{{1, 2}, {3, 4}, {5, 6}}
	result := x.Flatten(nestedSlice)
	expected := []int{1, 2, 3, 4, 5, 6}
	assert.Equal(t, expected, result)
}

func TestZip(t *testing.T) {
	slice1 := []int{1, 2, 3}
	slice2 := []string{"one", "two", "three"}
	result := x.Zip(slice1, slice2)
	assert.Len(t, result, 3)
	expectedPairs := []struct {
		First  int
		Second string
	}{
		{1, "one"},
		{2, "two"},
		{3, "three"},
	}
	for i, pair := range expectedPairs {
		assert.Equal(t, pair.First, result[i].First)
		assert.Equal(t, pair.Second, result[i].Second)
	}
}

func TestReverse(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	result := x.Reverse(numbers)
	expected := []int{5, 4, 3, 2, 1}
	assert.Equal(t, expected, result)
}

func TestParallelFor(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	results := make([]int, 5)
	var mutex sync.Mutex

	x.ParallelFor(numbers, func(n int) {
		mutex.Lock()
		defer mutex.Unlock()
		results[n-1] = n * 2
	})

	expected := []int{2, 4, 6, 8, 10}
	assert.Equal(t, expected, results)
}

func TestParallelMap(t *testing.T) {
	numbers := []int{1, 2, 3, 4, 5}
	result := x.ParallelMap(numbers, func(n int) int {
		return n * 2
	})

	expected := []int{2, 4, 6, 8, 10}
	sort.Ints(result) // Sort because parallel execution may change order
	assert.Equal(t, expected, result)
}

func TestDebounce(t *testing.T) {
	var mu sync.Mutex
	counter := 0
	f := func() {
		mu.Lock()
		defer mu.Unlock()
		counter++
	}

	debounced := x.Debounce(f, 50*time.Millisecond)

	for i := 0; i < 10; i++ {
		debounced()
	}

	time.Sleep(100 * time.Millisecond)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, 1, counter)
}

func TestThrottle(t *testing.T) {
	counter := 0
	f := func() {
		counter++
	}

	throttled := x.Throttle(f, 50*time.Millisecond)

	for i := 0; i < 10; i++ {
		throttled()
		time.Sleep(10 * time.Millisecond)
	}

	time.Sleep(60 * time.Millisecond)

	assert.GreaterOrEqual(t, counter, 2)
	assert.LessOrEqual(t, counter, 4)
}

func TestAsyncTask(t *testing.T) {
	t.Run("Successful task", func(t *testing.T) {
		task := x.NewAsyncTask(func() (int, error) {
			time.Sleep(50 * time.Millisecond)
			return 42, nil
		})

		result, err := task.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 42, result)
	})

	t.Run("Task with timeout", func(t *testing.T) {
		task := x.NewAsyncTask(func() (int, error) {
			time.Sleep(100 * time.Millisecond)
			return 42, nil
		})

		result, err, completed := task.WaitWithTimeout(50 * time.Millisecond)
		assert.False(t, completed)
		assert.NoError(t, err)
		assert.Equal(t, 0, result)
	})

	t.Run("Task with context", func(t *testing.T) {
		task := x.NewAsyncTask(func() (int, error) {
			time.Sleep(100 * time.Millisecond)
			return 42, nil
		})

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		result, err, completed := task.WaitWithContext(ctx)
		assert.False(t, completed)
		assert.Equal(t, context.DeadlineExceeded, err)
		assert.Equal(t, 0, result)
	})
}

func isAlphanumeric(s string) bool {
	for _, char := range s {
		if !unicode.IsLetter(char) && !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}

func isAlpha(s string) bool {
	for _, char := range s {
		if !unicode.IsLetter(char) {
			return false
		}
	}
	return true
}

func isNumeric(s string) bool {
	for _, char := range s {
		if !unicode.IsDigit(char) {
			return false
		}
	}
	return true
}

func isLowercase(s string) bool {
	for _, char := range s {
		if !unicode.IsLower(char) {
			return false
		}
	}
	return true
}

func isUppercase(s string) bool {
	for _, char := range s {
		if !unicode.IsUpper(char) {
			return false
		}
	}
	return true
}

func TestSafePool(t *testing.T) {
	t.Run("Bounded pool", func(t *testing.T) {
		pool := x.NewSafePool(2)
		var wg sync.WaitGroup
		wg.Add(3)

		start := time.Now()
		for i := 0; i < 3; i++ {
			pool.SafeGoNoError(func() {
				time.Sleep(100 * time.Millisecond)
				wg.Done()
			})
		}
		wg.Wait()
		duration := time.Since(start)

		assert.True(t, duration >= 190*time.Millisecond && duration < 210*time.Millisecond)
	})

	// FIXME
	// t.Run("Unbounded pool", func(t *testing.T) {
	// 	pool := x.NewSafePool(0)
	// 	var wg sync.WaitGroup
	// 	wg.Add(3)

	// 	start := time.Now()
	// 	for i := 0; i < 3; i++ {
	// 		pool.SafeGoNoError(func() {
	// 			time.Sleep(100 * time.Millisecond)
	// 			wg.Done()
	// 		})
	// 	}
	// 	wg.Wait()
	// 	duration := time.Since(start)

	// 	assert.True(t, duration >= 90*time.Millisecond && duration < 110*time.Millisecond)
	// })
}

func TestSafeGo(t *testing.T) {
	pool := x.NewSafePool(0)

	// Test normal execution
	errChan := pool.SafeGo(func() {
		// Do nothing
	})

	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Error("Timeout waiting for SafeGo to complete")
	}

	// Test panic recovery
	errChan = pool.SafeGo(func() {
		panic("test panic")
	})

	select {
	case err := <-errChan:
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "panic recovered: test panic")
	case <-time.After(time.Second):
		t.Error("Timeout waiting for SafeGo to complete")
	}
}

func TestSafeGoWithContext(t *testing.T) {
	pool := x.NewSafePool(0)

	// Test normal execution
	ctx := context.Background()
	errChan := pool.SafeGoWithContext(ctx, func() {
		// Do nothing
	})

	select {
	case err := <-errChan:
		assert.NoError(t, err)
	case <-time.After(time.Second):
		t.Error("Timeout waiting for SafeGoWithContext to complete")
	}

	// Test panic recovery
	errChan = pool.SafeGoWithContext(ctx, func() {
		panic("test panic")
	})

	select {
	case err := <-errChan:
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "panic recovered: test panic")
	case <-time.After(time.Second):
		t.Error("Timeout waiting for SafeGoWithContext to complete")
	}

	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	errChan = pool.SafeGoWithContext(ctx, func() {
		time.Sleep(time.Second)
	})
	cancel()

	select {
	case err := <-errChan:
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for SafeGoWithContext to complete")
	}
}

func TestSafeGoNoError(t *testing.T) {
	pool := x.NewSafePool(0)

	// Test normal execution
	done := make(chan bool)
	pool.SafeGoNoError(func() {
		done <- true
	})

	select {
	case <-done:
		// Test passed
	case <-time.After(time.Second):
		t.Error("Timeout waiting for SafeGoNoError to complete")
	}

	// Test panic recovery (no error channel, so we can't check the error)
	pool.SafeGoNoError(func() {
		panic("test panic")
	})

	// Wait a bit to ensure the panic doesn't crash the program
	time.Sleep(100 * time.Millisecond)
}

func TestSafeGoWithContextNoError(t *testing.T) {
	pool := x.NewSafePool(0)

	// Test normal execution
	ctx := context.Background()
	done := make(chan bool)
	pool.SafeGoWithContextNoError(ctx, func() {
		done <- true
	})

	select {
	case <-done:
		// Test passed
	case <-time.After(time.Second):
		t.Error("Timeout waiting for SafeGoWithContextNoError to complete")
	}

	// Test panic recovery (no error channel, so we can't check the error)
	pool.SafeGoWithContextNoError(ctx, func() {
		panic("test panic")
	})

	// Wait a bit to ensure the panic doesn't crash the program
	time.Sleep(100 * time.Millisecond)

	// Test context cancellation
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately
	pool.SafeGoWithContextNoError(ctx, func() {
		t.Error("This should not be executed")
	})

	// Wait a bit to ensure the cancelled context prevents execution
	time.Sleep(100 * time.Millisecond)
}

func TestWaitGroup(t *testing.T) {
	t.Run("Basic functionality", func(t *testing.T) {
		wg := x.NewWaitGroup()
		var mu sync.Mutex
		counter := 0
		for i := 0; i < 5; i++ {
			wg.Go(func() error {
				mu.Lock()
				defer mu.Unlock()
				counter++
				return nil
			})
		}
		err := wg.Wait()
		assert.NoError(t, err)
		assert.Equal(t, 5, counter)
	})

	t.Run("Error handling", func(t *testing.T) {
		wg := x.NewWaitGroup()
		wg.Go(func() error {
			return errors.New("test error 1")
		})
		wg.Go(func() error {
			return errors.New("test error 2")
		})
		err := wg.Wait()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "test error 1")
		assert.Contains(t, err.Error(), "test error 2")
	})

	t.Run("Panic recovery", func(t *testing.T) {
		wg := x.NewWaitGroup()
		wg.Go(func() error {
			panic("test panic")
		})
		err := wg.Wait()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "panic recovered: test panic")
	})

	t.Run("Context support", func(t *testing.T) {
		wg := x.NewWaitGroup()
		ctx, cancel := context.WithCancel(context.Background())
		wg.GoWithContext(ctx, func(ctx context.Context) error {
			<-ctx.Done()
			return ctx.Err()
		})
		cancel()
		err := wg.Wait()
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})

	t.Run("Wait with timeout", func(t *testing.T) {
		wg := x.NewWaitGroup()
		wg.Go(func() error {
			time.Sleep(200 * time.Millisecond)
			return nil
		})
		err, timedOut := wg.WaitWithTimeout(100 * time.Millisecond)
		assert.NoError(t, err)
		assert.True(t, timedOut)
	})

	t.Run("Wait with timeout - success", func(t *testing.T) {
		wg := x.NewWaitGroup()
		wg.Go(func() error {
			time.Sleep(50 * time.Millisecond)
			return nil
		})
		err, timedOut := wg.WaitWithTimeout(100 * time.Millisecond)
		assert.NoError(t, err)
		assert.False(t, timedOut)
	})
}

func TestToJSON(t *testing.T) {
	t.Run("Valid struct", func(t *testing.T) {
		type TestStruct struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}
		ts := TestStruct{Name: "John", Age: 30}
		json, err := x.ToJSON(ts)
		assert.NoError(t, err)
		assert.Equal(t, `{"name":"John","age":30}`, json)
	})

	t.Run("Invalid input", func(t *testing.T) {
		_, err := x.ToJSON(make(chan int))
		assert.Error(t, err)
	})
}

func TestMustToJSON(t *testing.T) {
	t.Run("Valid struct", func(t *testing.T) {
		type TestStruct struct {
			Name string `json:"name"`
		}
		ts := TestStruct{Name: "John"}
		assert.NotPanics(t, func() {
			json := x.MustToJSON(ts)
			assert.Equal(t, `{"name":"John"}`, json)
		})
	})

	t.Run("Invalid input", func(t *testing.T) {
		assert.Panics(t, func() {
			x.MustToJSON(make(chan int))
		})
	})
}

func TestTernary(t *testing.T) {
	assert.Equal(t, 1, x.Ternary(true, 1, 2))
	assert.Equal(t, 2, x.Ternary(false, 1, 2))
}

func TestTernaryF(t *testing.T) {
	assert.Equal(t, 1, x.TernaryF(true, func() int { return 1 }, func() int { return 2 }))
	assert.Equal(t, 2, x.TernaryF(false, func() int { return 1 }, func() int { return 2 }))
}

func TestIf(t *testing.T) {
	result := x.If(true, 1).
		ElseIf(false, 2).
		Else(3)
	assert.Equal(t, 1, result)

	result = x.If(false, 1).
		ElseIf(true, 2).
		Else(3)
	assert.Equal(t, 2, result)

	result = x.If(false, 1).
		ElseIf(false, 2).
		Else(3)
	assert.Equal(t, 3, result)
}

func TestSwitch(t *testing.T) {
	result := x.Switch[string, int]("b").
		Case("a", 1).
		Case("b", 2).
		Default(3)
	assert.Equal(t, 2, result)

	result = x.Switch[string, int]("c").
		Case("a", 1).
		Case("b", 2).
		Default(3)
	assert.Equal(t, 3, result)
}

func TestWaitGroupEdgeCases(t *testing.T) {
	t.Run("Empty WaitGroup", func(t *testing.T) {
		wg := x.NewWaitGroup()
		err := wg.Wait()
		assert.NoError(t, err)
	})

	t.Run("Multiple errors", func(t *testing.T) {
		wg := x.NewWaitGroup()
		for i := 0; i < 3; i++ {
			i := i
			wg.Go(func() error {
				return fmt.Errorf("error %d", i)
			})
		}
		err := wg.Wait()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "error 0")
		assert.Contains(t, err.Error(), "error 1")
		assert.Contains(t, err.Error(), "error 2")
	})

	t.Run("Mix of success and errors", func(t *testing.T) {
		wg := x.NewWaitGroup()
		wg.Go(func() error {
			return nil
		})
		wg.Go(func() error {
			return errors.New("test error")
		})
		err := wg.Wait()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "test error")
	})
}

func TestSafePoolEdgeCases(t *testing.T) {
	t.Run("Negative pool size", func(t *testing.T) {
		pool := x.NewSafePool(-1)
		assert.NotNil(t, pool)
		var wg sync.WaitGroup
		wg.Add(3)
		start := time.Now()
		for i := 0; i < 3; i++ {
			pool.SafeGoNoError(func() {
				time.Sleep(100 * time.Millisecond)
				wg.Done()
			})
		}
		wg.Wait()
		duration := time.Since(start)
		assert.True(t, duration >= 90*time.Millisecond && duration < 110*time.Millisecond)
	})
}

func TestIsEmpty(t *testing.T) {
	t.Run("Empty values", func(t *testing.T) {
		assert.True(t, x.IsEmpty(""))
		assert.True(t, x.IsEmpty(0))
		assert.True(t, x.IsEmpty(false))
		assert.True(t, x.IsEmpty([]int{}))
		assert.True(t, x.IsEmpty(map[string]int{}))
		var nilSlice []int
		assert.True(t, x.IsEmpty(nilSlice))
		var nilMap map[string]int
		assert.True(t, x.IsEmpty(nilMap))
		var nilPtr *int
		assert.True(t, x.IsEmpty(nilPtr))
	})

	t.Run("Non-empty values", func(t *testing.T) {
		assert.False(t, x.IsEmpty("hello"))
		assert.False(t, x.IsEmpty(42))
		assert.False(t, x.IsEmpty(true))
		assert.False(t, x.IsEmpty([]int{1, 2, 3}))
		assert.False(t, x.IsEmpty(map[string]int{"a": 1}))
	})
}

func TestIsNil(t *testing.T) {
	t.Run("Nil values", func(t *testing.T) {
		var nilPtr *int
		assert.True(t, x.IsNil(nilPtr))
		var nilSlice []int
		assert.True(t, x.IsNil(nilSlice))
		var nilMap map[string]int
		assert.True(t, x.IsNil(nilMap))
		var nilInterface interface{}
		assert.True(t, x.IsNil(nilInterface))
		assert.True(t, x.IsNil(nilInterface)) // 替换原来的 x.IsNil(nil)
	})

	t.Run("Non-nil values", func(t *testing.T) {
		intPtr := new(int)
		assert.False(t, x.IsNil(intPtr))
		assert.False(t, x.IsNil([]int{}))
		assert.False(t, x.IsNil(map[string]int{}))
		assert.False(t, x.IsNil(""))
		assert.False(t, x.IsNil(0))
		assert.False(t, x.IsNil(false))
	})
}

func TestIsZero(t *testing.T) {
	t.Run("Zero values", func(t *testing.T) {
		assert.True(t, x.IsZero(""))
		assert.True(t, x.IsZero(0))
		assert.True(t, x.IsZero(false))
		assert.True(t, x.IsZero([]int(nil)))
		assert.True(t, x.IsZero(map[string]int(nil)))
		var zeroStruct struct{}
		assert.True(t, x.IsZero(zeroStruct))
	})

	t.Run("Non-zero values", func(t *testing.T) {
		assert.False(t, x.IsZero("hello"))
		assert.False(t, x.IsZero(42))
		assert.False(t, x.IsZero(true))
		assert.False(t, x.IsZero([]int{1, 2, 3}))
		assert.False(t, x.IsZero(map[string]int{"a": 1}))
		nonZeroStruct := struct{ Value int }{Value: 1}
		assert.False(t, x.IsZero(nonZeroStruct))
	})
}

func TestOnlyErr(t *testing.T) {
	t.Run("No error", func(t *testing.T) {
		err := x.OnlyErr("some value", nil)
		assert.NoError(t, err)
	})

	t.Run("With error", func(t *testing.T) {
		expectedErr := errors.New("test error")
		err := x.OnlyErr("some value", expectedErr)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("No values", func(t *testing.T) {
		err := x.OnlyErr()
		assert.NoError(t, err)
	})

	t.Run("Multiple return values", func(t *testing.T) {
		err := x.OnlyErr(1, "string", true, nil)
		assert.NoError(t, err)

		expectedErr := errors.New("multiple values error")
		err = x.OnlyErr(1, "string", true, expectedErr)
		assert.Equal(t, expectedErr, err)
	})

	t.Run("Function with no error", func(t *testing.T) {
		f := func() (int, string, error) {
			return 42, "hello", nil
		}
		err := x.OnlyErr(f())
		assert.NoError(t, err)
	})

	t.Run("Function with error", func(t *testing.T) {
		f := func() (int, string, error) {
			return 0, "", errors.New("function error")
		}
		err := x.OnlyErr(f())
		assert.EqualError(t, err, "function error")
	})

	t.Run("Function with multiple return values and error", func(t *testing.T) {
		f := func() (int, string, bool, error) {
			return 42, "hello", true, errors.New("complex function error")
		}
		err := x.OnlyErr(f())
		assert.EqualError(t, err, "complex function error")
	})
}

func TestIsImageURL(t *testing.T) {
	tests := []struct {
		url      string
		expected bool
	}{
		{"https://example.com/image.jpg", true},
		{"http://example.com/image.png", true},
		{"https://example.com/image.gif", true},
		{"https://example.com/file.txt", false},
		{"not-a-url", false},
	}

	for _, test := range tests {
		t.Run(test.url, func(t *testing.T) {
			result := x.IsImageURL(test.url)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestIsBase64(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"SGVsbG8gV29ybGQ=", true},
		{"Invalid base64", false},
		{"", true}, // Empty string is valid base64
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := x.IsBase64(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestTrimSuffixes(t *testing.T) {
	tests := []struct {
		input    string
		suffixes []string
		expected string
	}{
		{"example.tar.gz", []string{".gz", ".tar"}, "example"},
		{"file.txt", []string{".pdf", ".doc"}, "file.txt"},
		{"multiple.suffixes.here", []string{".here", ".there"}, "multiple.suffixes"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := x.TrimSuffixes(test.input, test.suffixes...)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestSetNonZeroValues(t *testing.T) {
	t.Run("Basic usage", func(t *testing.T) {
		dst := map[string]interface{}{"a": 1, "b": ""}
		src := map[string]interface{}{"a": 2, "b": "hello", "c": 3}
		x.SetNonZeroValues(dst, src)
		expected := map[string]interface{}{"a": 2, "b": "hello", "c": 3}
		assert.Equal(t, expected, dst)
	})

	t.Run("With zero values", func(t *testing.T) {
		dst := map[string]interface{}{"a": 1, "b": "value"}
		src := map[string]interface{}{"a": 0, "b": "", "c": false}
		x.SetNonZeroValues(dst, src)
		expected := map[string]interface{}{"a": 1, "b": "value"}
		assert.Equal(t, expected, dst)
	})
}

func TestSetNonZeroValuesWithKeys(t *testing.T) {
	t.Run("Basic usage", func(t *testing.T) {
		dst := map[string]interface{}{"a": 1, "b": ""}
		src := map[string]interface{}{"a": 2, "b": "hello", "c": 3}
		x.SetNonZeroValuesWithKeys(dst, src, "a", "c")
		expected := map[string]interface{}{"a": 2, "b": "", "c": 3}
		assert.Equal(t, expected, dst)
	})

	t.Run("With non-existent keys", func(t *testing.T) {
		dst := map[string]interface{}{"a": 1, "b": "value"}
		src := map[string]interface{}{"a": 2, "c": 3}
		x.SetNonZeroValuesWithKeys(dst, src, "b", "c", "d")
		expected := map[string]interface{}{"a": 1, "b": "value", "c": 3}
		assert.Equal(t, expected, dst)
	})
}

func TestMapKeys(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	keys := x.MapKeys(m)
	assert.ElementsMatch(t, []string{"a", "b", "c"}, keys)
}

func TestMapValues(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	values := x.MapValues(m)
	assert.ElementsMatch(t, []int{1, 2, 3}, values)
}

func TestShuffle(t *testing.T) {
	original := []int{1, 2, 3, 4, 5}
	shuffled := make([]int, len(original))
	copy(shuffled, original)
	x.Shuffle(shuffled)

	assert.NotEqual(t, original, shuffled)
	assert.ElementsMatch(t, original, shuffled)
}

func TestFlattenMap(t *testing.T) {
	input := map[string]any{
		"a": 1,
		"b": map[string]any{
			"c": 2,
			"d": map[string]any{
				"e": 3,
			},
		},
	}
	expected := map[string]any{
		"a":     1,
		"b.c":   2,
		"b.d.e": 3,
	}
	result := x.FlattenMap(input, "")
	assert.Equal(t, expected, result)
}

func TestCopyMap(t *testing.T) {
	original := map[string]int{"a": 1, "b": 2}
	copied := x.CopyMap(original)

	assert.Equal(t, original, copied)
	assert.NotSame(t, original, copied)

	copied["c"] = 3
	assert.NotEqual(t, original, copied)
}

func TestFileExists(t *testing.T) {
	t.Run("Existing file", func(t *testing.T) {
		tempFile, err := os.CreateTemp("", "test")
		require.NoError(t, err)
		defer os.Remove(tempFile.Name())

		assert.True(t, x.FileExists(tempFile.Name()))
	})

	t.Run("Non-existing file", func(t *testing.T) {
		assert.False(t, x.FileExists("non_existing_file.txt"))
	})
}

func TestGenerateUUID(t *testing.T) {
	uuid1, err := x.GenerateUUID()
	assert.NoError(t, err)
	assert.Regexp(t, "^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$", uuid1)

	uuid2, err := x.GenerateUUID()
	assert.NoError(t, err)
	assert.NotEqual(t, uuid1, uuid2)
}

func TestDecodeUnicodeURL(t *testing.T) {
	tests := []struct {
		encoded  string
		expected string
	}{
		{"https://example.com/path?q=%E4%BD%A0%E5%A5%BD", "https://example.com/path?q=你好"},
		{"https://example.com/no-encoding", "https://example.com/no-encoding"},
	}

	for _, test := range tests {
		t.Run(test.encoded, func(t *testing.T) {
			decoded, err := x.DecodeUnicodeURL(test.encoded)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, decoded)
		})
	}
}

func TestEncodeUnicodeURL(t *testing.T) {
	tests := []struct {
		original string
		expected string
	}{
		{"https://example.com/path?q=你好", "https%3A%2F%2Fexample.com%2Fpath%3Fq%3D%E4%BD%A0%E5%A5%BD"},
		{"https://example.com/no-encoding", "https%3A%2F%2Fexample.com%2Fno-encoding"},
	}

	for _, test := range tests {
		t.Run(test.original, func(t *testing.T) {
			encoded, err := x.EncodeUnicodeURL(test.original)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, encoded)
		})
	}
}

func TestBindData(t *testing.T) {
	type TestStruct struct {
		Name  string `form:"name"`
		Age   int    `form:"age"`
		Email string
	}

	t.Run("Valid binding", func(t *testing.T) {
		data := map[string][]string{
			"name": {"John"},
			"age":  {"30"},
		}
		var result TestStruct
		err := x.BindData(&result, data)
		assert.NoError(t, err)
		assert.Equal(t, "John", result.Name)
		assert.Equal(t, 30, result.Age)
	})

	t.Run("Invalid type", func(t *testing.T) {
		data := map[string][]string{
			"age": {"invalid"},
		}
		var result TestStruct
		err := x.BindData(&result, data)
		assert.Error(t, err)
	})

	t.Run("Non-pointer argument", func(t *testing.T) {
		data := map[string][]string{}
		var result TestStruct
		err := x.BindData(result, data)
		assert.Error(t, err)
	})
}

func TestJSONToURLValues(t *testing.T) {
	tests := []struct {
		name     string
		jsonStr  string
		expected url.Values
		wantErr  bool
	}{
		{
			name:     "Valid JSON with string values",
			jsonStr:  `{"key1": "value1", "key2": "value2"}`,
			expected: url.Values{"key1": []string{"value1"}, "key2": []string{"value2"}},
			wantErr:  false,
		},
		{
			name:     "Valid JSON with array values",
			jsonStr:  `{"key1": ["value1", "value2"], "key2": "value3"}`,
			expected: url.Values{"key1": []string{"value1", "value2"}, "key2": []string{"value3"}},
			wantErr:  false,
		},
		{
			name:     "Valid JSON with mixed types",
			jsonStr:  `{"key1": "value1", "key2": 42, "key3": true}`,
			expected: url.Values{"key1": []string{"value1"}, "key2": []string{"42"}, "key3": []string{"true"}},
			wantErr:  false,
		},
		{
			name:     "Invalid JSON",
			jsonStr:  `{"key1": "value1", "key2": }`,
			expected: nil,
			wantErr:  true,
		},
		{
			name:     "Empty JSON",
			jsonStr:  `{}`,
			expected: url.Values{},
			wantErr:  false,
		},
		{
			name:     "JSON with large float64 value",
			jsonStr:  `{"key": 1e20}`,
			expected: url.Values{"key": []string{"100000000000000000000"}},
			wantErr:  false,
		},
		{
			name:     "JSON with nested objects",
			jsonStr:  `{"key1": {"nested": "value"}, "key2": "value2"}`,
			expected: url.Values{"key1": []string{"map[nested:value]"}, "key2": []string{"value2"}},
			wantErr:  false,
		},
		{
			name:     "JSON with int64 value",
			jsonStr:  `{"key": "9223372036854775807"}`,
			expected: url.Values{"key": []string{"9223372036854775807"}},
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := x.JSONToURLValues(tt.jsonStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("JSONToURLValues() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("JSONToURLValues() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMapToSlice(t *testing.T) {
	t.Run("Direct conversion with struct", func(t *testing.T) {
		type Person struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		data := map[string]interface{}{
			"person1": map[string]interface{}{"name": "Alice", "age": 30},
			"person2": map[string]interface{}{"name": "Bob", "age": 25},
		}

		data2 := map[string]Person{
			"person1": {Name: "Alice", Age: 30},
			"person2": {Name: "Bob", Age: 25},
		}

		result, err := x.MapToSlice[string, interface{}, Person](data, nil)
		assert.NoError(t, err)
		assert.Len(t, result, 2)

		result2, err := x.MapToSlice(data2, func(k string, v Person) Person {
			return v
		})
		assert.NoError(t, err)
		assert.Len(t, result2, 2)

		expected := []Person{
			{Name: "Alice", Age: 30},
			{Name: "Bob", Age: 25},
		}
		assert.ElementsMatch(t, expected, result)
	})

	t.Run("Custom transformation", func(t *testing.T) {
		data := map[int]int64{1: 4, 2: 5, 3: 6}

		result, err := x.MapToSlice(data, func(k int, v int64) string {
			return fmt.Sprintf("%d_%d", k, v)
		})

		assert.NoError(t, err)
		assert.Len(t, result, 3)
		assert.ElementsMatch(t, []string{"1_4", "2_5", "3_6"}, result)
	})

	t.Run("Empty map", func(t *testing.T) {
		data := map[string]int{}

		result, err := x.MapToSlice(data, func(k string, v int) string {
			return fmt.Sprintf("%s:%d", k, v)
		})

		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Nil map", func(t *testing.T) {
		var data map[string]int

		result, err := x.MapToSlice(data, func(k string, v int) string {
			return fmt.Sprintf("%s:%d", k, v)
		})

		assert.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Map with interface{} values", func(t *testing.T) {
		data := map[string]interface{}{
			"a": 1,
			"b": "hello",
			"c": true,
		}

		result, err := x.MapToSlice(data, func(k string, v interface{}) string {
			return fmt.Sprintf("%s:%v", k, v)
		})

		assert.NoError(t, err)
		assert.Len(t, result, 3)
		assert.ElementsMatch(t, []string{"a:1", "b:hello", "c:true"}, result)
	})

	t.Run("Direct conversion with non-struct type", func(t *testing.T) {
		data := map[string]int{"a": 1, "b": 2, "c": 3}

		result, err := x.MapToSlice[string, int, int](data, nil)
		assert.NoError(t, err)
		assert.Len(t, result, 3)
		assert.ElementsMatch(t, []int{1, 2, 3}, result)
	})

	t.Run("Direct conversion with unmarshalable data", func(t *testing.T) {
		data := map[string]interface{}{
			"a": make(chan int), // channels are not JSON marshalable
		}

		_, err := x.MapToSlice[string, interface{}, interface{}](data, nil)
		assert.Error(t, err)
	})

	t.Run("Direct conversion with invalid JSON", func(t *testing.T) {
		type InvalidStruct struct {
			Field int `json:"field"`
		}

		data := map[string]interface{}{
			"a": map[string]interface{}{"field": "not an int"},
		}

		_, err := x.MapToSlice[string, interface{}, InvalidStruct](data, nil)
		assert.Error(t, err)
	})
}

func TestPtr(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		value := 42
		ptr := x.Ptr(value)
		assert.NotNil(t, ptr)
		assert.Equal(t, value, *ptr)
	})

	t.Run("string", func(t *testing.T) {
		value := "hello"
		ptr := x.Ptr(value)
		assert.NotNil(t, ptr)
		assert.Equal(t, value, *ptr)
	})

	t.Run("bool", func(t *testing.T) {
		value := true
		ptr := x.Ptr(value)
		assert.NotNil(t, ptr)
		assert.Equal(t, value, *ptr)
	})

	t.Run("struct", func(t *testing.T) {
		type TestStruct struct {
			Field int
		}
		value := TestStruct{Field: 10}
		ptr := x.Ptr(value)
		assert.NotNil(t, ptr)
		assert.Equal(t, value, *ptr)
	})
}

func TestDeref(t *testing.T) {
	t.Run("non-nil int pointer", func(t *testing.T) {
		value := 42
		ptr := &value
		result := x.Deref(ptr)
		assert.Equal(t, value, result)
	})

	t.Run("nil int pointer", func(t *testing.T) {
		var ptr *int
		result := x.Deref(ptr)
		assert.Equal(t, 0, result)
	})

	t.Run("non-nil string pointer", func(t *testing.T) {
		value := "hello"
		ptr := &value
		result := x.Deref(ptr)
		assert.Equal(t, value, result)
	})

	t.Run("nil string pointer", func(t *testing.T) {
		var ptr *string
		result := x.Deref(ptr)
		assert.Equal(t, "", result)
	})

	t.Run("non-nil bool pointer", func(t *testing.T) {
		value := true
		ptr := &value
		result := x.Deref(ptr)
		assert.Equal(t, value, result)
	})

	t.Run("nil bool pointer", func(t *testing.T) {
		var ptr *bool
		result := x.Deref(ptr)
		assert.Equal(t, false, result)
	})

	t.Run("non-nil struct pointer", func(t *testing.T) {
		type TestStruct struct {
			Field int
		}
		value := TestStruct{Field: 10}
		ptr := &value
		result := x.Deref(ptr)
		assert.Equal(t, value, result)
	})

	t.Run("nil struct pointer", func(t *testing.T) {
		type TestStruct struct {
			Field int
		}
		var ptr *TestStruct
		result := x.Deref(ptr)
		assert.Equal(t, TestStruct{}, result)
	})
}

func TestMapToStruct(t *testing.T) {
	t.Run("Basic struct conversion", func(t *testing.T) {
		type User struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		data := map[string]interface{}{
			"name": "Alice",
			"age":  30,
		}

		result, err := x.MapToStruct[User](data)
		assert.NoError(t, err)
		assert.Equal(t, User{Name: "Alice", Age: 30}, result)
	})

	t.Run("Nested struct conversion", func(t *testing.T) {
		type Address struct {
			Street string `json:"street"`
			City   string `json:"city"`
		}
		type User struct {
			Name    string  `json:"name"`
			Address Address `json:"address"`
		}

		data := map[string]interface{}{
			"name": "Bob",
			"address": map[string]interface{}{
				"street": "123 Main St",
				"city":   "Anytown",
			},
		}

		result, err := x.MapToStruct[User](data)
		assert.NoError(t, err)
		expected := User{
			Name: "Bob",
			Address: Address{
				Street: "123 Main St",
				City:   "Anytown",
			},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("Partial data", func(t *testing.T) {
		type User struct {
			Name    string `json:"name"`
			Age     int    `json:"age"`
			Country string `json:"country"`
		}

		data := map[string]interface{}{
			"name": "Charlie",
			"age":  25,
		}

		result, err := x.MapToStruct[User](data)
		assert.NoError(t, err)
		assert.Equal(t, User{Name: "Charlie", Age: 25}, result)
	})

	t.Run("Type mismatch", func(t *testing.T) {
		type User struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		data := map[string]interface{}{
			"name": "David",
			"age":  "thirty", // This should cause an error
		}

		_, err := x.MapToStruct[User](data)
		assert.Error(t, err)
	})

	t.Run("Empty map", func(t *testing.T) {
		type User struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		data := map[string]interface{}{}

		result, err := x.MapToStruct[User](data)
		assert.NoError(t, err)
		assert.Equal(t, User{}, result)
	})

	t.Run("Nil map", func(t *testing.T) {
		type User struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		var data map[string]interface{}

		result, err := x.MapToStruct[User](data)
		assert.NoError(t, err)
		assert.Equal(t, User{}, result)
	})
}

func TestStructToMap(t *testing.T) {
	t.Run("Basic struct conversion", func(t *testing.T) {
		type User struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		}

		user := User{Name: "Alice", Age: 30}

		result, err := x.StructToMap(user)
		assert.NoError(t, err)
		expected := map[string]interface{}{
			"name": "Alice",
			"age":  float64(30), // JSON numbers are floats
		}
		assert.Equal(t, expected, result)
	})

	t.Run("Nested struct conversion", func(t *testing.T) {
		type Address struct {
			Street string `json:"street"`
			City   string `json:"city"`
		}
		type User struct {
			Name    string  `json:"name"`
			Address Address `json:"address"`
		}

		user := User{
			Name: "Bob",
			Address: Address{
				Street: "123 Main St",
				City:   "Anytown",
			},
		}

		result, err := x.StructToMap(user)
		assert.NoError(t, err)
		expected := map[string]interface{}{
			"name": "Bob",
			"address": map[string]interface{}{
				"street": "123 Main St",
				"city":   "Anytown",
			},
		}
		assert.Equal(t, expected, result)
	})

	t.Run("Struct with unexported fields", func(t *testing.T) {
		type User struct {
			Name string `json:"name"`
			age  int    // unexported field
		}

		user := User{Name: "Charlie", age: 25}

		result, err := x.StructToMap(user)
		assert.NoError(t, err)
		expected := map[string]interface{}{
			"name": "Charlie",
		}
		assert.Equal(t, expected, result)
	})

	t.Run("Struct with custom types", func(t *testing.T) {
		type CustomInt int
		type User struct {
			Name string    `json:"name"`
			Age  CustomInt `json:"age"`
		}

		user := User{Name: "David", Age: CustomInt(40)}

		result, err := x.StructToMap(user)
		assert.NoError(t, err)
		expected := map[string]interface{}{
			"name": "David",
			"age":  float64(40), // JSON numbers are floats
		}
		assert.Equal(t, expected, result)
	})

	t.Run("Empty struct", func(t *testing.T) {
		type EmptyStruct struct{}

		empty := EmptyStruct{}

		result, err := x.StructToMap(empty)
		assert.NoError(t, err)
		assert.Empty(t, result)
	})

	t.Run("Non-struct input", func(t *testing.T) {
		input := 42

		_, err := x.StructToMap(input)
		assert.Error(t, err)
	})
}

func TestUniqueByKey(t *testing.T) {
	t.Run("Basic struct slice", func(t *testing.T) {
		type Person struct {
			ID   int
			Name string
		}
		input := []Person{
			{1, "Alice"},
			{2, "Bob"},
			{1, "Alice Clone"},
			{3, "Charlie"},
		}
		result := x.UniqueByKey(input, func(p Person) int { return p.ID })
		assert.Len(t, result, 3)
		assert.ElementsMatch(t, []Person{
			{1, "Alice"},
			{2, "Bob"},
			{3, "Charlie"},
		}, result)
	})

	t.Run("Empty slice", func(t *testing.T) {
		var input []string
		result := x.UniqueByKey(input, func(s string) string { return s })
		assert.Empty(t, result)
	})

	t.Run("Nil slice", func(t *testing.T) {
		var input []int
		result := x.UniqueByKey(input, func(i int) int { return i })
		assert.Nil(t, result)
	})

	t.Run("Complex key function", func(t *testing.T) {
		type User struct {
			FirstName string
			LastName  string
			Age       int
		}
		input := []User{
			{"John", "Doe", 30},
			{"Jane", "Doe", 25},
			{"John", "Smith", 35},
			{"John", "Doe", 40},
		}
		// Unique by full name
		result := x.UniqueByKey(input, func(u User) string {
			return u.FirstName + " " + u.LastName
		})
		assert.Len(t, result, 3)
		assert.ElementsMatch(t, []User{
			{"John", "Doe", 30},
			{"Jane", "Doe", 25},
			{"John", "Smith", 35},
		}, result)
	})

	t.Run("Composite key", func(t *testing.T) {
		type Event struct {
			Date     string
			Category string
			ID       int
		}
		input := []Event{
			{"2024-01-01", "A", 1},
			{"2024-01-01", "A", 2},
			{"2024-01-01", "B", 1},
			{"2024-01-02", "A", 1},
		}
		// Unique by date and category combination
		result := x.UniqueByKey(input, func(e Event) struct {
			date     string
			category string
		} {
			return struct {
				date     string
				category string
			}{e.Date, e.Category}
		})
		assert.Len(t, result, 3)
		assert.ElementsMatch(t, []Event{
			{"2024-01-01", "A", 1},
			{"2024-01-01", "B", 1},
			{"2024-01-02", "A", 1},
		}, result)
	})
}

func TestForEachMap(t *testing.T) {
	t.Run("Basic map iteration", func(t *testing.T) {
		m := map[string]int{
			"a": 1,
			"b": 2,
			"c": 3,
		}
		result := make(map[string]int)
		x.ForEachMap(m, func(k string, v int) {
			result[k] = v * 2
		})
		assert.Equal(t, map[string]int{
			"a": 2,
			"b": 4,
			"c": 6,
		}, result)
	})

	t.Run("Empty map", func(t *testing.T) {
		m := map[string]int{}
		count := 0
		x.ForEachMap(m, func(k string, v int) {
			count++
		})
		assert.Zero(t, count)
	})

	t.Run("Nil map", func(t *testing.T) {
		var m map[string]int
		count := 0
		x.ForEachMap(m, func(k string, v int) {
			count++
		})
		assert.Zero(t, count)
	})

	t.Run("Nil action", func(t *testing.T) {
		m := map[string]int{"a": 1}
		assert.NotPanics(t, func() {
			x.ForEachMap(m, nil)
		})
	})

	t.Run("Complex value type", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}
		m := map[int]Person{
			1: {Name: "Alice", Age: 30},
			2: {Name: "Bob", Age: 25},
		}
		names := make([]string, 0)
		x.ForEachMap(m, func(k int, v Person) {
			names = append(names, v.Name)
		})
		assert.ElementsMatch(t, []string{"Alice", "Bob"}, names)
	})
}

func TestForEachMapWithError(t *testing.T) {
	t.Run("Successful iteration", func(t *testing.T) {
		m := map[string]int{
			"a": 1,
			"b": 2,
			"c": 3,
		}
		result := make(map[string]int)
		err := x.ForEachMapWithError(m, func(k string, v int) error {
			result[k] = v * 2
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, map[string]int{
			"a": 2,
			"b": 4,
			"c": 6,
		}, result)
	})

	t.Run("Error during iteration", func(t *testing.T) {
		m := map[string]int{
			"a": 1,
			"b": -1,
			"c": 3,
		}
		result := make(map[string]int)
		err := x.ForEachMapWithError(m, func(k string, v int) error {
			if v < 0 {
				return fmt.Errorf("negative value found: %d", v)
			}
			result[k] = v * 2
			return nil
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "negative value found: -1")
		assert.Equal(t, map[string]int{
			"a": 2,
		}, result)
	})

	t.Run("Empty map", func(t *testing.T) {
		m := map[string]int{}
		count := 0
		err := x.ForEachMapWithError(m, func(k string, v int) error {
			count++
			return nil
		})
		assert.NoError(t, err)
		assert.Zero(t, count)
	})

	t.Run("Nil map", func(t *testing.T) {
		var m map[string]int
		count := 0
		err := x.ForEachMapWithError(m, func(k string, v int) error {
			count++
			return nil
		})
		assert.NoError(t, err)
		assert.Zero(t, count)
	})

	t.Run("Nil action", func(t *testing.T) {
		m := map[string]int{"a": 1}
		err := x.ForEachMapWithError(m, nil)
		assert.NoError(t, err)
	})

	t.Run("Complex value type with error", func(t *testing.T) {
		type Person struct {
			Name string
			Age  int
		}
		m := map[int]Person{
			1: {Name: "Alice", Age: 30},
			2: {Name: "Bob", Age: -1},
		}
		names := make([]string, 0)
		err := x.ForEachMapWithError(m, func(k int, v Person) error {
			if v.Age < 0 {
				return fmt.Errorf("invalid age %d for person %s", v.Age, v.Name)
			}
			names = append(names, v.Name)
			return nil
		})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid age -1 for person Bob")
		assert.Equal(t, []string{"Alice"}, names)
	})
}

func TestIsBlank(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "Single space",
			input:    " ",
			expected: true,
		},
		{
			name:     "Multiple spaces",
			input:    "   ",
			expected: true,
		},
		{
			name:     "Tabs and newlines",
			input:    "\t\n\r",
			expected: true,
		},
		{
			name:     "Mixed whitespace",
			input:    " \t \n \r ",
			expected: true,
		},
		{
			name:     "Non-blank string",
			input:    "hello",
			expected: false,
		},
		{
			name:     "String with spaces",
			input:    " hello ",
			expected: false,
		},
		{
			name:     "String with tabs",
			input:    "\thello\t",
			expected: false,
		},
		{
			name:     "Unicode whitespace",
			input:    "\u2000\u2001\u2002",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := x.IsBlank(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRetry(t *testing.T) {
	t.Run("Successful first attempt", func(t *testing.T) {
		attempts := 0
		err := x.Retry(func(info x.RetryInfo) error {
			attempts++
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 1, attempts)
	})

	t.Run("Successful after retries", func(t *testing.T) {
		attempts := 0
		err := x.Retry(func(info x.RetryInfo) error {
			attempts++
			if attempts < 3 {
				return errors.New("temporary error")
			}
			return nil
		})
		assert.NoError(t, err)
		assert.Equal(t, 3, attempts)
	})

	t.Run("Max attempts exceeded", func(t *testing.T) {
		attempts := 0
		err := x.Retry(func(info x.RetryInfo) error {
			attempts++
			return errors.New("persistent error")
		}, x.WithMaxAttempts(3))
		assert.Error(t, err)
		assert.Equal(t, 3, attempts)
		assert.Contains(t, err.Error(), "retry failed after 3 attempts")
	})

	t.Run("Context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		attempts := 0
		err := x.Retry(func(info x.RetryInfo) error {
			attempts++
			if attempts == 2 {
				cancel()
			}
			return errors.New("error")
		}, x.WithContext(ctx))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("Exponential backoff", func(t *testing.T) {
		attempts := 0
		start := time.Now()
		err := x.Retry(func(info x.RetryInfo) error {
			attempts++
			if attempts < 3 {
				return errors.New("error")
			}
			return nil
		},
			x.WithDelay(100*time.Millisecond),
			x.WithExponentialBackoff(2),
			x.WithMaxAttempts(3))
		duration := time.Since(start)
		assert.NoError(t, err)
		assert.Equal(t, 3, attempts)
		// First attempt immediate, second after 100ms, third after 200ms
		assert.True(t, duration >= 300*time.Millisecond)
	})

	t.Run("RetryIf condition", func(t *testing.T) {
		attempts := 0
		err := x.Retry(func(info x.RetryInfo) error {
			attempts++
			return errors.New("stop retry")
		}, x.WithRetryIf(func(err error) bool {
			return err.Error() != "stop retry"
		}))
		assert.Error(t, err)
		assert.Equal(t, 1, attempts)
		assert.Contains(t, err.Error(), "retry stopped by retryIf condition")
	})

	t.Run("OnRetry callback", func(t *testing.T) {
		var retryInfos []x.RetryInfo
		err := x.Retry(func(info x.RetryInfo) error {
			if info.Attempt < 3 {
				return errors.New("error")
			}
			return nil
		},
			x.WithMaxAttempts(3),
			x.WithOnRetry(func(info x.RetryInfo) {
				retryInfos = append(retryInfos, info)
			}))
		assert.NoError(t, err)
		assert.Len(t, retryInfos, 2) // Called before 2nd and 3rd attempts
		assert.Equal(t, 2, retryInfos[0].Attempt)
		assert.Equal(t, 3, retryInfos[1].Attempt)
	})
}

func TestRetryWithResult(t *testing.T) {
	t.Run("Successful first attempt", func(t *testing.T) {
		attempts := 0
		result, err := x.RetryWithResult(func(info x.RetryInfo) (string, error) {
			attempts++
			return "success", nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "success", result)
		assert.Equal(t, 1, attempts)
	})

	t.Run("Successful after retries", func(t *testing.T) {
		attempts := 0
		result, err := x.RetryWithResult(func(info x.RetryInfo) (string, error) {
			attempts++
			if attempts < 3 {
				return "", errors.New("temporary error")
			}
			return "success", nil
		})
		assert.NoError(t, err)
		assert.Equal(t, "success", result)
		assert.Equal(t, 3, attempts)
	})

	t.Run("Max attempts exceeded", func(t *testing.T) {
		attempts := 0
		result, err := x.RetryWithResult(func(info x.RetryInfo) (string, error) {
			attempts++
			return "failed", errors.New("persistent error")
		}, x.WithMaxAttempts(3))
		assert.Error(t, err)
		assert.Equal(t, "failed", result)
		assert.Equal(t, 3, attempts)
		assert.Contains(t, err.Error(), "retry failed after 3 attempts")
	})

	t.Run("Context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		attempts := 0
		result, err := x.RetryWithResult(func(info x.RetryInfo) (string, error) {
			attempts++
			if attempts == 2 {
				cancel()
			}
			return "cancelled", errors.New("error")
		}, x.WithContext(ctx))
		assert.Error(t, err)
		assert.Equal(t, "cancelled", result)
		assert.Contains(t, err.Error(), "context canceled")
	})

	t.Run("Exponential backoff", func(t *testing.T) {
		attempts := 0
		start := time.Now()
		result, err := x.RetryWithResult(func(info x.RetryInfo) (int, error) {
			attempts++
			if attempts < 3 {
				return 0, errors.New("error")
			}
			return 42, nil
		},
			x.WithDelay(100*time.Millisecond),
			x.WithExponentialBackoff(2),
			x.WithMaxAttempts(3))
		duration := time.Since(start)
		assert.NoError(t, err)
		assert.Equal(t, 42, result)
		assert.Equal(t, 3, attempts)
		// First attempt immediate, second after 100ms, third after 200ms
		assert.True(t, duration >= 300*time.Millisecond)
	})

	t.Run("RetryIf condition", func(t *testing.T) {
		attempts := 0
		result, err := x.RetryWithResult(func(info x.RetryInfo) (string, error) {
			attempts++
			return "stop", errors.New("stop retry")
		}, x.WithRetryIf(func(err error) bool {
			return err.Error() != "stop retry"
		}))
		assert.Error(t, err)
		assert.Equal(t, "stop", result)
		assert.Equal(t, 1, attempts)
		assert.Contains(t, err.Error(), "retry stopped by retryIf condition")
	})

	t.Run("Complex result type", func(t *testing.T) {
		type Response struct {
			Code    int
			Message string
		}

		attempts := 0
		result, err := x.RetryWithResult(func(info x.RetryInfo) (Response, error) {
			attempts++
			if attempts < 2 {
				return Response{Code: 500, Message: "Server Error"}, errors.New("server error")
			}
			return Response{Code: 200, Message: "OK"}, nil
		})
		assert.NoError(t, err)
		assert.Equal(t, Response{Code: 200, Message: "OK"}, result)
		assert.Equal(t, 2, attempts)
	})

	t.Run("With jitter", func(t *testing.T) {
		attempts := 0
		start := time.Now()
		result, err := x.RetryWithResult(func(info x.RetryInfo) (int, error) {
			attempts++
			if attempts < 3 {
				return 0, errors.New("error")
			}
			return 42, nil
		},
			x.WithDelay(100*time.Millisecond),
			x.WithJitter(50*time.Millisecond),
			x.WithMaxAttempts(3))
		duration := time.Since(start)
		assert.NoError(t, err)
		assert.Equal(t, 42, result)
		assert.Equal(t, 3, attempts)
		// With jitter, we can only verify that some delay occurred
		assert.True(t, duration >= 200*time.Millisecond)
	})

	t.Run("With onRetry callback", func(t *testing.T) {
		var retryInfos []x.RetryInfo
		result, err := x.RetryWithResult(func(info x.RetryInfo) (int, error) {
			if info.Attempt < 3 {
				return 0, errors.New("error")
			}
			return 42, nil
		},
			x.WithMaxAttempts(3),
			x.WithOnRetry(func(info x.RetryInfo) {
				retryInfos = append(retryInfos, info)
			}))
		assert.NoError(t, err)
		assert.Equal(t, 42, result)
		assert.Equal(t, 2, len(retryInfos)) // Should have 2 retries
		assert.Equal(t, 2, retryInfos[0].Attempt)
		assert.Equal(t, 3, retryInfos[1].Attempt)
	})
}

func TestTakeRight(t *testing.T) {
	t.Run("Normal case", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4, 5}
		result := x.TakeRight(numbers, 3)
		assert.Equal(t, []int{3, 4, 5}, result)
	})

	t.Run("Take more than length", func(t *testing.T) {
		numbers := []int{1, 2, 3}
		result := x.TakeRight(numbers, 5)
		assert.Equal(t, []int{1, 2, 3}, result)
	})

	t.Run("Take zero elements", func(t *testing.T) {
		numbers := []int{1, 2, 3}
		result := x.TakeRight(numbers, 0)
		assert.Nil(t, result)
	})

	t.Run("Take negative elements", func(t *testing.T) {
		numbers := []int{1, 2, 3}
		result := x.TakeRight(numbers, -1)
		assert.Nil(t, result)
	})

	t.Run("Nil slice", func(t *testing.T) {
		var numbers []int
		result := x.TakeRight(numbers, 3)
		assert.Nil(t, result)
	})
}

func TestDropRight(t *testing.T) {
	t.Run("Normal case", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4, 5}
		result := x.DropRight(numbers, 2)
		assert.Equal(t, []int{1, 2, 3}, result)
	})

	t.Run("Drop more than length", func(t *testing.T) {
		numbers := []int{1, 2, 3}
		result := x.DropRight(numbers, 5)
		assert.Equal(t, []int{}, result)
	})

	t.Run("Drop zero elements", func(t *testing.T) {
		numbers := []int{1, 2, 3}
		result := x.DropRight(numbers, 0)
		assert.Equal(t, []int{1, 2, 3}, result)
	})

	t.Run("Drop negative elements", func(t *testing.T) {
		numbers := []int{1, 2, 3}
		result := x.DropRight(numbers, -1)
		assert.Nil(t, result)
	})

	t.Run("Nil slice", func(t *testing.T) {
		var numbers []int
		result := x.DropRight(numbers, 3)
		assert.Nil(t, result)
	})
}

func TestPopFirst(t *testing.T) {
	t.Run("Normal case", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4}
		first, rest := x.PopFirst(numbers)
		assert.Equal(t, 1, first)
		assert.Equal(t, []int{2, 3, 4}, rest)
	})

	t.Run("Single element", func(t *testing.T) {
		numbers := []int{1}
		first, rest := x.PopFirst(numbers)
		assert.Equal(t, 1, first)
		assert.Nil(t, rest)
	})

	t.Run("Empty slice", func(t *testing.T) {
		var numbers []int
		first, rest := x.PopFirst(numbers)
		assert.Zero(t, first)
		assert.Nil(t, rest)
	})

	t.Run("Nil slice", func(t *testing.T) {
		var numbers []int
		first, rest := x.PopFirst(numbers)
		assert.Zero(t, first)
		assert.Nil(t, rest)
	})

	t.Run("String slice", func(t *testing.T) {
		strings := []string{"first", "second", "third"}
		first, rest := x.PopFirst(strings)
		assert.Equal(t, "first", first)
		assert.Equal(t, []string{"second", "third"}, rest)
	})
}

func TestPopLast(t *testing.T) {
	t.Run("Normal case", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4}
		last, rest := x.PopLast(numbers)
		assert.Equal(t, 4, last)
		assert.Equal(t, []int{1, 2, 3}, rest)
	})

	t.Run("Single element", func(t *testing.T) {
		numbers := []int{1}
		last, rest := x.PopLast(numbers)
		assert.Equal(t, 1, last)
		assert.Nil(t, rest)
	})

	t.Run("Empty slice", func(t *testing.T) {
		var numbers []int
		last, rest := x.PopLast(numbers)
		assert.Zero(t, last)
		assert.Nil(t, rest)
	})

	t.Run("Nil slice", func(t *testing.T) {
		var numbers []int
		last, rest := x.PopLast(numbers)
		assert.Zero(t, last)
		assert.Nil(t, rest)
	})

	t.Run("String slice", func(t *testing.T) {
		strings := []string{"first", "second", "third"}
		last, rest := x.PopLast(strings)
		assert.Equal(t, "third", last)
		assert.Equal(t, []string{"first", "second"}, rest)
	})
}

func TestHead(t *testing.T) {
	t.Run("Normal case", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4}
		result := x.Head(numbers)
		assert.Equal(t, []int{1, 2, 3}, result)
	})

	t.Run("Two elements", func(t *testing.T) {
		numbers := []int{1, 2}
		result := x.Head(numbers)
		assert.Equal(t, []int{1}, result)
	})

	t.Run("Single element", func(t *testing.T) {
		numbers := []int{1}
		result := x.Head(numbers)
		assert.Nil(t, result)
	})

	t.Run("Empty slice", func(t *testing.T) {
		var numbers []int
		result := x.Head(numbers)
		assert.Nil(t, result)
	})

	t.Run("Nil slice", func(t *testing.T) {
		var numbers []int
		result := x.Head(numbers)
		assert.Nil(t, result)
	})

	t.Run("String slice", func(t *testing.T) {
		strings := []string{"first", "second", "third"}
		result := x.Head(strings)
		assert.Equal(t, []string{"first", "second"}, result)
	})
}

func TestTail(t *testing.T) {
	t.Run("Normal case", func(t *testing.T) {
		numbers := []int{1, 2, 3, 4}
		result := x.Tail(numbers)
		assert.Equal(t, []int{2, 3, 4}, result)
	})

	t.Run("Two elements", func(t *testing.T) {
		numbers := []int{1, 2}
		result := x.Tail(numbers)
		assert.Equal(t, []int{2}, result)
	})

	t.Run("Single element", func(t *testing.T) {
		numbers := []int{1}
		result := x.Tail(numbers)
		assert.Nil(t, result)
	})

	t.Run("Empty slice", func(t *testing.T) {
		var numbers []int
		result := x.Tail(numbers)
		assert.Nil(t, result)
	})

	t.Run("Nil slice", func(t *testing.T) {
		var numbers []int
		result := x.Tail(numbers)
		assert.Nil(t, result)
	})

	t.Run("String slice", func(t *testing.T) {
		strings := []string{"first", "second", "third"}
		result := x.Tail(strings)
		assert.Equal(t, []string{"second", "third"}, result)
	})
}

// TestDifference tests the Difference function with various scenarios
func TestDifference(t *testing.T) {
	t.Run("Basic difference", func(t *testing.T) {
		slice1 := []int{1, 2, 3, 4}
		slice2 := []int{3, 4, 5, 6}
		left, right := x.Difference(slice1, slice2)
		assert.Equal(t, []int{1, 2}, left)
		assert.Equal(t, []int{5, 6}, right)
	})

	t.Run("Empty slices", func(t *testing.T) {
		left, right := x.Difference([]int{}, []int{})
		assert.Empty(t, left)
		assert.Empty(t, right)
	})

	t.Run("Nil slices", func(t *testing.T) {
		left, right := x.Difference[int](nil, nil)
		assert.Nil(t, left)
		assert.Nil(t, right)

		left, right = x.Difference([]int{1, 2}, nil)
		assert.Equal(t, []int{1, 2}, left)
		assert.Nil(t, right)

		left, right = x.Difference[int](nil, []int{1, 2})
		assert.Equal(t, []int{1, 2}, left)
		assert.Nil(t, right)
	})

	t.Run("No differences", func(t *testing.T) {
		slice1 := []int{1, 2, 3}
		slice2 := []int{1, 2, 3}
		left, right := x.Difference(slice1, slice2)
		assert.Empty(t, left)
		assert.Empty(t, right)
	})

	t.Run("String slices", func(t *testing.T) {
		slice1 := []string{"a", "b", "c"}
		slice2 := []string{"b", "c", "d"}
		left, right := x.Difference(slice1, slice2)
		assert.Equal(t, []string{"a"}, left)
		assert.Equal(t, []string{"d"}, right)
	})
}

// TestDifferenceBy tests the DifferenceBy function with various scenarios
func TestDifferenceBy(t *testing.T) {
	type Person struct {
		ID   int
		Name string
	}

	t.Run("Basic difference by ID", func(t *testing.T) {
		slice1 := []Person{{1, "Alice"}, {2, "Bob"}, {3, "Charlie"}}
		slice2 := []Person{{2, "Bob"}, {3, "Charlie"}, {4, "David"}}
		left, right := x.DifferenceBy(slice1, slice2, func(p Person) int { return p.ID })
		assert.Equal(t, []Person{{1, "Alice"}}, left)
		assert.Equal(t, []Person{{4, "David"}}, right)
	})

	t.Run("Empty slices", func(t *testing.T) {
		left, right := x.DifferenceBy([]Person{}, []Person{}, func(p Person) int { return p.ID })
		assert.Empty(t, left)
		assert.Empty(t, right)
	})

	t.Run("Nil slices", func(t *testing.T) {
		left, right := x.DifferenceBy[Person, int](nil, nil, func(p Person) int { return p.ID })
		assert.Nil(t, left)
		assert.Nil(t, right)

		slice := []Person{{1, "Alice"}}
		left, right = x.DifferenceBy(slice, nil, func(p Person) int { return p.ID })
		assert.Equal(t, slice, left)
		assert.Nil(t, right)

		left, right = x.DifferenceBy[Person, int](nil, slice, func(p Person) int { return p.ID })
		assert.Nil(t, left)
		assert.Equal(t, slice, right)
	})

	t.Run("No differences", func(t *testing.T) {
		slice1 := []Person{{1, "Alice"}, {2, "Bob"}}
		slice2 := []Person{{1, "Alice"}, {2, "Bob"}}
		left, right := x.DifferenceBy(slice1, slice2, func(p Person) int { return p.ID })
		assert.Empty(t, left)
		assert.Empty(t, right)
	})

	t.Run("Different compare functions", func(t *testing.T) {
		slice1 := []Person{{1, "Alice"}, {2, "Bob"}, {3, "Charlie"}}
		slice2 := []Person{{1, "Alice"}, {2, "Bobby"}, {4, "David"}}

		// Compare by ID
		left, right := x.DifferenceBy(slice1, slice2, func(p Person) int { return p.ID })
		assert.Equal(t, []Person{{3, "Charlie"}}, left)
		assert.Equal(t, []Person{{4, "David"}}, right)

		// Compare by Name
		left, right = x.DifferenceBy(slice1, slice2, func(p Person) string { return p.Name })
		assert.Equal(t, []Person{{2, "Bob"}, {3, "Charlie"}}, left)
		assert.Equal(t, []Person{{2, "Bobby"}, {4, "David"}}, right)
	})

	t.Run("Complex compare function", func(t *testing.T) {
		type Item struct {
			ID    int
			Value float64
		}
		slice1 := []Item{{1, 1.5}, {2, 2.5}, {3, 3.5}}
		slice2 := []Item{{1, 1.6}, {2, 2.5}, {4, 4.5}}

		// Compare by rounded Value
		left, right := x.DifferenceBy(slice1, slice2, func(i Item) int {
			return int(math.Round(i.Value))
		})
		assert.Equal(t, []Item{{3, 3.5}}, left)
		assert.Equal(t, []Item{{4, 4.5}}, right)
	})
}
