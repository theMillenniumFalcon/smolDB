## Durability guarantees (WAL) — implementation and testing plan

### Overview
- WAL protects PUT/PATCH/DELETE by appending entries before mutating files.
- Durability modes: `none`, `commit` (fsync per append), `grouped` (time and/or batch-triggered fsync).
- Sync modes: `none`, `fsync`, `dsync` (best-effort; mapped to fsync on most platforms).
- Explicit `COMMIT` markers appended at each flush boundary for auditability.
- `sh.Setup` initializes WAL at startup and replays it to recover from crashes.
- `PATCH` now routes through `index.I.Put` to ensure WAL capture.

### Files changed
- `index/wal.go`: New WAL implementation (append, replay, durability levels, grouped fsync via time and batch triggers, ts/csum fields).
- `index/index.go`: Added WAL fields, InitWAL/InitWALWithOptions, WALAvailable, WALReplay; WAL appends in Put/Delete.
- `api/api.go`: `PatchKeyField` now persists via `index.I.Put`.
- `sh/shell.go`: WAL initialization and replay (`Setup`, `SetupWithOptions`), shell variant with options.

### Configuration
- Default durability if unspecified: `commit` (fsync per append).
- WAL storage location: `<dir>/.smoldb/wal.log`.
- CLI/env flags:
  - `--durability none|commit|grouped` (env: `SMOLDB_DURABILITY`)
  - `--group-commit-ms <N>` (env: `SMOLDB_GROUP_COMMIT_MS`) — used when `durability=grouped`.
  - `--group-commit-batch <N>` (env: `SMOLDB_GROUP_COMMIT_BATCH`) — fsync after N appends (grouped mode).

Example runs
```powershell
# server, grouped commits every 5ms, or at 100-entry batches (whichever first)
.\smoldb.exe --dir .\db --port 8080 --durability grouped --group-commit-ms 5 --group-commit-batch 100 start

# server, no fsync (throughput testing only)
.\smoldb.exe --dir .\db --port 8080 --durability none start

# shell, commit-per-append (default)
.\smoldb.exe --dir .\db shell
```

---

## How to test — manual and automated

### 1) Unit tests (deterministic)
1. WAL append/parse
   - Use in-memory FS (`afero.NewMemMapFs`).
   - Create index with dir `.`; init WAL `DurabilityCommit`.
   - Append `PUT` and `DELETE` entries; reopen WAL file; ensure lines are present and parseable.

2. Replay applies state
   - Start with empty FS and index; write a crafted `wal.log` containing:
     - `PUT k1` body `{"a":1}`
     - `PUT k2` body `{"b":2}`
     - `DELETE k1`
   - Run `WALReplay`; assert `k2.json` exists with expected content; `k1.json` absent; index contains only `k2`.

3. API paths write through WAL
   - For `PUT /key/k`: call handler with body; assert WAL contains a `PUT` for `k` and file contents updated.
   - For `PATCH /key/k/field/f`: call handler; assert WAL has `PUT` for `k` with updated merged content.
   - For `DELETE /key/k`: call handler; assert WAL has `DELETE` and file removed.

4. Malformed line tolerance
   - Preload WAL with a bad JSON line in between valid lines; `Replay` should skip/break gracefully and not panic; resulting state should reflect all valid entries up to the malformed line.

5. Concurrency safety
   - Run multiple `Put` operations across keys concurrently; ensure no race detector failures (`go test -race`).

### 2) Crash-recovery tests (local manual)
1. Happy-path recovery
   - Start server with `--dir ./db`.
   - `PUT /key/k1` with some body.
   - Kill the process forcefully (e.g., `kill -9` or stop container) immediately after.
   - Restart server pointing at same dir.
   - `GET /key/k1` should return the previously written body (WAL replay should recover).

2. Crash between WAL append and file write (simulate)
   - Temporarily instrument code or use a failpoint to exit after WAL append but before file write.
   - On restart, replay should apply the `PUT` to the file and index, recovering the write.

3. Crash during file write
   - Simulate by truncating file mid-write or killing process during a large `PUT`.
   - On restart, replay should re-apply `PUT` and correct the partial file.

4. Delete recovery
   - Write a key; force-crash after WAL `DELETE` append but before file removal.
   - On restart, replay should remove the file.

### 3) Durability behavior verification
1. fsync effectiveness
   - Use a filesystem that honors sync; run repeated `PUT` with `DurabilityCommit`.
   - Pull the plug (kill power on VM or ungraceful shutdown) and verify no acknowledged writes are lost.

2. Performance trade-off
   - Benchmark simple `PUT` throughput with `commit` vs `none` modes.

