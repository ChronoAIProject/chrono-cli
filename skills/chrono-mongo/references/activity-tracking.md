# Activity Tracking Patterns

## Update Scenarios

### 1. On Login (Always Update)
```javascript
await db.users.updateOne(
  { _id: userId },
  {
    $set: {
      lastLoginTime: new Date(),
      lastActiveTime: new Date()
    }
  }
);
```

### 2. On Activity (Optimistic - 5 Minute Throttle)
Reduces database writes by only updating if 5+ minutes have passed.

```javascript
const fiveMinutesAgo = new Date(Date.now() - 5 * 60 * 1000);

await db.users.updateOne(
  {
    _id: userId,
    $or: [
      { lastActiveTime: { $exists: false } },
      { lastActiveTime: { $lt: fiveMinutesAgo } }
    ]
  },
  { $set: { lastActiveTime: new Date() } }
);
```

### 3. Express Middleware (Fire and Forget)
```javascript
function activityMiddleware(req, res, next) {
  if (!req.user?.userId) return next();

  const fiveMinutesAgo = new Date(Date.now() - 5 * 60 * 1000);

  // Don't await - fire and forget
  db.users.updateOne(
    {
      _id: new mongoose.Types.ObjectId(req.user.userId),
      lastActiveTime: { $lt: fiveMinutesAgo }
    },
    { $set: { lastActiveTime: new Date() } }
  ).catch(() => {});

  next();
}
```

### 4. Using findOneAndUpdate with Projection
```javascript
const user = await db.users.findOneAndUpdate(
  {
    _id: userId,
    lastActiveTime: { $lt: fiveMinutesAgo }
  },
  { $set: { lastActiveTime: new Date() } },
  { returnDocument: 'after' }
);
```

## Activity Tracking Best Practices

1. **Throttle writes** - Don't update on every request
2. **Use fire-and-forget** for non-critical tracking
3. **Index `lastActiveTime`** for querying inactive users
4. **Consider TTL indexes** for auto-cleanup of stale data

## Throttle Intervals by Use Case

| Use Case | Recommended Interval |
|----------|---------------------|
| Real-time presence | 1-2 minutes |
| General activity | 5 minutes |
| Engagement metrics | 15-30 minutes |
| Daily active users | 1 hour |

## Express.js Complete Example

```javascript
import express from 'express';
import mongoose from 'mongoose';

const app = express();

// Activity tracking middleware
app.use('/api', (req, res, next) => {
  if (req.user?.userId) {
    const fiveMinutesAgo = new Date(Date.now() - 5 * 60 * 1000);
    mongoose.model('User').updateOne(
      { _id: req.user.userId, lastActiveTime: { $lt: fiveMinutesAgo } },
      { $set: { lastActiveTime: new Date() } }
    ).catch(() => {});
  }
  next();
});

// All /api routes will trigger activity updates for authenticated users
```
