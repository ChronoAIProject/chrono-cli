# User Schema Reference

## Complete Mongoose User Schema

```javascript
import mongoose from 'mongoose';

const userSchema = new mongoose.Schema({
  email: {
    type: String,
    required: true,
    unique: true,
    lowercase: true,
    trim: true,
    match: /^[^\s@]+@[^\s@]+\.[^\s@]+$/
  },
  name: {
    type: String,
    required: true,
    trim: true,
    maxlength: 100
  },
  passwordHash: {
    type: String,
    required: true,
    select: false  // Exclude by default from queries
  },
  // Activity tracking fields
  lastLoginTime: {
    type: Date,
    default: null
  },
  lastActiveTime: {
    type: Date,
    default: null,
    index: true  // For querying inactive users
  },
  // Status fields
  isActive: {
    type: Boolean,
    default: true,
    index: true
  },
  isVerified: {
    type: Boolean,
    default: false
  },
  // Embedded 1-few: preferences
  preferences: {
    notifications: {
      email: { type: Boolean, default: true },
      push: { type: Boolean, default: true }
    },
    theme: { type: String, enum: ['light', 'dark', 'auto'], default: 'auto' },
    language: { type: String, default: 'en' }
  }
}, {
  timestamps: true  // Adds createdAt, updatedAt
});

// Indexes
userSchema.index({ email: 1 }, { unique: true });
userSchema.index({ isActive: 1, lastActiveTime: 1 });
userSchema.index({ createdAt: -1 });

module.exports = mongoose.model('User', userSchema);
```

## Field Descriptions

| Field | Type | Purpose |
|-------|------|---------|
| `_id` | ObjectId | Auto-generated primary key |
| `email` | String | User email (unique, indexed) |
| `name` | String | Display name |
| `passwordHash` | String | Bcrypt hash (select: false) |
| `lastLoginTime` | Date | Last successful authentication |
| `lastActiveTime` | Date | Last API request/activity |
| `isActive` | Boolean | Account active status |
| `isVerified` | Boolean | Email verification status |
| `preferences` | Object | Embedded user settings |
| `createdAt` | Date | Account creation timestamp |
| `updatedAt` | Date | Last modification timestamp |

## TypeScript Interface

```typescript
interface IUser {
  _id: mongoose.Types.ObjectId;
  email: string;
  name: string;
  passwordHash: string;
  lastLoginTime: Date | null;
  lastActiveTime: Date | null;
  isActive: boolean;
  isVerified: boolean;
  preferences: {
    notifications: {
      email: boolean;
      push: boolean;
    };
    theme: 'light' | 'dark' | 'auto';
    language: string;
  };
  createdAt: Date;
  updatedAt: Date;
}