3. Grouped commit loss window
   - Time window: run with `--durability grouped --group-commit-ms 10 --sync-mode fsync`.
   - Rapidly issue `PUT`s; power-cut (kill -9 / docker kill -s KILL).
   - After restart: ensure at most ~10ms worth of latest acknowledged appends are missing.
   - Batch window: run with `--group-commit-batch 50` and no `--group-commit-ms`.
   - After power-cut: ensure at most 49 latest acknowledged appends are missing.

### 4) Integration tests (scripted)
- Create `./db` temp dir per test; run server on random port.
- Perform sequences: `PUT` x N, `PATCH` x M, `DELETE` subset.
- Hard kill server; restart; assert final state equals intended state.

### 5) Metrics and observability (future when metrics added)
- Assert counters for WAL appends and fsync timings are exposed; use them during load to validate behavior.

---

## Step-by-Step Testing Checklist (PowerShell)

Here's a concise, step-by-step checklist to test each durability feature using Windows PowerShell with `Invoke-RestMethod`.

### Prerequisites
Build and start the server (adjust flags per test):
```powershell
go build -o smoldb.exe
.\smoldb.exe --dir .\db --port 8080 --durability commit --sync-mode fsync start
```

After a crash test, delete lock before restart:
```powershell
del .\db\smoldb_lock
```

WAL path: `.\db\.smoldb\wal.log`

### Test 1: WAL append-before-mutate
Write a key:
```powershell
Invoke-RestMethod -Method Put -Uri 'http://localhost:8080/key/k1' -ContentType 'application/json' -Body '{"a":1}'
```

**Validate:**
- `wal.log` contains a PUT entry for `k1`
- `.\db\k1.json` equals `{"a":1}`

### Test 2: Startup replay (crash-recovery)
Write a key, then hard-kill the process:
```powershell
Invoke-RestMethod -Method Put -Uri 'http://localhost:8080/key/recover' -ContentType 'application/json' -Body '{"ok":true}'
Get-Process smoldb | Stop-Process -Force
del .\db\smoldb_lock
.\smoldb.exe --dir .\db --port 8080 start
```

**Verify:**
```powershell
Invoke-RestMethod -Uri 'http://localhost:8080/key/recover'
```
Expect `{"ok":true}`

### Test 3: PATCH persists via WAL
Create then patch a document:
```powershell
Invoke-RestMethod -Method Put -Uri 'http://localhost:8080/key/doc' -ContentType 'application/json' -Body '{"name":"test"}'
Invoke-RestMethod -Method Patch -Uri 'http://localhost:8080/key/doc/field/info' -ContentType 'application/json' -Body '{"nested":123}'
```

Hard-kill, remove lock, restart:
```powershell
Get-Process smoldb | Stop-Process -Force
del .\db\smoldb_lock
.\smoldb.exe --dir .\db --port 8080 start
```

**Verify:**
- `doc` includes `"info": {"nested":123}`
- `wal.log` shows a PUT for `doc` with merged content

### Test 4: Durability modes (none vs commit)

**None mode** (expected possible loss on crash):
```powershell
.\smoldb.exe --dir .\db --port 8080 --durability none --sync-mode none start
```

**Commit mode** (expected no loss for acknowledged writes):
```powershell
.\smoldb.exe --dir .\db --port 8080 --durability commit --sync-mode fsync start
```

Write PUT, hard-kill; after restart, verify acknowledged writes survive in commit mode but may be lost in none mode.

### Test 5: Sync modes (none|fsync|dsync)

**None:** Use `durability=commit, sync-mode=none` → acknowledged writes can be lost; verify loss on crash.

**Fsync:** `commit+fsync` → no loss; verify as in Test 4.

**Dsync:** `commit+dsync` behaves like fsync on Windows; verify no loss.

### Test 6: Grouped commit (time trigger)
Start grouped with time window:
```powershell
.\smoldb.exe --dir .\db --port 8080 --durability grouped --group-commit-ms 10 --sync-mode fsync start
```

Rapidly issue writes:
```powershell
1..200 | ForEach-Object { Invoke-RestMethod -Method Put -Uri "http://localhost:8080/key/t$_" -ContentType 'application/json' -Body '{"v":1}' }
```

Hard-kill quickly:
```powershell
Get-Process smoldb | Stop-Process -Force
del .\db\smoldb_lock
.\smoldb.exe --dir .\db --port 8080 start
```

**Check:**
- How many last keys are missing
- Loss should be bounded to roughly the 10ms window
- `wal.log` will show last COMMIT timestamp
- Missing keys are after the last COMMIT

### Test 7: Grouped commit (batch trigger)
Start grouped with batch window:
```powershell
.\smoldb.exe --dir .\db --port 8080 --durability grouped --group-commit-batch 50 --sync-mode fsync start
```

