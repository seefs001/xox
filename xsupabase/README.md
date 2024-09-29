# xsupabase

xsupabase is a Go client library for interacting with Supabase, providing a convenient interface for database operations, user management, storage handling, and more.

## Installation

```bash
go get github.com/seefspkg/xsupabase
```

## Usage

### Initializing the Client

```go
import "github.com/seefspkg/xsupabase"

client := xsupabase.NewClient("https://your-project-url.supabase.co", "your-api-key")

// Optional: Enable debug mode
client.SetDebug(true)
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
record := xsupabase.Record{
    "column1": "value1",
    "column2": 42,
}
insertedRecord, err := client.Insert(ctx, "your_table", record)
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

#### Upsert Records

```go
records := []xsupabase.Record{
    {"id": 1, "name": "John"},
    {"id": 2, "name": "Jane"},
}
upsertedRecords, err := client.Upsert(ctx, "your_table", records, "id")
```

#### Batch Operations

```go
operations := []map[string]interface{}{
    {"method": "INSERT", "data": xsupabase.Record{"name": "John"}},
    {"method": "UPDATE", "data": xsupabase.Record{"id": 1, "name": "Jane"}},
}
results, err := client.BatchOperation(ctx, "your_table", operations)
```

### User Management

#### Create User

```go
user, err := client.CreateUser(ctx, "user@example.com", "password123",
    xsupabase.WithUserMetadata(map[string]interface{}{"name": "John Doe"}),
    xsupabase.WithPhone("+1234567890"),
    xsupabase.WithEmailConfirmed(true),
)
```

#### Get User

```go
user, err := client.GetUser(ctx, "user_id")
```

#### Update User

```go
updatedUser, err := client.UpdateUser(ctx, "user_id",
    xsupabase.WithEmail("newemail@example.com"),
    xsupabase.WithUserMetadataUpdate(map[string]interface{}{"age": 30}),
)
```

#### Delete User

```go
err := client.DeleteUser(ctx, "user_id")
```

#### List Users

```go
users, err := client.ListUsers(ctx,
    xsupabase.WithPage(1),
    xsupabase.WithPerPage(20),
    xsupabase.WithUserMetadataFilter("role", "admin"),
)
```

### Storage Operations

#### Upload File

```go
file, _ := os.Open("path/to/file.jpg")
defer file.Close()
err := client.UploadFile(ctx, "bucket_name", "path/in/bucket/file.jpg", file)
```

#### Get Public URL

```go
url, err := client.GetStoragePublicURL(ctx, "bucket_name", "path/in/bucket/file.jpg")
```

#### Delete File

```go
err := client.DeleteFile(ctx, "bucket_name", "path/in/bucket/file.jpg")
```

#### List Files

```go
files, err := client.ListFiles(ctx, "bucket_name", "path/prefix")
```

#### Create Bucket

```go
err := client.CreateBucket(ctx, "new_bucket", true)
```

#### Delete Bucket

```go
err := client.DeleteBucket(ctx, "bucket_name")
```

#### Get Bucket Details

```go
details, err := client.GetBucketDetails(ctx, "bucket_name")
```

#### Update Bucket Details

```go
updates := map[string]interface{}{
    "public": false,
    "file_size_limit": 100 * 1024 * 1024,
}
err := client.UpdateBucketDetails(ctx, "bucket_name", updates)
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

#### Invite User by Email

```go
err := client.InviteUserByEmail(ctx, "user@example.com", "admin")
```

#### Get Project Settings

```go
settings, err := client.GetProjectSettings(ctx)
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

To enable debug logging:

```go
client.SetDebug(true)
```

This will log detailed information about API calls and responses using the `xlog` package.

## Asynchronous Operations

The library supports asynchronous operations for select, insert, update, and delete:

```go
task := client.AsyncSelect(ctx, "your_table", queryParams)
records, err := task.Wait()
```

Similar methods exist for `AsyncInsert`, `AsyncUpdate`, and `AsyncDelete`.

## Note on Real-time Subscriptions

The `SubscribeToChanges` method is currently a placeholder and not implemented. Real-time functionality may be added in future versions.
