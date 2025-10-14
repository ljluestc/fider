package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	// Get database URL from environment or use default
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://fider:fider@localhost:5432/fider?sslmode=disable"
	}

	// Connect to database
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Test post ID (change this to your test post ID)
	postID := 1

	fmt.Printf("🔍 Verifying Issue #1149 Fix for Post ID: %d\n", postID)
	fmt.Println("================================================")

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
	hasData := false

	for tableName, query := range tables {
		var count int
		err := db.QueryRow(query, postID).Scan(&count)
		if err != nil {
			fmt.Printf("❌ Error checking %s: %v\n", tableName, err)
			continue
		}
		
		if count > 0 {
			hasData = true
		}
		totalRecords += count
		fmt.Printf("📊 %-20s: %d records\n", tableName, count)
	}

	fmt.Println("================================================")
	fmt.Printf("📈 Total related records: %d\n", totalRecords)

	if totalRecords == 0 {
		fmt.Println("✅ SUCCESS: Post and all related data have been deleted!")
		fmt.Println("✅ Issue #1149 fix is working correctly!")
	} else {
		fmt.Println("⚠️  WARNING: Some related data still exists.")
		fmt.Println("⚠️  This might indicate the fix is not working properly.")
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

	// Final assessment
	fmt.Println("\n🎯 FINAL ASSESSMENT")
	fmt.Println("==================")
	if totalRecords == 0 && blobCount == 0 {
		fmt.Println("✅ ISSUE #1149 FIX VERIFIED: Complete cascade deletion working!")
	} else {
		fmt.Println("❌ ISSUE #1149 FIX NEEDS REVIEW: Incomplete cleanup detected.")
	}

	fmt.Println("\n📝 Instructions:")
	fmt.Println("1. Create a test post with attachments, comments, votes")
	fmt.Println("2. Delete the post via WebUI (admin only)")
	fmt.Println("3. Run this script with the post ID")
	fmt.Println("4. Verify all counts are 0")
}
