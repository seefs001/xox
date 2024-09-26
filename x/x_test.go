package x_test

import (
	"context"
	"errors"
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

	t.Run("Unbounded pool", func(t *testing.T) {
		pool := x.NewSafePool(0)
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
		assert.True(t, timedOut)
		assert.NoError(t, err)
	})

	t.Run("Wait with timeout - success", func(t *testing.T) {
		wg := x.NewWaitGroup()
		wg.Go(func() error {
			time.Sleep(50 * time.Millisecond)
			return nil
		})
		err, timedOut := wg.WaitWithTimeout(100 * time.Millisecond)
		assert.False(t, timedOut)
		assert.NoError(t, err)
	})
}
