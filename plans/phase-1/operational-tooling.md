# Operational Tooling

## Overview
This document outlines the operational tooling features for smolDB, specifically the offline compaction and integrity scanning capabilities.

## Features

### 1. Offline Compaction Command
The `smoldb admin compact` command performs offline compaction of the database by:
- Rewriting JSON files to remove any fragmentation
- Trimming the Write-Ahead Log (WAL) to remove processed entries
- Optimizing storage usage and improving read performance

#### Implementation Details
- Only runs when database is offline (no active locks)
- Validates JSON integrity before and after compaction
- Updates file index after compaction
- Reports statistics on space saved and files processed

### 2. Integrity Scanner
The `smoldb admin verify` command provides integrity verification by:
- Scanning all keys in the database
- Verifying JSON syntax and structure
- Checking file-level checksums
- Optionally repairing detected issues

#### Verify Mode
- Reports issues without making changes
- Validates JSON syntax
- Checks file permissions and ownership
- Verifies index consistency
- Generates detailed report of findings

#### Repair Mode
When run with `--repair` flag:
- Fixes malformed JSON if possible
- Rebuilds index entries
- Reports unfixable issues
- Creates backup of files before repair

## Command Interface

### Compaction Command
```
smoldb admin compact [options]
  --dir     Database directory (default: ./db)
  --force   Force compaction even if lock exists
```

### Verify Command
```
smoldb admin verify [options]
  --dir     Database directory (default: ./db)
  --repair  Attempt to repair issues found
```

## Success Criteria
1. Compaction reduces storage size without data loss
2. Integrity scanner detects and reports JSON issues
3. Repair functionality fixes common corruption cases
4. Commands are safe to run with proper locking
5. Clear reporting of actions and results