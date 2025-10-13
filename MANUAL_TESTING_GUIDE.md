# Manual Testing Guide for Post Deletion

## 🎯 **Objective**
Verify that when a post is deleted from the WebUI, all related data is properly removed from the database.

## 📋 **Prerequisites**
1. Fider instance running locally
2. Database access (PostgreSQL)
3. Admin user account
4. Test data with posts containing various related data

## 🔧 **Setup Test Data**

### Step 1: Create a Test Post with All Related Data
```sql
-- Create a test post
INSERT INTO posts (title, slug, number, description, tenant_id, user_id, created_at, status) 
VALUES ('Test Post for Deletion', 'test-post-for-deletion', 999, 'This post will be deleted', 1, 1, NOW(), 0);

-- Get the post ID (let's assume it's 999)
SET @post_id = 999;

-- Add votes
INSERT INTO post_votes (post_id, user_id, tenant_id, created_at) VALUES (@post_id, 1, 1, NOW());
INSERT INTO post_votes (post_id, user_id, tenant_id, created_at) VALUES (@post_id, 2, 1, NOW());

-- Add subscribers
INSERT INTO post_subscribers (post_id, user_id, tenant_id, created_at) VALUES (@post_id, 1, 1, NOW());
INSERT INTO post_subscribers (post_id, user_id, tenant_id, created_at) VALUES (@post_id, 2, 1, NOW());

-- Add comments
INSERT INTO comments (post_id, user_id, tenant_id, content, created_at) VALUES (@post_id, 1, 1, 'Test comment 1', NOW());
INSERT INTO comments (post_id, user_id, tenant_id, content, created_at) VALUES (@post_id, 2, 1, 'Test comment 2', NOW());

-- Get comment IDs
SET @comment1_id = (SELECT id FROM comments WHERE post_id = @post_id AND content = 'Test comment 1');
SET @comment2_id = (SELECT id FROM comments WHERE post_id = @post_id AND content = 'Test comment 2');

-- Add reactions to comments
INSERT INTO reactions (comment_id, user_id, emoji, created_on) VALUES (@comment1_id, 1, '👍', NOW());
INSERT INTO reactions (comment_id, user_id, emoji, created_on) VALUES (@comment2_id, 2, '❤️', NOW());

-- Add attachments
INSERT INTO attachments (post_id, user_id, tenant_id, attachment_bkey) VALUES (@post_id, 1, 1, 'test-attachment-1');
INSERT INTO attachments (post_id, user_id, tenant_id, attachment_bkey) VALUES (@post_id, 2, 1, 'test-attachment-2');

-- Add notifications
INSERT INTO notifications (post_id, user_id, tenant_id, title, link, read, author_id, created_on) 
VALUES (@post_id, 1, 1, 'Test notification', '', false, 2, NOW());

-- Add mention notifications
INSERT INTO mention_notifications (post_id, user_id, tenant_id, created_on) VALUES (@post_id, 1, 1, NOW());

-- Add blob entries
INSERT INTO blobs (key, tenant_id, size, content_type, file, created_at) 
VALUES ('test-attachment-1', 1, 1024, 'image/png', E'\\x89504E470D0A1A0A', NOW());
INSERT INTO blobs (key, tenant_id, size, content_type, file, created_at) 
VALUES ('test-attachment-2', 1, 2048, 'image/jpeg', E'\\xFFD8FFE0', NOW());
```

## 🧪 **Testing Steps**

