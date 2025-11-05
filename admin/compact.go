package admin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/themillenniumfalcon/smolDB/index"
)

// CompactDB performs database compaction by rewriting JSON files and trimming WAL
func CompactDB(dir string, force bool) (*CompactionStats, error) {
	// Check for active lock unless force flag is used
	if !force {
		if _, err := os.Stat(filepath.Join(dir, "smoldb_lock")); !os.IsNotExist(err) {
			return nil, fmt.Errorf("database is in use (lock exists). Use --force to override")
		}
	}

	stats := &CompactionStats{}
	idx := index.NewFileIndex(dir)

	// Process each file
	for _, key := range idx.ListKeys() {
		file, ok := idx.Lookup(key)
		if !ok {
			continue
		}

		// Read original file
		filePath := filepath.Join(dir, file.FileName+".json")
		data, err := os.ReadFile(filePath)
		if err != nil {
			return stats, fmt.Errorf("error reading %s: %v", key, err)
		}
		stats.BytesBefore += int64(len(data))

		// Parse and rewrite JSON to remove any fragmentation
		var parsed interface{}
		if err := json.Unmarshal(data, &parsed); err != nil {
			return stats, fmt.Errorf("invalid JSON in %s: %v", key, err)
		}

		compacted, err := json.Marshal(parsed)
		if err != nil {
			return stats, fmt.Errorf("error rewriting %s: %v", key, err)
		}

		if err := os.WriteFile(filePath, compacted, 0644); err != nil {
			return stats, fmt.Errorf("error saving %s: %v", key, err)
		}

		stats.BytesAfter += int64(len(compacted))
		stats.FilesProcessed++
	}

	// TODO: WAL trimming when implemented
	// stats.WalEntriesTrimmed = trimWAL(dir)

	return stats, nil
}
