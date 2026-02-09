# Common Queries

## User Lookup Patterns

### Find by email with password hash (for login)
```javascript
const user = await db.users.findOne(
  { email: email.toLowerCase() },
  { projection: { passwordHash: 1 } }  // Only select needed fields
);
```

### Find user by ID
```javascript
const user = await db.users.findOne({
  _id: new mongoose.Types.ObjectId(userId)
});
```

### Find verified users
```javascript
const verifiedUsers = await db.users.find({
  isVerified: true
}).toArray();
```

## Login-Based Queries

### Users who never logged in
```javascript
const neverLoggedIn = await db.users.find({
  lastLoginTime: { $exists: false }
}).toArray();
```

### Recently logged in users (past 7 days)
```javascript
const sevenDaysAgo = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000);

const recentUsers = await db.users.find({
  lastLoginTime: { $gte: sevenDaysAgo }
}).toArray();
```

### Count users by login activity
```javascript
const thirtyDaysAgo = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000);

const stats = await db.users.aggregate([
  {
    $addFields: {
      loginCategory: {
        $cond: {
          if: { $gte: ['$lastLoginTime', thirtyDaysAgo] },
          then: 'active',
          else: {
            $cond: {
              if: { $gt: ['$lastLoginTime', null] },
              then: 'inactive',
              else: 'never'
            }
          }
        }
      }
    }
  },
  {
    $group: {
      _id: '$loginCategory',
      count: { $sum: 1 }
    }
  }
]).toArray();
```

## Cleanup and Retention Queries

### Hard-delete unverified users (30 days old, never logged in)
```javascript
const thirtyDaysAgo = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000);

const result = await db.users.deleteMany({
  isVerified: false,
  lastLoginTime: { $exists: false },
  createdAt: { $lt: thirtyDaysAgo }
});

console.log(`Deleted ${result.deletedCount} unverified users`);
```