### Step 1: Verify Initial Data
```sql
-- Check that all data exists before deletion
SELECT 'post_votes' as table_name, COUNT(*) as count FROM post_votes WHERE post_id = @post_id
UNION ALL
SELECT 'post_subscribers', COUNT(*) FROM post_subscribers WHERE post_id = @post_id
UNION ALL
SELECT 'comments', COUNT(*) FROM comments WHERE post_id = @post_id
UNION ALL
SELECT 'reactions', COUNT(*) FROM reactions WHERE comment_id IN (SELECT id FROM comments WHERE post_id = @post_id)
UNION ALL
SELECT 'attachments', COUNT(*) FROM attachments WHERE post_id = @post_id
UNION ALL
SELECT 'notifications', COUNT(*) FROM notifications WHERE post_id = @post_id
UNION ALL
SELECT 'mention_notifications', COUNT(*) FROM mention_notifications WHERE post_id = @post_id
UNION ALL
SELECT 'posts', COUNT(*) FROM posts WHERE id = @post_id
UNION ALL
SELECT 'blobs', COUNT(*) FROM blobs WHERE key IN ('test-attachment-1', 'test-attachment-2');
```

### Step 2: Delete Post via WebUI
1. Log in as an administrator
2. Navigate to the test post (number 999)
3. Click the delete button
4. Enter a deletion reason (optional)
5. Confirm deletion

### Step 3: Verify Complete Deletion
```sql
-- Check that all data is deleted after deletion
SELECT 'post_votes' as table_name, COUNT(*) as count FROM post_votes WHERE post_id = @post_id
UNION ALL
SELECT 'post_subscribers', COUNT(*) FROM post_subscribers WHERE post_id = @post_id
UNION ALL
SELECT 'comments', COUNT(*) FROM comments WHERE post_id = @post_id
UNION ALL
SELECT 'reactions', COUNT(*) FROM reactions WHERE comment_id IN (SELECT id FROM comments WHERE post_id = @post_id)
UNION ALL
SELECT 'attachments', COUNT(*) FROM attachments WHERE post_id = @post_id
UNION ALL
SELECT 'notifications', COUNT(*) FROM notifications WHERE post_id = @post_id
UNION ALL
SELECT 'mention_notifications', COUNT(*) FROM mention_notifications WHERE post_id = @post_id
UNION ALL
SELECT 'posts', COUNT(*) FROM posts WHERE id = @post_id
UNION ALL
SELECT 'blobs', COUNT(*) FROM blobs WHERE key IN ('test-attachment-1', 'test-attachment-2');
```

## ✅ **Expected Results**

### Before Deletion:
- post_votes: 2
- post_subscribers: 2
- comments: 2
- reactions: 2
- attachments: 2
- notifications: 1
- mention_notifications: 1
- posts: 1
- blobs: 2

### After Deletion:
- post_votes: 0
- post_subscribers: 0
- comments: 0
- reactions: 0
- attachments: 0
- notifications: 0
- mention_notifications: 0
- posts: 0
- blobs: 0

## 🚨 **Error Scenarios to Test**

### 1. Non-Admin User Attempting Deletion
- Log in as a regular user
- Try to delete a post
- Expected: 403 Forbidden error

### 2. Deleting Referenced Post
- Create a post that references another post (duplicate)
- Try to delete the referenced post
- Expected: Validation error preventing deletion

### 3. Post Not Found
- Try to delete a non-existent post
- Expected: 404 Not Found error

## 📊 **Performance Testing**

### Database Size Impact
```sql
-- Check database size before and after deletion
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables 
WHERE schemaname = 'public' 
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

## 🔍 **Debugging Tips**

### Check Transaction Logs
```sql
-- Enable query logging to see the deletion sequence
SET log_statement = 'all';
SET log_min_duration_statement = 0;
```

### Verify Blob Cleanup
```bash
# Check blob storage directory/file system
ls -la /path/to/blob/storage/tenants/1/
```

## 📝 **Test Report Template**

```
Post Deletion Test Report
========================

Test Date: [DATE]
Tester: [NAME]
Environment: [LOCAL/STAGING/PRODUCTION]

Test Results:
- Database cascade deletion: ✅/❌
- API handler functionality: ✅/❌
- Authorization checks: ✅/❌
- Error handling: ✅/❌
- Blob cleanup: ✅/❌
- Performance impact: ✅/❌

Issues Found:
- [List any issues]

Recommendations:
- [Any recommendations]
```
