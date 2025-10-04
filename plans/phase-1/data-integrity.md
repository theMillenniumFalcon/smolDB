# Data Integrity Feature Implementation

## Overview
The data integrity feature uses xxHash checksums stored in `.meta` files to detect and repair potential data corruption or tampering in smolDB JSON files.

## Implementation Details

### Metadata Structure
Each JSON document has an accompanying `.meta` file containing:
```json
{
  "checksum": "xxhash64_hex_string",
  "created": "ISO8601_timestamp",
  "modified": "ISO8601_timestamp"
}
```

### Key Components
1. **Checksum Calculation (index/checksum.go)**
   - Uses xxHash for fast, reliable checksums
   - Automatically computes checksums on file writes
   - Stores metadata in `.json.meta` files

2. **File Operations (index/io.go)**
   - Validates checksums on read operations
   - Updates checksums and timestamps on writes
   - Maintains atomic file operations with mutex locks

3. **API Endpoints (api/integrity.go)**
   - `GET /integrity/:key` - Check file integrity
   - `POST /integrity/:key/repair` - Update/repair checksums

## Testing Guide

### Manual Testing

1. Create a test document:
```bash
curl -X PUT http://localhost:8080/key/test -d '{"data": "test"}'
```

2. Verify integrity:
```bash
curl http://localhost:8080/integrity/test
# Expected: "integrity check passed for key 'test'"
```

3. Simulate corruption by manually editing the JSON file:
```bash
echo '{"data": "corrupted"}' > db/test.json
```

4. Check integrity again:
```bash
curl http://localhost:8080/integrity/test
# Expected: Error indicating checksum mismatch
```

5. Repair the checksum:
```bash
curl -X POST http://localhost:8080/integrity/test/repair
# Expected: "integrity repaired for key 'test'"
```

### Automated Test Cases

1. **Basic Integrity**
   - Create new document
   - Verify checksum is created
   - Validate checksum matches content

2. **Corruption Detection**
   - Create document
   - Modify content directly (bypassing API)
   - Verify integrity check fails
   - Verify repair succeeds

3. **Update Operations**
   - Create document
   - Update via API
   - Verify checksum updates automatically
   - Verify integrity check passes

4. **Edge Cases**
   - Empty documents
   - Large documents (>1MB)
   - Special characters in content
   - Concurrent modifications

## Performance Considerations
- xxHash chosen for speed (8GB/s+ on modern hardware)
- Metadata operations add ~1-2 filesystem operations per write
- Checksum verification adds minimal overhead to reads

## Error Handling
1. **Missing Metadata**
   - Auto-repair option creates new metadata
   - Logs warning about missing metadata

2. **Corruption Detection**
   - Returns HTTP 400 with detailed error
   - Indicates original vs current checksum

3. **Repair Operations**
   - Atomic metadata updates
   - Handles concurrent access safely

## Future Enhancements
1. Batch integrity checking for multiple keys
2. Periodic background integrity scanning
3. Configurable checksum algorithms
4. Integrity statistics and monitoring
5. Integration with backup/restore operations

## Integration with Other Features
- WAL: Checksums recorded in WAL entries
- Replication: Metadata synchronized with data
- Backup: Metadata included in backups