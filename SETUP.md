# Primkit Setup Guide

Step-by-step instructions for setting up primkit primitives with Cloudflare R2 replication.

---

## Prerequisites

- Go 1.26+ installed
- A Cloudflare account (free tier works)

## 1. Build

```bash
git clone git@github.com:propifly/primkit.git
cd primkit
make build
```

Binaries are placed in `./bin/`:
```
bin/taskprim
bin/stateprim
bin/knowledgeprim
bin/queueprim
```

## 2. Quick Start (No Replication)

```bash
# taskprim — add and list tasks
./bin/taskprim add "Review PR from Johanna" --list sprint-12
./bin/taskprim list

# stateprim — set and get state
./bin/stateprim set agents/johanna last_run '"2025-03-03T10:00:00Z"'
./bin/stateprim get agents/johanna last_run
```

Both create SQLite databases at `~/.taskprim/default.db` and `~/.stateprim/default.db` respectively.

```bash
# queueprim — enqueue and dequeue jobs
./bin/queueprim enqueue demo '{"task":"hello"}'
./bin/queueprim dequeue demo --worker local
```

Creates `~/.queueprim/default.db` on first use.

---

## 3. Cloudflare R2 Setup

R2 is S3-compatible object storage from Cloudflare. Free tier includes 10 GB storage and 10 million reads/month.

### 3a. Create an R2 Bucket

