---
name: chrono-mongo
description: MongoDB design patterns for ChronoAI apps. Use when designing database schemas, implementing user authentication with login tracking, or building analytics queries. Covers schema design, indexing strategies, connection pooling, and common patterns for user management.
---

# ChronoAI MongoDB Patterns

## Quick Start

User schema with login tracking:

```javascript
const userSchema = {
  _id: ObjectId,
  email: String,
  name: String,
  passwordHash: String,
  createdAt: Date,
  updatedAt: Date,
  lastLoginTime: Date,    // Last successful login
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
  { $set: { lastLoginTime: new Date() } }
);
```

## Index Strategy

- Email lookup: `db.users.createIndex({ email: 1 }, { unique: true })`
- Login queries: `db.users.createIndex({ lastLoginTime: 1 })`

## Schema Design Principles

1. **Embed** for 1-few relationships (user -> preferences)
2. **Reference** for 1-many relationships (user -> orders)
3. **Denormalize** for read-heavy data

## Reference Docs

- **User Schema:** [user-schema.md](references/user-schema.md)
- **Authentication:** [authentication.md](references/authentication.md)
- **Common Queries:** [queries.md](references/queries.md)
- **Index Strategies:** [index-strategies.md](references/index-strategies.md)
