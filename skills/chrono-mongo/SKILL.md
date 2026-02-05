---
name: chrono-mongo
description: MongoDB design patterns for ChronoAI apps. Use when designing database schemas, implementing user authentication with login tracking, creating activity monitoring, or building analytics queries. Covers schema design, indexing strategies, connection pooling, and common patterns for user management.
---

# ChronoAI MongoDB Patterns

## Quick Start

User schema with activity tracking:

```javascript
const userSchema = {
  _id: ObjectId,
  email: String,
  name: String,
  passwordHash: String,
  createdAt: Date,
  updatedAt: Date,
  lastLoginTime: Date,    // Last successful login
  lastActiveTime: Date,   // Last activity (API request)
  isActive: Boolean,
  isVerified: Boolean,
};
```

## Environment Variables

| Variable | Value | Purpose |
|----------|-------|---------|
| `MONGODB_URI` | Connection string | Shared MongoDB cluster connection |
| `MONGODB_DATABASE` | `{appName}` | App-specific database name |

## Connection Setup (Mongoose)

```javascript
import mongoose from 'mongoose';

const conn = await mongoose.connect(process.env.MONGODB_URI, {
  dbName: process.env.MONGODB_DATABASE,
});
```

## Core Update Patterns

**On login:**
```javascript
await db.users.updateOne(
  { _id: userId },
  { $set: { lastLoginTime: new Date(), lastActiveTime: new Date() } }
);
```

**On activity (optimistic - only if >5 min):**
```javascript
await db.users.updateOne(
  { _id: userId, $or: [
    { lastActiveTime: { $exists: false } },
    { lastActiveTime: { $lt: new Date(Date.now() - 5 * 60 * 1000) } }
  ]},
  { $set: { lastActiveTime: new Date() } }
);
```

## Index Strategy

- Email lookup: `db.users.createIndex({ email: 1 }, { unique: true })`
- Activity queries: `db.users.createIndex({ lastActiveTime: 1 })`
- Compound: `db.users.createIndex({ isActive: 1, lastActiveTime: 1 })`

## Schema Design Principles

1. **Embed** for 1-few relationships (user -> preferences)
2. **Reference** for 1-many relationships (user -> orders)
3. **Denormalize** for read-heavy data

## Reference Docs

- **User Schema:** [user-schema.md](references/user-schema.md)
- **Authentication:** [authentication.md](references/authentication.md)
- **Activity Tracking:** [activity-tracking.md](references/activity-tracking.md)
- **Common Queries:** [queries.md](references/queries.md)
- **Index Strategies:** [index-strategies.md](references/index-strategies.md)