1. Log in to [Cloudflare Dashboard](https://dash.cloudflare.com)
2. In the left sidebar: **Storage & databases** → **R2 object storage** → **Overview**
3. Click **Create bucket**
4. Name it (e.g., `primkit-replication`)
5. Location: **Automatic** is fine
6. Click **Create bucket**

### 3b. Create R2 API Credentials

1. Go back to the **R2 overview page** (click "R2 object storage" → "Overview" in the left sidebar — **not** inside a bucket)
2. In the top-right area of the overview page, click **Manage R2 API Tokens**
3. Click **Create API token**
4. Token name: `primkit-replication`
5. Permissions: **Object Read & Write**
6. Specify bucket (optional): scope to your bucket for security
7. Click **Create API Token**
8. **Copy and save** the Access Key ID and Secret Access Key — they are shown only once

### 3c. Find Your Cloudflare Account ID

Your Account ID is in the dashboard URL:
```
https://dash.cloudflare.com/<ACCOUNT_ID>/r2/overview
```

Or: go to any domain → **Overview** → right sidebar shows **Account ID**.

---

## 4. Configure Replication

Create a config file (e.g., `config.yaml`):

```yaml
storage:
  db: ~/.taskprim/default.db
  replicate:
    enabled: true
    provider: r2
    bucket: primkit-replication
    path: taskprim.db
    endpoint: https://<YOUR_ACCOUNT_ID>.r2.cloudflarestorage.com
    access_key_id: ${R2_ACCESS_KEY_ID}
    secret_access_key: ${R2_SECRET_ACCESS_KEY}

server:
  port: 8090
```

Set the credentials as environment variables:

```bash
export R2_ACCESS_KEY_ID="your-access-key-id-here"
export R2_SECRET_ACCESS_KEY="your-secret-access-key-here"
```

> **Tip:** Add these to a `.env` file and source it, or use a secrets manager. Never commit credentials to git.

---

## 5. Test Replication

### 5a. Write some data with replication active

```bash
# Every CLI command now replicates when config has replication enabled
./bin/taskprim add "Test replication" --list demo --config config.yaml
./bin/taskprim add "Another task" --list demo --config config.yaml
./bin/taskprim list --config config.yaml
```

Each command: restores if DB missing → opens DB → starts Litestream → runs command → syncs → stops Litestream.

### 5b. Verify data is on R2

Check your R2 bucket in the Cloudflare dashboard. You should see files under the `taskprim.db/` prefix (Litestream creates LTX files and snapshots).

### 5c. Test restore from R2

```bash
# Delete the local database
rm -rf ~/.taskprim/

# Restore from R2
./bin/taskprim restore --config config.yaml

# Verify your tasks are back
./bin/taskprim list --config config.yaml
```

### 5d. Serve mode (continuous replication)

```bash
# Start the HTTP API — Litestream runs continuously in the background
./bin/taskprim serve --config config.yaml

# In another terminal, add tasks via API
curl -X POST http://localhost:8090/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{"what":"Created via API","list":"demo","source":"curl"}'

# Stop the server with Ctrl+C — final sync happens automatically
```

---

## 6. Stateprim Replication

Same setup, different DB path and port:

```yaml
storage:
  db: ~/.stateprim/default.db
  replicate:
    enabled: true
    provider: r2
    bucket: primkit-replication
    path: stateprim.db
    endpoint: https://<YOUR_ACCOUNT_ID>.r2.cloudflarestorage.com
    access_key_id: ${R2_ACCESS_KEY_ID}
    secret_access_key: ${R2_SECRET_ACCESS_KEY}

server:
  port: 8091
```

```bash
./bin/stateprim set agents/johanna last_run '"2025-03-03"' --config stateprim-config.yaml
./bin/stateprim restore --config stateprim-config.yaml
```

---

## 7. Environment Variable Overrides

Every config field can be overridden with env vars (prefixes: `TASKPRIM_`, `STATEPRIM_`, `KNOWLEDGEPRIM_`, `QUEUEPRIM_`):

| Env Var | Config Field |
|---------|-------------|
| `TASKPRIM_DB` | `storage.db` |
| `TASKPRIM_REPLICATE_ENABLED` | `storage.replicate.enabled` |
| `TASKPRIM_REPLICATE_PROVIDER` | `storage.replicate.provider` |
| `TASKPRIM_REPLICATE_BUCKET` | `storage.replicate.bucket` |
| `TASKPRIM_REPLICATE_PATH` | `storage.replicate.path` |
| `TASKPRIM_REPLICATE_ENDPOINT` | `storage.replicate.endpoint` |
| `TASKPRIM_REPLICATE_ACCESS_KEY_ID` | `storage.replicate.access_key_id` |
| `TASKPRIM_REPLICATE_SECRET_ACCESS_KEY` | `storage.replicate.secret_access_key` |
| `TASKPRIM_SERVER_PORT` | `server.port` |

This means you can skip the config file entirely:

```bash
export TASKPRIM_REPLICATE_ENABLED=true
export TASKPRIM_REPLICATE_PROVIDER=r2
export TASKPRIM_REPLICATE_BUCKET=primkit-replication
export TASKPRIM_REPLICATE_PATH=taskprim.db
export TASKPRIM_REPLICATE_ENDPOINT=https://<ACCOUNT_ID>.r2.cloudflarestorage.com
export TASKPRIM_REPLICATE_ACCESS_KEY_ID=your-key
export TASKPRIM_REPLICATE_SECRET_ACCESS_KEY=your-secret

# No --config needed — env vars are enough
./bin/taskprim add "Works without config file" --list demo
```

---

## 8. MCP Mode with Replication

For Claude Desktop or other MCP clients:

```json
{
  "mcpServers": {
    "taskprim": {
      "command": "/path/to/taskprim",
      "args": ["mcp", "--transport", "stdio", "--config", "/path/to/config.yaml"]
    }
  }
}
```

Replication runs automatically during the MCP session. When the client disconnects, a final sync pushes remaining changes to R2.

---

## Troubleshooting

**"replication is not enabled in the config"**
→ Check that `storage.replicate.enabled: true` is set in your config, or `TASKPRIM_REPLICATE_ENABLED=true` is exported.

**Restore says "no generation found, creating new database"**
→ No backup exists on R2 yet. This is normal on first run — Litestream needs to replicate at least once before restore works.

**"restoring from replica" errors with access denied**
→ Check your R2 credentials and that the bucket name matches exactly. Verify the endpoint URL includes your Cloudflare Account ID.

**Writes seem slow with replication enabled**
→ Litestream monitors the WAL file asynchronously. Writes to SQLite are not slowed down — replication happens in the background. The only added latency is the final sync on command exit.
