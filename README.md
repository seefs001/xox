# xo

A self-used Golang utility package with no third-party dependencies, inspired by various sources.

## Introduction

xo is a lightweight Golang utility package designed to provide common functionalities and practical tools without introducing any third-party dependencies. This project draws inspiration and best practices from various excellent open-source projects, carefully curated and optimized to meet everyday development needs.

## Features

- No third-party dependencies: All functionalities are implemented using the Golang standard library
- Lightweight: Includes only the most commonly used and practical utility functions
- Easy to use: Clean API design for seamless integration into existing projects
- Continuous improvement: Constantly refined and expanded based on real-world usage

## Usage

1. Clone the repository:

   ```
   git clone https://github.com/seefs001/xox.git
   ```

2. Import the package in your Go code:

   ```go
   import "github.com/seefs001/xox/x"
   ```

3. Use the provided utilities in your project as needed. Here are some examples:

   ```go
   // Must functions
   result := x.Must0(func() (int, error) {
       return 42, nil
   })

   // Ignore functions
   value := x.Ignore1(someFunction())

   // Slice operations
   filtered := x.Where([]int{1, 2, 3, 4, 5}, func(n int) bool {
       return n%2 == 0
   })

   mapped := x.Select([]int{1, 2, 3}, func(n int) string {
       return string(rune('A' + n - 1))
   })

   sum := x.Aggregate([]int{1, 2, 3, 4, 5}, 0, func(acc, n int) int {
       return acc + n
   })

   x.ForEach([]int{1, 2, 3}, func(n int) {
       fmt.Println(n)
   })

   numbers := x.Range(1, 5)

   evenCount := x.Count([]int{1, 2, 3, 4, 5}, func(n int) bool {
       return n%2 == 0
   })

   grouped := x.GroupBy([]int{1, 2, 3, 4, 5}, func(n int) string {
       if n%2 == 0 {
           return "even"
       }
       return "odd"
   })

   first, ok := x.First([]int{1, 2, 3, 4, 5}, func(n int) bool {
       return n%2 == 0
   })

   last, ok := x.Last([]int{1, 2, 3, 4, 5}, func(n int) bool {
       return n%2 == 0
   })

   hasEven := x.Any([]int{1, 2, 3, 4, 5}, func(n int) bool {
       return n%2 == 0
   })

   allEven := x.All([]int{2, 4, 6, 8}, func(n int) bool {
       return n%2 == 0
   })

   // Random string generation
   randomStr, err := x.RandomString(10, x.ModeAlphanumeric)

   // Tuple operations
   tuple := x.NewTuple(1, "two")

   // Parallel operations
   x.ParallelFor([]int{1, 2, 3, 4, 5}, func(n int) {
       fmt.Println(n * 2)
   })

   doubled := x.ParallelMap([]int{1, 2, 3, 4, 5}, func(n int) int {
       return n * 2
   })

   // Async operations
   task := x.NewAsyncTask(func() (int, error) {
       time.Sleep(time.Second)
       return 42, nil
   })
   result, err := task.Wait()

   // Debounce and Throttle
   debouncedFunc := x.Debounce(someFunction, 100*time.Millisecond)
   throttledFunc := x.Throttle(someFunction, 100*time.Millisecond)
   ```

## Contributing

Contributions are welcome! If you have any ideas, improvements, or bug fixes, please feel free to open an issue or submit a pull request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
