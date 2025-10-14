# Testing and Verification Guide for Issue #1149 Fix

## 🎯 **Objective**
Verify that the cascade deletion fix properly cleans up all related data when posts are deleted from the WebUI.

## 🧪 **Automated Testing**

### **1. Run Unit Tests**

```bash
# Database layer tests
go test ./app/services/sqlstore/postgres -run TestPostStorage_DeletePost -v

# API handler tests  
go test ./app/handlers/apiv1 -run TestDeletePostHandler -v

# Run all tests
make test-server
```

### **2. Expected Test Results**

All tests should pass with output similar to:
```
=== RUN   TestPostStorage_DeletePost
--- PASS: TestPostStorage_DeletePost (0.05s)
=== RUN   TestPostStorage_DeletePostWithTags
--- PASS: TestPostStorage_DeletePostWithTags (0.03s)
=== RUN   TestPostStorage_DeletePostWithMentionNotifications
--- PASS: TestPostStorage_DeletePostWithMentionNotifications (0.02s)
=== RUN   TestDeletePostHandler
--- PASS: TestDeletePostHandler (0.01s)
=== RUN   TestDeletePostHandler_Unauthorized
--- PASS: TestDeletePostHandler_Unauthorized (0.01s)
=== RUN   TestDeletePostHandler_PostNotFound
--- PASS: TestDeletePostHandler_PostNotFound (0.01s)
=== RUN   TestDeletePostHandler_PostReferenced
--- PASS: TestDeletePostHandler_PostReferenced (0.01s)
```

## 🔍 **Manual Testing**

### **Step 1: Setup Test Data**

1. **Start Fider application**
2. **Login as administrator**
3. **Create a test post with comprehensive data**:
   - Post title: "Test Post for Deletion"
   - Post description: "This post will be deleted to test cascade cleanup"
   - Add 2-3 file attachments
   - Add 2-3 comments
   - Add reactions to comments
   - Assign 1-2 tags
   - Vote on the post

### **Step 2: Verify Data Exists Before Deletion**

Check database for related data (replace `POST_ID` with actual post ID):

```sql
-- Check all related tables
SELECT 'post_votes' as table_name, COUNT(*) as count FROM post_votes WHERE post_id = POST_ID
UNION ALL
SELECT 'post_subscribers', COUNT(*) FROM post_subscribers WHERE post_id = POST_ID
UNION ALL
SELECT 'comments', COUNT(*) FROM comments WHERE post_id = POST_ID
UNION ALL
SELECT 'reactions', COUNT(*) FROM reactions WHERE comment_id IN (SELECT id FROM comments WHERE post_id = POST_ID)
UNION ALL
SELECT 'attachments', COUNT(*) FROM attachments WHERE post_id = POST_ID
UNION ALL
SELECT 'notifications', COUNT(*) FROM notifications WHERE post_id = POST_ID
UNION ALL
SELECT 'mention_notifications', COUNT(*) FROM mention_notifications WHERE post_id = POST_ID
UNION ALL
SELECT 'posts', COUNT(*) FROM posts WHERE id = POST_ID;
```

**Expected Result**: Should show records in multiple tables (votes, comments, attachments, etc.)

### **Step 3: Delete Post via WebUI**

1. **Navigate to the test post**
2. **Click the delete button** (admin only)
3. **Enter deletion reason** (optional)
4. **Confirm deletion**

### **Step 4: Verify Complete Cleanup**

Run the same SQL query from Step 2.

**Expected Result**: All counts should be 0
```
post_votes: 0
post_subscribers: 0
comments: 0
reactions: 0
attachments: 0
notifications: 0
mention_notifications: 0
posts: 0
```

### **Step 5: Verify Blob Cleanup**

Check if attachment blobs were cleaned up:

```sql
-- Check for orphaned blobs
SELECT COUNT(*) as orphaned_blobs FROM blobs 
WHERE key IN (
    SELECT DISTINCT attachment_bkey 
    FROM attachments 
    WHERE post_id = POST_ID
);
```

**Expected Result**: `orphaned_blobs: 0`

## 🚨 **Error Scenarios Testing**

### **1. Unauthorized Deletion**
- Login as regular user (non-admin)
- Try to delete a post
- **Expected**: 403 Forbidden error

### **2. Post Not Found**
- Try to delete non-existent post
- **Expected**: 404 Not Found error

### **3. Referenced Post**
- Create a post that references another post (duplicate)
- Try to delete the referenced post
- **Expected**: Validation error preventing deletion

## 📊 **Performance Verification**

### **Database Size Impact**

```sql
-- Check database size before and after
SELECT 
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables 
WHERE schemaname = 'public' 
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

**Expected**: Reduced size in `attachments`, `comments`, `blobs` tables after cleanup

## 🔧 **Debugging**

### **Check Transaction Logs**

```sql
-- Enable query logging to see deletion sequence
SET log_statement = 'all';
SET log_min_duration_statement = 0;
```

### **Verify Foreign Key Constraints**

```sql
-- Check if foreign key constraints are working
SELECT 
    tc.table_name, 
    kcu.column_name, 
    ccu.table_name AS foreign_table_name,
    ccu.column_name AS foreign_column_name 
FROM 
    information_schema.table_constraints AS tc 
    JOIN information_schema.key_column_usage AS kcu
      ON tc.constraint_name = kcu.constraint_name
      AND tc.table_schema = kcu.table_schema
    JOIN information_schema.constraint_column_usage AS ccu
      ON ccu.constraint_name = tc.constraint_name
      AND ccu.table_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY' 
AND tc.table_name IN ('post_votes', 'comments', 'attachments', 'reactions');
```

## ✅ **Success Criteria**

The fix is working correctly if:

1. **All unit tests pass** ✅
2. **Manual deletion removes all related data** ✅
3. **No orphaned records remain in database** ✅
4. **Blob storage is cleaned up** ✅
5. **Authorization and validation still work** ✅
6. **No performance regression** ✅
7. **Database size is reduced** ✅

## 🎯 **Verification Checklist**

- [ ] Unit tests pass
- [ ] Manual deletion test completed
- [ ] Database cleanup verified
- [ ] Blob cleanup verified
- [ ] Authorization tests pass
- [ ] Error handling tests pass
- [ ] Performance impact measured
- [ ] No regression in existing functionality

## 📝 **Test Report Template**

```
Issue #1149 Fix Verification Report
==================================

Test Date: [DATE]
Tester: [NAME]
Environment: [LOCAL/STAGING/PRODUCTION]

Test Results:
- Unit Tests: ✅/❌
- Manual Deletion: ✅/❌
- Database Cleanup: ✅/❌
- Blob Cleanup: ✅/❌
- Authorization: ✅/❌
- Error Handling: ✅/❌
- Performance: ✅/❌

Issues Found:
- [List any issues]

Recommendation:
- [APPROVED/NEEDS WORK]
```

This comprehensive testing ensures that Issue #1149 is completely resolved and the database cleanup works as expected.
