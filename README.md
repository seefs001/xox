> [!WARNING]
> This package is for personal use by seefs. Use at your own risk. The author takes no responsibility for any issues arising from its usage and will not provide support or bug fixes for external users.

> [!CAUTION]
> By using this package, you acknowledge that you are doing so at your own risk and without any expectation of support or maintenance from the author.

# xox 🧰

xox is a comprehensive Golang utility package that aggregates multiple sub-packages, each designed to provide specific functionalities without introducing third-party dependencies. Inspired by various sources, xox aims to streamline everyday development tasks with clean and efficient solutions.
## Features 🌟

- **No third-party dependencies**: All functionalities are implemented using the Golang standard library.
- **Lightweight**: Includes only the most commonly used and practical utility functions.
- **Easy to use**: Clean API design for seamless integration into existing projects.
- **Continuous improvement**: Constantly refined and expanded based on real-world usage.

## Packages 📦

| Package    | Description                                          | Status        | Example                                                                | Test File                                                           | Documentation |
|------------|------------------------------------------------------|---------------|------------------------------------------------------------------------|---------------------------------------------------------------------|---------------|
| [x](https://github.com/seefs001/xox/tree/master/x)          | Core utility methods for various workflows           | 🚧 Working     |                                                                        | [Test](https://github.com/seefs001/xox/blob/master/x/x_test.go)     | [Docs](https://github.com/seefs001/xox/tree/master/x) |
| [xai](https://github.com/seefs001/xox/tree/master/xai)        | AI-related functionalities                           | 🚧 Alpha       | [Example](https://github.com/seefs001/xox/tree/master/examples/xai_example) | 🚧 To be added | [Docs](https://github.com/seefs001/xox/tree/master/xai) |
| [xcast](https://github.com/seefs001/xox/tree/master/xcast)      | Type conversion utilities                            | 🚧 Alpha       |                                                                        | [Test](https://github.com/seefs001/xox/blob/master/xcast/xcast_test.go) | [Docs](https://github.com/seefs001/xox/tree/master/xcast) |
| [xcli](https://github.com/seefs001/xox/tree/master/xcli)       | CLI application building tools                       | 🚧 Alpha       | [Example](https://github.com/seefs001/xox/tree/master/examples/xcli_example) | 🚧 To be added | [Docs](https://github.com/seefs001/xox/tree/master/xcli) |
| [xcolor](https://github.com/seefs001/xox/tree/master/xcolor)     | Color-related utilities                              | 🚧 Alpha       |                                                                        | 🚧 To be added | [Docs](https://github.com/seefs001/xox/tree/master/xcolor) |
| [xconfig](https://github.com/seefs001/xox/tree/master/xconfig)    | Configuration management                             | 🚧 Alpha       |                                                                        | [Test](https://github.com/seefs001/xox/blob/master/xconfig/xconfig_test.go) | [Docs](https://github.com/seefs001/xox/tree/master/xconfig) |
| [xd](https://github.com/seefs001/xox/tree/master/xd)         | Dependency injection                                 | 🚧 Alpha       |                                                                        | [Test](https://github.com/seefs001/xox/blob/master/xd/xd_test.go)   | [Docs](https://github.com/seefs001/xox/tree/master/xd) |
| [xenv](https://github.com/seefs001/xox/tree/master/xenv)       | Environment variable handling                        | 🚧 Alpha       |                                                                        | 🚧 To be added | [Docs](https://github.com/seefs001/xox/tree/master/xenv) |
| [xerror](https://github.com/seefs001/xox/tree/master/xerror)     | Error handling and processing                        | ⚠️ Known Issues| [Example](https://github.com/seefs001/xox/tree/master/examples/xerror_example) | [Test](https://github.com/seefs001/xox/blob/master/xerror/xerror_test.go) | [Docs](https://github.com/seefs001/xox/tree/master/xerror) |
| [xhttp](https://github.com/seefs001/xox/tree/master/xhttp)      | HTTP server standard library helpers                 | 🚧 Alpha       | [Example](https://github.com/seefs001/xox/tree/master/examples/xhttp_example) | 🚧 To be added | [Docs](https://github.com/seefs001/xox/tree/master/xhttp) |
| [xhttpc](https://github.com/seefs001/xox/tree/master/xhttpc)     | HTTP client standard library helpers                 | 🚧 Alpha       |                                                                        | [Test](https://github.com/seefs001/xox/blob/master/xhttpc/xhttpc_test.go) | [Docs](https://github.com/seefs001/xox/tree/master/xhttpc) |
| [xjson](https://github.com/seefs001/xox/tree/master/xjson)      | JSON path data retrieval                             | ⚠️ Known Issues|                                                                        | [Test](https://github.com/seefs001/xox/blob/master/xjson/xjson_test.go) | [Docs](https://github.com/seefs001/xox/tree/master/xjson) |
| [xlog](https://github.com/seefs001/xox/tree/master/xlog)       | Logging utilities with various handlers              | 🚧 Alpha       | [Example](https://github.com/seefs001/xox/tree/master/examples/xlog_example) | [Test](https://github.com/seefs001/xox/blob/master/xlog/xlog_test.go) | [Docs](https://github.com/seefs001/xox/tree/master/xlog) |
| [xmw](https://github.com/seefs001/xox/tree/master/xmw)        | Middleware for standard HTTP servers                 | 🚧 Alpha       | [Example](https://github.com/seefs001/xox/tree/master/examples/xmw_example) | [Test](https://github.com/seefs001/xox/blob/master/xmw/xmw_test.go) | [Docs](https://github.com/seefs001/xox/tree/master/xmw) |
| [xsb](https://github.com/seefs001/xox/tree/master/xsb)        | SQL builder for database interactions                | 🚧 Working     |                                                                        | [Test](https://github.com/seefs001/xox/blob/master/xsb/xsb_test.go) | [Docs](https://github.com/seefs001/xox/tree/master/xsb) |
| [xsched](https://github.com/seefs001/xox/tree/master/xsched)     | Task scheduling utilities                            | 🚧 Working     | [Example](https://github.com/seefs001/xox/tree/master/examples/xsched_example) | [Test](https://github.com/seefs001/xox/blob/master/xsched/xsched_test.go) | [Docs](https://github.com/seefs001/xox/tree/master/xsched) |
| [xsupabase](https://github.com/seefs001/xox/tree/master/xsupabase)  | Supabase integration                                 | 🚧 Alpha       | [Example](https://github.com/seefs001/xox/tree/master/examples/xsupabase_example) | [Test](https://github.com/seefs001/xox/blob/master/xsupabase/xsupabase_test.go) | [Docs](https://github.com/seefs001/xox/tree/master/xsupabase) |
| [xtelebot](https://github.com/seefs001/xox/tree/master/xtelebot)   | Telegram bot API integration                         | 🚧 Alpha       | [Example](https://github.com/seefs001/xox/tree/master/examples/xtelebot_example) | 🚧 To be added | [Docs](https://github.com/seefs001/xox/tree/master/xtelebot) |
| [xvalidator](https://github.com/seefs001/xox/tree/master/xvalidator) | Data validation utilities                            | ⚠️ Known Issues|                                                                        | [Test](https://github.com/seefs001/xox/blob/master/xvalidator/xvalidator_test.go) | [Docs](https://github.com/seefs001/xox/tree/master/xvalidator) |
| [xtime](https://github.com/seefs001/xox/tree/master/xtime)      | Time handling utilities                              | 🚧 Alpha       |                                                                        | [Test](https://github.com/seefs001/xox/blob/master/xtime/xtime_test.go) | [Docs](https://github.com/seefs001/xox/tree/master/xtime) |
| [xresend](https://github.com/seefs001/xox/tree/master/xresend)  | Go client for Resend API                             | 🚧 Alpha       | [Example](https://github.com/seefs001/xox/tree/master/examples/xresend_example) |  | [Docs](https://github.com/seefs001/xox/tree/master/xresend) |

## Usage 🚀

To use xox and its sub-packages, follow these steps:

1. Install the package using go get:
   ```
   go get github.com/seefs001/xox
   ```

2. Import the desired sub-package in your Go code:
   ```go
   import "github.com/seefs001/xox/x"
   ```

3. Use the provided utilities as needed in your project.

## Examples 📚

Each sub-package comes with its own set of examples demonstrating how to use its functionalities. You can find the examples in the `examples/` directory of the repository.

## Testing 🧪

Each sub-package includes unit tests to ensure functionality and reliability. Tests can be found alongside their respective packages.

To run all tests, navigate to the root directory and execute:

```
go test ./...
```

## Contributing 🤝

While this package is primarily for personal use, contributions are welcome. Please ensure you follow the existing code style and include appropriate tests for any new functionality.

## License 📄

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments 🙏

- Inspired by various open-source projects and utility libraries in the Go ecosystem.
