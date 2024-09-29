package x_test

import (
	"context"
	"errors"
	"fmt"
	"os"
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
	counter := 0
	f := func() {
		counter++
	}

	debounced := x.Debounce(f, 50*time.Millisecond)

	for i := 0; i < 10; i++ {
		debounced()
	}

	time.Sleep(100 * time.Millisecond)

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

	assert.True(t, counter == 2 || counter == 3)
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
		counter := 0
		for i := 0; i < 5; i++ {
			wg.Go(func() error {
				counter++
				return nil
			})
		}
		err := wg.Wait()
		assert.NoError(t, err)
		// assert.Equal(t, 5, counter)
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

func TestStringToBool(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"1", true},
		{"t", true},
		{"true", true},
		{"yes", true},
		{"y", true},
		{"on", true},
		{"0", false},
		{"f", false},
		{"false", false},
		{"no", false},
		{"n", false},
		{"off", false},
		{"", false},
		{"random", false},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := x.StringToBool(test.input)
			assert.Equal(t, test.expected, result)
		})
	}
}

func TestStringToInt(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		hasError bool
	}{
		{"42", 42, false},
		{"-42", -42, false},
		{"0", 0, false},
		{"", 0, true},
		{"abc", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := x.StringToInt(test.input)
			if test.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestStringToInt64(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
		hasError bool
	}{
		{"9223372036854775807", 9223372036854775807, false},
		{"-9223372036854775808", -9223372036854775808, false},
		{"0", 0, false},
		{"", 0, true},
		{"abc", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := x.StringToInt64(test.input)
			if test.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestStringToUint(t *testing.T) {
	tests := []struct {
		input    string
		expected uint
		hasError bool
	}{
		{"42", 42, false},
		{"0", 0, false},
		{"", 0, true},
		{"-1", 0, true},
		{"abc", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := x.StringToUint(test.input)
			if test.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestStringToUint64(t *testing.T) {
	tests := []struct {
		input    string
		expected uint64
		hasError bool
	}{
		{"18446744073709551615", 18446744073709551615, false},
		{"0", 0, false},
		{"", 0, true},
		{"-1", 0, true},
		{"abc", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := x.StringToUint64(test.input)
			if test.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestStringToFloat64(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
		hasError bool
	}{
		{"3.14", 3.14, false},
		{"-2.5", -2.5, false},
		{"0", 0, false},
		{"", 0, true},
		{"abc", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := x.StringToFloat64(test.input)
			if test.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestStringToDuration(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		hasError bool
	}{
		{"5s", 5 * time.Second, false},
		{"10m", 10 * time.Minute, false},
		{"2h30m", 2*time.Hour + 30*time.Minute, false},
		{"", 0, true},
		{"invalid", 0, true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result, err := x.StringToDuration(test.input)
			if test.hasError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.expected, result)
			}
		})
	}
}

func TestStringToMap(t *testing.T) {
	tests := []struct {
		input    string
		pairSep  string
		kvSep    string
		expected map[string]string
	}{
		{"key1=value1,key2=value2", ",", "=", map[string]string{"key1": "value1", "key2": "value2"}},
		{"k1:v1;k2:v2", ";", ":", map[string]string{"k1": "v1", "k2": "v2"}},
		{"", ",", "=", map[string]string{}},
		{"invalid", ",", "=", map[string]string{}},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			result := x.StringToMap(test.input, test.pairSep, test.kvSep)
			assert.Equal(t, test.expected, result)
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
