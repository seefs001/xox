# xsupabase

xsupabase is a Go client library for interacting with Supabase, providing a convenient interface for database operations, user management, storage handling, and more.

## Installation

```bash
go get github.com/seefs001/xox/xsupabase
```

## Usage

### Initializing the Client

```go
import (
	"github.com/seefs001/xox/xsupabase"
	"github.com/seefs001/xox/xhttpc"
)

client := xsupabase.NewClient(
	"https://your-project-url.supabase.co",
	"your-api-key",
	xhttpc.WithDebug(true),
	xhttpc.WithLogOptions(xhttpc.LogOptions{
		LogHeaders:      true,
		LogBody:         true,
		LogResponse:     true,
		HeaderKeysToLog: []string{"Authorization", "apikey"},
		MaxBodyLogSize:  300,
	}),
	xhttpc.WithTimeout(15*time.Second),
)
```

### Database Operations

#### Select Records

```go
records, err := client.Select(ctx, "your_table", xsupabase.QueryParams{
	Select: "column1,column2",
	Order:  "column1.asc",
	Limit:  10,
	Offset: 0,
	Filter: "column1=eq.value",
})
```

#### Insert Record

```go
newRecord := xsupabase.Record{
	"column1": "value1",
	"column2": 42,
}
insertedRecord, err := client.Insert(ctx, "your_table", newRecord)
```

#### Update Record

```go
updates := xsupabase.Record{
	"column1": "new_value",
}
updatedRecord, err := client.Update(ctx, "your_table", recordID, updates)
```

#### Delete Record

```go
err := client.Delete(ctx, "your_table", recordID)
```

#### Count Records

```go
count, err := client.Count(ctx, "your_table", "column1=eq.value")
```

### User Management

#### Create User

```go
newUser, err := client.CreateUser(ctx, "user@example.com", "password123",
	xsupabase.WithUserMetadata(map[string]interface{}{"name": "John Doe"}),
)
```

#### Get User

```go
user, err := client.GetUser(ctx, "user_id")
```

#### Update User

```go
updatedUser, err := client.UpdateUser(ctx, "user_id", func(updates map[string]interface{}) {
	updates["user_metadata"] = map[string]interface{}{"name": "Updated Name"}
})
```

#### List Users

```go
users, err := client.ListUsers(ctx,
	xsupabase.WithPage(1),
	xsupabase.WithPerPage(10),
)
```

#### Delete User

```go
err := client.DeleteUser(ctx, "user_id")
```

### Storage Operations

#### Upload File

```go
file, _ := os.Open("path/to/file.txt")
defer file.Close()
err := client.UploadFile(ctx, "bucket_name", "path/in/bucket/file.txt", file)
```

#### List Files

```go
files, err := client.ListFiles(ctx, "bucket_name", "prefix")
```

#### Get Public URL

```go
url := client.GetStoragePublicURL(ctx, "bucket_name", "path/in/bucket/file.txt")
```

#### Delete File

```go
err := client.DeleteFile(ctx, "bucket_name", "path/in/bucket/file.txt")
```

### Other Operations

#### Execute RPC

```go
params := map[string]interface{}{
	"param1": "value1",
	"param2": 42,
}
result, err := client.ExecuteRPC(ctx, "function_name", params)
```

#### Execute GraphQL

```go
query := `
	query GetUser($id: UUID!) {
		user(id: $id) {
			id
			name
			email
		}
	}
`
variables := map[string]interface{}{
	"id": "user-uuid",
}
result, err := client.ExecuteGraphQL(ctx, query, variables)
```

## Error Handling

The library uses the `xerror` package for error handling. You can check for specific error types and codes:

```go
if err != nil {
	if xerror.IsCode(err, http.StatusNotFound) {
		// Handle 404 error
	} else {
		// Handle other errors
	}
}
```

## Debugging

Debugging is enabled through the `xhttpc.WithDebug` option when creating the client:

```go
client := xsupabase.NewClient(
	supabaseURL,
	supabaseKey,
	xhttpc.WithDebug(true),
	// ... other options
)
```

This will log detailed information about API calls and responses using the `xlog` package.

## Asynchronous Operations

The library supports asynchronous operations for select, insert, update, and delete:

```go
task := client.AsyncSelect(ctx, "your_table", queryParams)
records, err := task.Wait()
```

Similar methods exist for `AsyncInsert`, `AsyncUpdate`, and `AsyncDelete`.

## Example

Here's a comprehensive example demonstrating various features of xsupabase:

```go
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xenv"
	"github.com/seefs001/xox/xhttpc"
	"github.com/seefs001/xox/xlog"
	"github.com/seefs001/xox/xsupabase"
)

func main() {
	xenv.Load()
	supabaseURL := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_API_KEY")

	client := xsupabase.NewClient(
		supabaseURL,
		supabaseKey,
		xhttpc.WithDebug(true),
		xhttpc.WithLogOptions(xhttpc.LogOptions{
			LogHeaders:      true,
			LogBody:         true,
			LogResponse:     true,
			HeaderKeysToLog: []string{"Authorization", "apikey"},
			MaxBodyLogSize:  300,
		}),
		xhttpc.WithTimeout(15*time.Second),
	)

	ctx := context.Background()

	// Insert a new user
	newUser := xsupabase.Record{
		"name":  "John Doe",
		"email": "john@example.com",
	}
	insertResp, err := client.Insert(ctx, "users", newUser)
	if err != nil {
		xlog.Error("Failed to insert user", "error", err)
		return
	}
	xlog.Info("Inserted user", "response", insertResp)

	// Select users
	selectResp, err := client.Select(ctx, "users", xsupabase.QueryParams{})
	if err != nil {
		xlog.Error("Failed to select users", "error", err)
		return
	}
	xlog.Info("Selected users", "response", selectResp)

	// Update a user
	if len(selectResp) > 0 {
		userID := selectResp[0]["id"]
		updateData := xsupabase.Record{
			"name": "John Updated",
		}
		updateResp, err := client.Update(ctx, "users", userID, updateData)
		if err != nil {
			xlog.Error("Failed to update user", "error", err)
			return
		}
		xlog.Info("Updated user", "response", updateResp)
	}

	// Count users
	count, err := client.Count(ctx, "users", "")
	if err != nil {
		xlog.Error("Failed to count users", "error", err)
		return
	}
	xlog.Info("User count", "count", count)

	// Create a new user in the auth system
	randomEmail := fmt.Sprintf("%s%d@example.com", x.Must1(x.RandomString(8, x.ModeAlpha)), x.Must1(x.RandomInt(100, 999)))
	newAuthUser, err := client.CreateUser(ctx, randomEmail, "password123", xsupabase.WithUserMetadata(map[string]interface{}{"name": "New User"}))
	if err != nil {
		xlog.Error("Failed to create new auth user", "error", err)
		return
	}
	xlog.Info("Created new auth user", "user", newAuthUser)

	// List users
	users, err := client.ListUsers(ctx, xsupabase.WithPage(1), xsupabase.WithPerPage(10))
	if err != nil {
		xlog.Error("Failed to list users", "error", err)
		return
	}
	xlog.Info("Listed users", "count", len(users))
}
```

This example demonstrates database operations, user management, and error handling using the xsupabase package.

## Contributing

Contributions to xsupabase are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
