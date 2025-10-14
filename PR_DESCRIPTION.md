# Fix: Implement cascade deletion for posts and related data

## 🐛 **Issue**
Fixes #1149 - Database information not purged after deleting posts from WebUI

**Problem**: When posts are deleted from the WebUI, related data in `attachments`, `comments`, `blobs`, and other tables remains in the database, causing:
- Database bloat over time
- Orphaned records consuming storage
- Potential performance degradation
- Data integrity issues

## 🔧 **Solution**

Implemented comprehensive cascade deletion that properly cleans up all related data when a post is deleted.

### **Core Changes**

1. **New DeletePost Command** (`app/models/cmd/post.go`)
   ```go
   type DeletePost struct {
       Post *entity.Post
   }
   ```

2. **Cascade Deletion Logic** (`app/services/sqlstore/postgres/post.go`)
   - Added `deletePost` function with complete cascade deletion
   - Deletes all related data in proper order to maintain referential integrity

3. **Updated API Handler** (`app/handlers/apiv1/post.go`)
   - Modified to use new `cmd.DeletePost` instead of just setting status to `PostDeleted`
   - Maintains existing authorization and validation logic

4. **Service Registration** (`app/services/sqlstore/postgres/postgres.go`)
   - Registered the new `deletePost` handler

### **Database Tables Cleaned**

The implementation now properly deletes data from these tables:
- ✅ `post_votes` - User votes on posts
- ✅ `post_subscribers` - Post notification subscribers  
- ✅ `post_tags` - Tag assignments
- ✅ `comments` - Post comments
- ✅ `reactions` - Comment reactions
- ✅ `attachments` - File attachments
- ✅ `notifications` - Web notifications
- ✅ `mention_notifications` - Mention notifications
- ✅ `posts` - The post itself
- ✅ `blobs` - Blob storage entries

## 🧪 **Testing**

### **Unit Tests Added**

**Database Layer Tests** (`app/services/sqlstore/postgres/post_test.go`):
- `TestPostStorage_DeletePost` - Complete cascade deletion with all related data
- `TestPostStorage_DeletePostWithTags` - Tag relationship cleanup
- `TestPostStorage_DeletePostWithMentionNotifications` - Mention cleanup

**API Handler Tests** (`app/handlers/apiv1/post_test.go`):
- `TestDeletePostHandler` - Successful deletion
- `TestDeletePostHandler_Unauthorized` - Authorization check
- `TestDeletePostHandler_PostNotFound` - Error handling
- `TestDeletePostHandler_PostReferenced` - Validation check

### **How to Test**

1. **Run Unit Tests**:
   ```bash
   go test ./app/services/sqlstore/postgres -run TestPostStorage_DeletePost -v
   go test ./app/handlers/apiv1 -run TestDeletePostHandler -v
   ```

2. **Run All Tests**:
   ```bash
   make test-server
   ```

3. **Manual Testing**:
   - Create a post with attachments, comments, votes, and tags
   - Delete the post via WebUI (admin only)
   - Verify all related data is removed from database

## 🔒 **Safety Features**

- **Transaction Safety**: All deletions happen within a single database transaction
- **Tenant Isolation**: All operations are scoped to the current tenant
- **Authorization**: Maintains existing admin-only deletion permissions
- **Validation**: Preserves existing validation (e.g., can't delete referenced posts)
- **Blob Cleanup**: Handles blob deletion gracefully with error handling

## 📈 **Benefits**

- **Database Growth Prevention**: Eliminates orphaned data that was causing database bloat
- **Storage Cleanup**: Removes unused blob storage entries
- **Data Integrity**: Maintains referential integrity through proper cascade deletion
- **Performance**: Reduces database size and improves query performance
- **Compliance**: Ensures complete data removal when posts are deleted

## 🔄 **Backward Compatibility**

- ✅ Maintains existing API endpoints
- ✅ Preserves all authorization and validation logic
- ✅ No breaking changes to existing functionality
- ✅ Existing notification system continues to work

## 📊 **Files Changed**

- `app/models/cmd/post.go` - Added DeletePost command
- `app/services/sqlstore/postgres/post.go` - Implemented cascade deletion logic
- `app/services/sqlstore/postgres/postgres.go` - Registered deletePost handler
- `app/handlers/apiv1/post.go` - Updated API handler
- `app/services/sqlstore/postgres/post_test.go` - Added database tests
- `app/handlers/apiv1/post_test.go` - Added API handler tests

## ✅ **Verification**

The fix can be verified by:

1. **Before Fix**: Create a post with attachments, delete it, check database for orphaned records
2. **After Fix**: Create a post with attachments, delete it, verify all related data is cleaned up
3. **Run Tests**: Execute the comprehensive test suite to ensure all scenarios work

This implementation completely resolves Issue #1149 by ensuring that when posts are deleted from the WebUI, all associated data is properly purged from the database.
