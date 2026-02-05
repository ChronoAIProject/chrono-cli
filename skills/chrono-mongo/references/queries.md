# Common Queries

## Find Inactive Users

### Users inactive for 30+ days
```javascript
const thirtyDaysAgo = new Date(Date.now() - 30 * 24 * 60 * 60 * 1000);

const inactiveUsers = await db.users.find({
  isActive: true,
  $or: [
    { lastActiveTime: { $lt: thirtyDaysAgo } },
    { lastActiveTime: { $exists: false } }
  ]
}).toArray();
```

### Users who never logged in
```javascript
const neverLoggedIn = await db.users.find({
  lastLoginTime: { $exists: false }
}).toArray();
```

### Recently active users (past 24 hours)
```javascript
const oneDayAgo = new Date(Date.now() - 24 * 60 * 60 * 1000);

const activeUsers = await db.users.find({
  lastActiveTime: { $gte: oneDayAgo }
}).toArray();
```

## Aggregation Pipeline Examples

### Daily Active Users (DAU) - Last 30 Days
```javascript
const dau = await db.users.aggregate([
  {
    $match: {
      lastActiveTime: { $gte: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000) }
    }
  },
  {
    $group: {
      _id: {
        $dateToString: { format: '%Y-%m-%d', date: '$lastActiveTime' }
      },
      count: { $sum: 1 }
    }
  },
  { $sort: { _id: 1 } }
]).toArray();
```

### Weekly Active Users (WAU)
```javascript
const wau = await db.users.aggregate([
  {
    $match: {
      lastActiveTime: { $gte: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000) }
    }
  },
  {
    $group: {
      _id: {
        year: { $year: '$lastActiveTime' },
        week: { $week: '$lastActiveTime' }
      },
      count: { $sum: 1 }
    }
  }
]).toArray();
```

### Monthly Active Users (MAU)
```javascript
const mau = await db.users.aggregate([
  {
    $match: {
      lastActiveTime: { $gte: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000) }
    }
  },
  {
    $group: {
      _id: {
        year: { $year: '$lastActiveTime' },
        month: { $month: '$lastActiveTime' }
      },
      count: { $sum: 1 }
    }
  }
]).toArray();
```

### Engagement Segmentation
```javascript
const segments = await db.users.aggregate([
  {
    $addFields: {
      daysSinceActive: {
        $divide: [
          { $subtract: [new Date(), '$lastActiveTime'] },
          1000 * 60 * 60 * 24
        ]
      }
    }
  },
  {
    $bucket: {
      groupBy: '$daysSinceActive',
      boundaries: [0, 7, 30, 90, Infinity],
      default: 'Never',
      output: {
        count: { $sum: 1 },
        users: { $push: '$email' }
      }
    }
  }
]).toArray();

// Result format:
// [{ _id: 0, count: 45 },     // Active (0-7 days)
//  { _id: 7, count: 120 },    // Cooling (7-30 days)
//  { _id: 30, count: 200 },   // Inactive (30-90 days)
//  { _id: 90, count: 350 }]   // Churned (90+ days)
```

## Cleanup and Retention Queries

### Soft-delete stale users (90+ days inactive)
```javascript
const ninetyDaysAgo = new Date(Date.now() - 90 * 24 * 60 * 60 * 1000);

const result = await db.users.updateMany(
  {
    isActive: true,
    lastActiveTime: { $lt: ninetyDaysAgo }
  },
  { $set: { isActive: false } }
);

console.log(`Deactivated ${result.modifiedCount} stale users`);
```

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

## User Lookup Patterns

### Find by email with password hash (for login)
```javascript
const user = await db.users.findOne(
  { email: email.toLowerCase() },
  { projection: { passwordHash: 1 } }  // Only select needed fields
);
```

### Find active user by ID
```javascript
const user = await db.users.findOne({
  _id: new mongoose.Types.ObjectId(userId),
  isActive: true
});
```
