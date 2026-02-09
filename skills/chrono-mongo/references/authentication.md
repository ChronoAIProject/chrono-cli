# Authentication Patterns

## Password Hashing with Bcrypt

```javascript
import bcrypt from 'bcrypt';

const SALT_ROUNDS = 12;

// Hash password on registration
async function hashPassword(password) {
  return await bcrypt.hash(password, SALT_ROUNDS);
}

// Verify password on login
async function verifyPassword(password, hash) {
  return await bcrypt.compare(password, hash);
}
```

## Registration Flow

```javascript
async function registerUser(email, name, password) {
  const existingUser = await db.users.findOne({ email });
  if (existingUser) {
    throw new Error('Email already registered');
  }

  const passwordHash = await hashPassword(password);

  const user = await db.users.insertOne({
    email,
    name,
    passwordHash,
    isVerified: false,
    // lastLoginTime null until first login
  });

  return user;
}
```

## Login Flow with Activity Tracking

```javascript
async function loginUser(email, password) {
  const user = await db.users.findOne({ email });
  if (!user) {
    throw new Error('Invalid credentials');
  }

  const isValid = await verifyPassword(password, user.passwordHash);
  if (!isValid) {
    throw new Error('Invalid credentials');
  }

  // Update login timestamp
  await db.users.updateOne(
    { _id: user._id },
    { $set: { lastLoginTime: new Date() } }
  );

  // Generate JWT token
  const token = generateJWT(user);

  return { user, token };
}
```

## JWT Token Generation

```javascript
import jwt from 'jsonwebtoken';

const JWT_SECRET = process.env.JWT_SECRET;
const JWT_EXPIRES_IN = '7d';

function generateJWT(user) {
  return jwt.sign(
    {
      userId: user._id.toString(),
      email: user.email
    },
    JWT_SECRET,
    { expiresIn: JWT_EXPIRES_IN }
  );
}

function verifyJWT(token) {
  try {
    return jwt.verify(token, JWT_SECRET);
  } catch (err) {
    throw new Error('Invalid token');
  }
}
```

## Session Middleware (Express)

```javascript
function authenticate(req, res, next) {
  const authHeader = req.headers.authorization;
  if (!authHeader?.startsWith('Bearer ')) {
    return res.status(401).json({ error: 'Unauthorized' });
  }

  const token = authHeader.substring(7);
  try {
    const payload = verifyJWT(token);
    req.user = payload;
    next();
  } catch (err) {
    res.status(401).json({ error: 'Invalid token' });
  }
}
