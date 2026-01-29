# Rolling Restart Backend Deployment

Perform a Kubernetes rolling restart of the backend deployment. This restarts all pods with the existing image - useful for config changes or resetting pod state.

**Note:** This does NOT rebuild or redeploy new code. For new code, use `/deploy`.

## Step 1: Get Current Repo

```bash
git config --get remote.origin.url
git branch --show-current
```

Parse repo as owner/repo.

## Step 2: Find Existing Pipeline

Use `list_pipelines` to find the pipeline for this repo+branch.

**If not found:** Tell user "No pipeline found. Use /deploy first to create one."

## Step 3: Trigger Rolling Restart

Use `restart_deployment` with the pipelineId.

Tell user: "Rolling restart triggered. Waiting for pods to restart..."

## Step 4: Poll Until Restart Complete (REQUIRED)

**You MUST poll for status - do not skip this step.**

Call `get_deployment_status` with the pipelineId in a loop:

```
LOOP (max 24 iterations = 2 minutes):
  1. Call get_deployment_status(pipelineId)
  2. Check response:
     - If ready=true → DONE, go to Step 5 (success)
     - If failed=true → DONE, go to Step 5 (failure)
     - If rolloutInProgress=true OR ready=false → Continue polling
  3. Report progress: "Pods: {readyReplicas}/{replicas} ready, rollout in progress..."
  4. Wait 5 seconds
  5. Repeat
```

**Response fields to check:**
- `ready`: boolean - true when rollout is fully complete
- `rolloutInProgress`: boolean - true while pods are being updated
- `readyReplicas`: number of ready pods
- `replicas`: total desired pods
- `failed`: boolean - true if pod failure detected
- `failureReason`: string - error details if failed

## Step 5: Report Final Status

**On success (ready=true):**
```
✓ Rolling restart complete for {appName}
All pods healthy: {readyReplicas}/{replicas} ready
URLs: {deployedUrls}
```

**On failure (failed=true):**
```
✗ Restart failed: {failureReason}
Pod details: {pods}
```

**On timeout (24 polls without ready=true):**
```
⚠ Restart timeout - pods may still be cycling
Current: {readyReplicas}/{replicas} ready
Check pod status manually or try again
```

## When to Use

- Config/env changes picked up from secrets
- Reset application state
- Clear memory leaks
- Quick pod refresh

## When NOT to Use

- Deploying new code changes → use `/deploy` instead
- First-time deployment → use `/deploy` instead
