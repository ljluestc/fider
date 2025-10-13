package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Get database URL from environment
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://fider:fider@localhost:5432/fider_test?sslmode=disable"
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test post ID (you can change this)
	postID := 999

	fmt.Printf("🔍 Verifying data for post ID: %d\n", postID)
	fmt.Println("=====================================")

	// Check all related tables
	tables := map[string]string{
		"post_votes":            "SELECT COUNT(*) FROM post_votes WHERE post_id = $1",
		"post_subscribers":      "SELECT COUNT(*) FROM post_subscribers WHERE post_id = $1",
		"comments":              "SELECT COUNT(*) FROM comments WHERE post_id = $1",
		"reactions":             "SELECT COUNT(*) FROM reactions WHERE comment_id IN (SELECT id FROM comments WHERE post_id = $1)",
		"attachments":           "SELECT COUNT(*) FROM attachments WHERE post_id = $1",
		"notifications":         "SELECT COUNT(*) FROM notifications WHERE post_id = $1",
		"mention_notifications": "SELECT COUNT(*) FROM mention_notifications WHERE post_id = $1",
		"posts":                 "SELECT COUNT(*) FROM posts WHERE id = $1",
	}

	totalRecords := 0
	for tableName, query := range tables {
		var count int
		err := db.QueryRow(query, postID).Scan(&count)
		if err != nil {
			fmt.Printf("❌ Error checking %s: %v\n", tableName, err)
			continue
		}
		
		fmt.Printf("📊 %-20s: %d records\n", tableName, count)
		totalRecords += count
	}

	fmt.Println("=====================================")
	fmt.Printf("📈 Total related records: %d\n", totalRecords)

	if totalRecords == 0 {
		fmt.Println("✅ Post and all related data have been successfully deleted!")
	} else {
		fmt.Println("⚠️  Some related data still exists. Deletion may be incomplete.")
	}

	// Check blob cleanup
	fmt.Println("\n🗂️  Checking blob cleanup...")
	blobQuery := `
		SELECT COUNT(*) FROM blobs 
		WHERE key IN (
			SELECT DISTINCT attachment_bkey 
			FROM attachments 
			WHERE post_id = $1
		)
	`
	var blobCount int
	err = db.QueryRow(blobQuery, postID).Scan(&blobCount)
	if err != nil {
		fmt.Printf("❌ Error checking blobs: %v\n", err)
	} else {
		fmt.Printf("📦 Orphaned blobs: %d\n", blobCount)
		if blobCount == 0 {
			fmt.Println("✅ Blob cleanup successful!")
		} else {
			fmt.Println("⚠️  Some blobs may not have been cleaned up.")
		}
	}
}
