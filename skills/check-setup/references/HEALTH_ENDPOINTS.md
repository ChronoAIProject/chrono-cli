# Health Endpoint Implementations

Your backend MUST implement a `/health` endpoint for Kubernetes readiness/liveness probes.

## Requirements

| Setting | Value |
|---------|-------|
| **Default path** | `/health` (configurable via `healthCheckPath`) |
| **Success codes** | HTTP 200-399 (2xx = healthy, 3xx = redirect) |
| **Failure codes** | HTTP 400-599 (4xx/5xx = not ready/unhealthy) |
| **Timeout** | 10 seconds per request |
| **Poll interval** | Every 10 seconds |

**Expected behavior:**
- Return quickly (< 5 seconds recommended, < 10 seconds required)
- Check critical dependencies (database, external services)
- Return HTTP 200 with simple response like `{"status": "ok"}`

## Implementations by Language

### Node.js

**Express:**
```javascript
app.get('/health', (req, res) => {
  res.status(200).json({ status: 'ok' });
});
```

**Fastify:**
```javascript
fastify.get('/health', async (request, reply) => {
  return { status: 'ok' };
});
```

### Go

**net/http:**
```go
http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"ok"}`))
})
```

**gin:**
```go
r.GET("/health", func(c *gin.Context) {
    c.JSON(http.StatusOK, gin.H{"status": "ok"})
})
```

### Python

**Flask:**
```python
@app.route('/health')
def health():
    return {'status': 'ok'}, 200
```

**FastAPI:**
```python
@app.get('/health')
def health():
    return {'status': 'ok'}
```

## Advanced: With Dependency Checks

Check database connection:
```python
@app.route('/health')
def health():
    try:
        db.session.execute(db.text('SELECT 1'))
        return {'status': 'ok', 'database': 'connected'}, 200
    except:
        return {'status': 'error', 'database': 'disconnected'}, 503
```
