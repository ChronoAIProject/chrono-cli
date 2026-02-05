# Index Strategies

## Essential Indexes

### User Collection
```javascript
// Email lookup (unique, for login)
db.users.createIndex({ email: 1 }, { unique: true });

// Activity tracking queries
db.users.createIndex({ lastActiveTime: 1 });

// Active + compound queries
db.users.createIndex({ isActive: 1, lastActiveTime: 1 });

// Created_at sorting
db.users.createIndex({ createdAt: -1 });
```

## Compound Indexes

### For querying active users by activity
```javascript
// Supports: { isActive: true, lastActiveTime: { $gte: date } }
db.users.createIndex({ isActive: 1, lastActiveTime: 1 });
```

### For admin user search
```javascript
// Supports: { email: pattern }, sorting by createdAt
db.users.createIndex({ email: 1, createdAt: -1 });
```

## TTL Indexes (Auto-Expiration)

### Auto-delete soft-deleted users after 1 year
```javascript
// Requires a date field (e.g., deletedAt)
db.users.createIndex(
  { deletedAt: 1 },
  { expireAfterSeconds: 365 * 24 * 60 * 60 }
);
```

### Auto-cleanup old session data
```javascript
db.sessions.createIndex(
  { createdAt: 1 },
  { expireAfterSeconds: 7 * 24 * 60 * 60 }  // 7 days
);
```

## Partial Indexes (Smaller, Faster)

### Index only active users for activity queries
```javascript
db.users.createIndex(
  { lastActiveTime: 1 },
  { partialFilterExpression: { isActive: true } }
);
```

### Index only verified users
```javascript
db.users.createIndex(
  { email: 1 },
  {
    unique: true,
    partialFilterExpression: { isVerified: true }
  }
);
```

## Index Best Practices

1. **ESR Rule** - Equality, Sort, Range
   ```javascript
   // Good: E-S-R order
   db.users.find({ isActive: true, lastActiveTime: { $gte: date } })
     .sort({ createdAt: -1 })

   // Index: { isActive: 1, createdAt: -1, lastActiveTime: 1 }
   ```

2. **Covered Queries** - Index includes all queried fields
   ```javascript
   db.users.createIndex({ email: 1, name: 1, isActive: 1 });

   // Query only uses index, doesn't fetch documents
   db.users.find({ email: 'test@example.com' }, { _id: 0, name: 1, isActive: 1 })
   ```

3. **Avoid over-indexing** - Each index has write overhead

## Index Analysis

### Check index usage
```javascript
db.users.getIndexes();
```

### Explain query plan
```javascript
db.users.find({ email: 'test@example.com' }).explain('executionStats');
```

### Find unused indexes
```javascript
// Check with MongoDB Atlas or query $indexStats
db.users.aggregate([{ $indexStats: {} }]);
```
