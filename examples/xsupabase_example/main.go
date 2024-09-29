package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/seefs001/xox/x"
	"github.com/seefs001/xox/xenv"
	"github.com/seefs001/xox/xhttpc"
	"github.com/seefs001/xox/xlog"
	"github.com/seefs001/xox/xsupabase"
)

// User represents a user in our system
type User struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	xenv.Load()
	// Get Supabase configuration from environment variables
	supabaseURL := os.Getenv("SUPABASE_URL")
	if supabaseURL == "" {
		xlog.Error("SUPABASE_URL environment variable is not set")
		return
	}

	supabaseKey := os.Getenv("SUPABASE_API_KEY")
	if supabaseKey == "" {
		xlog.Error("SUPABASE_API_KEY environment variable is not set")
		return
	}

	// Initialize Supabase client
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

	// upload file
	// Open the file to upload
	// file, err := os.Open("./examples/data/example.txt")
	// if err != nil {
	// 	xlog.Error("Failed to open file", "error", err)
	// 	return
	// }
	// defer file.Close()

	// // Upload the file
	// err = client.UploadFile(ctx, "test", fmt.Sprintf("%s-%s", x.Must1(x.RandomString(8, x.ModeAlpha)), "txt"), file)
	// if err != nil {
	// 	xlog.Error("Failed to upload file", "error", err)
	// 	return
	// }
	// xlog.Info("File uploaded successfully")

	// // list file info
	// fileInfo, err := client.ListFiles(ctx, "test", "")
	// if err != nil {
	// 	xlog.Error("Failed to list file info", "error", err)
	// 	return
	// }
	// xlog.Info("File info retrieved successfully", "info", fileInfo)

	// return
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

		// Delete the user
		// err = client.Delete(ctx, "users", userID)
		// if err != nil {
		// 	xlog.Error("Failed to delete user", "error", err)
		// 	return
		// }
		// xlog.Info("User deleted successfully")
	} else {
		xlog.Info("No users found to update or delete")
	}

	// Count users
	count, err := client.Count(ctx, "users", "")
	if err != nil {
		xlog.Error("Failed to count users", "error", err)
		return
	}
	xlog.Info("User count", "count", count)

	// Execute a stored procedure
	// params := map[string]interface{}{
	// 	"name_pattern": "%John%",
	// }
	// rpcResp, err := client.ExecuteRPC(ctx, "get_users_by_name", params)
	// if err != nil {
	// 	xlog.Error("Failed to execute RPC", "error", err)
	// 	return
	// }
	// xlog.Info("RPC result", "response", string(rpcResp))
	// Create a new user in the auth system
	randomEmail := fmt.Sprintf("%s%d@gmail.com", x.Must1(x.RandomString(8, x.ModeAlpha)), x.Must1(x.RandomInt(100, 999)))
	newAuthUser, err := client.CreateUser(ctx, randomEmail, "password123", xsupabase.WithUserMetadata(map[string]interface{}{"name": "New User"}))
	if err != nil {
		if strings.Contains(err.Error(), "User not allowed") {
			xlog.Error("Failed to create new auth user: insufficient permissions. Make sure you're using an admin API key.", "error", err)
		} else {
			xlog.Error("Failed to create new auth user", "error", err)
		}
	} else {
		xlog.Info("Created new auth user", "user", newAuthUser)

		// Get user by ID
		fetchedUser, err := client.GetUser(ctx, newAuthUser.ID)
		if err != nil {
			xlog.Error("Failed to get user", "error", err)
			return
		}
		xlog.Info("Fetched user", "user", fetchedUser)

		// Update user
		updateOption := func(updates map[string]interface{}) {
			updates["user_metadata"] = map[string]interface{}{"name": "Updated New User"}
		}
		updatedUser, err := client.UpdateUser(ctx, newAuthUser.ID, updateOption)
		if err != nil {
			xlog.Error("Failed to update user", "error", err)
			return
		}
		xlog.Info("Updated auth user", "user", updatedUser)

		// List users
		users, err := client.ListUsers(ctx, xsupabase.WithPage(1), xsupabase.WithPerPage(10))
		if err != nil {
			xlog.Error("Failed to list users", "error", err)
			return
		}
		xlog.Info("Listed users", "count", len(users))

		// Delete user
		// err = client.DeleteUser(ctx, newAuthUser.ID)
		// if err != nil {
		// 	xlog.Error("Failed to delete user", "error", err)
		// 	return
		// }
		// xlog.Info("Deleted auth user successfully")
	}
}
