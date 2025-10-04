# Crash Recovery in smolDB

## Overview
smolDB implements a robust crash recovery mechanism using a combination of Write-Ahead Logging (WAL) and periodic checkpoints. This document describes how the crash recovery system works and how to test it.

## Components

### 1. Write-Ahead Log (WAL)
- Located at `<db_dir>/.smoldb/wal.log`
- Records all mutations (PUT/DELETE) before they're applied
- Contains checksummed entries in JSON format
- Used for replaying changes after a crash
- Automatically truncated after successful checkpoints

### 2. Checkpoints
- Located at `<db_dir>/checkpoint/000000000.snap`
- Contains a complete snapshot of database state at a point in time
- Includes:
  - Timestamp
  - Complete key-value state
  - WAL offset marker
- Created periodically (default: every 5 minutes)
- Reduces recovery time by avoiding full WAL replay
- Triggers WAL truncation to prevent unbounded growth

## Recovery Process

When smolDB starts up, it follows this sequence:

1. **Lock File Check**
   - Verifies if `smoldb_lock` exists
   - If present, warns about unclean shutdown
   
2. **Checkpoint Restoration**
   - Locates latest checkpoint file in `checkpoint/` directory
   - Restores database state from checkpoint
   - Notes the WAL offset where checkpoint was taken

3. **WAL Replay**
   - Opens WAL file
   - Seeks to last checkpoint offset
   - Replays all operations after that point
   - Stops at first corrupted/incomplete entry

4. **Index Regeneration**
   - Rebuilds in-memory index from restored state
   - Verifies consistency with on-disk files

## WAL Management

### Automatic Truncation
After each successful checkpoint creation:
1. Current WAL offset is recorded in checkpoint metadata
2. WAL is safely truncated to this offset using an atomic operation:
   - Creates temporary file with preserved content
   - Copies content up to checkpoint offset
   - Atomically replaces old WAL with truncated version
3. New writes continue appending after truncation point

This process ensures:
- WAL size remains proportional to uncommitted changes
- Disk space is managed efficiently
- No risk of data loss during truncation
- Atomic operations maintain consistency

## Testing Crash Recovery

### 1. Basic Crash Test

```powershell
# Start server in first terminal
.\smoldb.exe --dir .\db --port 8080 start

# In second terminal, insert some data
$body = @{
    "name" = "test"
    "value" = 42
} | ConvertTo-Json
Invoke-RestMethod -Method Put -Uri 'http://localhost:8080/key/k1' -Body $body -ContentType 'application/json'

# Force crash the server
Get-Process smoldb | Stop-Process -Force

# Delete lock file
Remove-Item .\db\smoldb_lock

# Restart server
.\smoldb.exe --dir .\db --port 8080 start

# Verify data survived
Invoke-RestMethod -Uri 'http://localhost:8080/key/k1'
```

### 2. Checkpoint Test

```powershell
# Start server
.\smoldb.exe --dir .\db --port 8080 start

# Insert multiple records
1..10 | ForEach-Object {
    $body = @{
        "name" = "test$_"
        "value" = $_
    } | ConvertTo-Json
    Invoke-RestMethod -Method Put -Uri "http://localhost:8080/key/k$_" -Body $body -ContentType 'application/json'
    Start-Sleep -Seconds 1  # Space out writes
}

# Wait for checkpoint (default: 5 minutes)
Start-Sleep -Seconds 300

# Verify WAL truncation
$walSize = (Get-Item .\db\.smoldb\wal.log).Length
Write-Host "WAL size after checkpoint: $walSize bytes"

# Force crash
Get-Process smoldb | Stop-Process -Force

# Delete some files manually to simulate corruption
Remove-Item .\db\k1.json
Remove-Item .\db\k2.json

# Delete lock
Remove-Item .\db\smoldb_lock

# Restart and verify recovery
.\smoldb.exe --dir .\db --port 8080 start

# Check if data was recovered from checkpoint
Invoke-RestMethod -Uri 'http://localhost:8080/key/k1'
```

### 3. WAL Replay Test

```powershell
# Start server
.\smoldb.exe --dir .\db --port 8080 start

# Wait for checkpoint
Start-Sleep -Seconds 300

# Write new data
$body = @{
    "name" = "post_checkpoint"
    "value" = 100
} | ConvertTo-Json
Invoke-RestMethod -Method Put -Uri 'http://localhost:8080/key/post1' -Body $body -ContentType 'application/json'

# Force crash immediately
Get-Process smoldb | Stop-Process -Force

# Restart and verify WAL replay
Remove-Item .\db\smoldb_lock
.\smoldb.exe --dir .\db --port 8080 start

# Verify post-checkpoint write survived
Invoke-RestMethod -Uri 'http://localhost:8080/key/post1'
```

## Verification Points

After running crash recovery tests, verify:

1. **Data Integrity**
   - All committed writes before crash are present
   - No partial/corrupted records exist
   - Index is consistent with on-disk state

2. **Checkpoint Files**
   - Located in `checkpoint/` directory
   - Named with timestamp format
   - Contains valid JSON state

3. **WAL File**
   - Located at `.smoldb/wal.log`
   - Contains valid JSON entries
   - Entries have timestamps and checksums
   - Size decreases after checkpoints

## Common Issues

1. **Lock File Remains**
   - Symptom: Server won't start, complains about lock
   - Solution: Remove `smoldb_lock` after confirming no other instance is running

2. **Corrupted WAL**
   - Symptom: Replay stops at corrupted entry
   - Solution: Server will truncate WAL at last valid entry
   - No manual intervention needed

3. **Missing Checkpoint**
   - Symptom: Full WAL replay on startup
   - Solution: Normal behavior if no checkpoint exists
   - Will create new checkpoint on next interval

4. **Failed WAL Truncation**
   - Symptom: WAL size continues growing
   - Solution: Check disk space and permissions
   - Server continues operating safely

## Best Practices

1. Always use appropriate durability settings:
   ```powershell
   .\smoldb.exe --dir .\db --port 8080 --durability commit --sync-mode fsync start
   ```

2. Monitor checkpoint creation:
   - Check `checkpoint/` directory for recent files
   - Verify checkpoint size growth is reasonable
   - Verify WAL size decreases after checkpoints

3. Regular Maintenance:
   - Periodically remove old checkpoint files
   - Keep WAL size in check
   - Monitor disk space

4. Backup Strategy:
   - Take backup when creating checkpoint
   - Include both checkpoint and WAL files
   - Document backup timestamp for recovery