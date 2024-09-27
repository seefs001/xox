package main

import (
	"context"
	"os"
	"time"

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
		// Insert a new user to delete
		newUserToDelete := xsupabase.Record{
			"name":  "Jane Doe",
			"email": "jane@example.com",
		}
		insertRespToDelete, err := client.Insert(ctx, "users", newUserToDelete)
		if err != nil {
			xlog.Error("Failed to insert user to delete", "error", err)
			return
		}
		xlog.Info("Inserted user to delete", "response", insertRespToDelete)

		// Delete the newly inserted user
		newUserID := insertRespToDelete["id"]
		err = client.Delete(ctx, "users", newUserID)
		if err != nil {
			xlog.Error("Failed to delete new user", "error", err)
			return
		}
		xlog.Info("New user deleted successfully")

		// Delete the original user
		err = client.Delete(ctx, "users", userID)
		if err != nil {
			xlog.Error("Failed to delete original user", "error", err)
			return
		}
		xlog.Info("Original user deleted successfully")
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
	params := map[string]interface{}{
		"name_pattern": "%John%",
	}
	rpcResp, err := client.ExecuteRPC(ctx, "get_users_by_name", params)
	if err != nil {
		xlog.Error("Failed to execute RPC", "error", err)
		return
	}
	xlog.Info("RPC result", "response", string(rpcResp))
}