Write 1..200 keys rapidly, hard-kill during a batch:
```powershell
1..200 | ForEach-Object { Invoke-RestMethod -Method Put -Uri "http://localhost:8080/key/b$_" -ContentType 'application/json' -Body '{"v":1}' }
Get-Process smoldb | Stop-Process -Force
```

**After restart:**
- Ensure loss ≤ 49 acknowledged appends (the tail before next COMMIT)
- Confirm with `wal.log`: count PUTs after last COMMIT

### Test 8: COMMIT markers
Any mode with flushing (commit or grouped) will append COMMIT lines.

**Validate:**
- Tail `wal.log`; you should see lines like `{"op":"COMMIT",...}`
- Replay ignores COMMIT: crash and restart should not error
- State is determined only by PUT/DELETE entries up to last durable point

### Test 9: DELETE durability
Create then delete:
```powershell
Invoke-RestMethod -Method Put -Uri 'http://localhost:8080/key/todelete' -ContentType 'application/json' -Body '{"x":1}'
Invoke-RestMethod -Method Delete -Uri 'http://localhost:8080/key/todelete'
```

Hard-kill, remove lock, restart:
```powershell
Get-Process smoldb | Stop-Process -Force
del .\db\smoldb_lock
.\smoldb.exe --dir .\db --port 8080 start
```

**Verify:**
- GET should return 404
- `wal.log` contains DELETE and a COMMIT after the flush boundary

### Test 10: Checksum and timestamp fields
`wal.log` entries include `ts` and `csum`.

**Spot-check:**
- A line's `csum` by recomputing over `op|key|body` (documented as simple FNV-1a style)
- For quick audit, confirm `ts` increases and `csum` is present

---

## Test commands and tooling
- Unit: `go test ./... -race -v`
- Fault-injection: introduce temporary failpoints or use environment flags to trigger exits at specific lines; validate replay.
- Manual crash: Docker run server, cURL requests, then `docker kill -s KILL <container>`; restart and verify.

---

## Acceptance criteria
- No acknowledged writes lost under `DurabilityCommit` in crash scenarios.
- WAL replay restores consistent index and file contents.
- `PATCH` changes are persisted via WAL and survive crashes.
- Unit tests cover append, replay, malformed lines, and concurrency.

Additional checks for new options
- With `durability=grouped` and `--group-commit-ms=N`, acknowledged writes may lag fsync by up to N ms; confirm loss window does not exceed N ms in power-cut tests.
- With `durability=grouped` and `--group-commit-batch=B`, fsync occurs at every B appends at worst; confirm loss does not exceed B-1 most recent appends if no time trigger fires.
- WAL lines include `ts` and `csum`; verify checksum matches recomputed value for random samples.

---

## PowerShell guides: basic operations

### Prerequisites
Server running:
```powershell
# from repo root
go build -o smoldb.exe
.\smoldb.exe --port 8080 --dir .\db start
```

### Write and verify (PUT/GET)
```powershell
Invoke-RestMethod -Method Put -Uri 'http://localhost:8080/key/k1' -ContentType 'application/json' -Body '{"a":1}'
Invoke-RestMethod -Uri 'http://localhost:8080/key/k1'
```

### Patch a field (routes through WAL via Put)
```powershell
Invoke-RestMethod -Method Patch -Uri 'http://localhost:8080/key/k1/field/info' -ContentType 'application/json' -Body '{"nested":123}'
Invoke-RestMethod -Uri 'http://localhost:8080/key/k1'
```

### Delete
```powershell
Invoke-RestMethod -Method Delete -Uri 'http://localhost:8080/key/k1'
```

### Crash-recovery flow
```powershell
# window A: server is running

# window B: force-crash the process
Get-Process smoldb | Stop-Process -Force

# remove crash-time lock and restart server
del .\db\smoldb_lock
.\smoldb.exe --port 8080 --dir .\db start

# verify data (if you wrote before crashing)
Invoke-RestMethod -Uri 'http://localhost:8080/key/k1'
```

### Files to inspect
- WAL: `.\db\.smoldb\wal.log` (should contain `PUT`/`DELETE` lines)
- Lock: `.\db\smoldb_lock` (delete this after ungraceful crashes before restart)
- WAL entries now include `ts` and `csum`; use these to validate ordering/integrity during tests

---

## Scripted verification helper (optional)

**Tip:** For automated verification, you can create a small Go or PowerShell script to parse `wal.log` and report:
- Last COMMIT timestamp
- Count of PUT/DELETE after last COMMIT (at-risk writes)
- Mismatch between keys on disk and last durable set