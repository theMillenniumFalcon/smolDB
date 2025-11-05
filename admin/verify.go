package admin

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/themillenniumfalcon/smolDB/index"
)

// VerifyDB scans database files and checks for integrity issues
func VerifyDB(dir string, repair bool) (*IntegrityReport, error) {
	report := &IntegrityReport{}
	idx := index.NewFileIndex(dir)
	var mu sync.Mutex
	var wg sync.WaitGroup

	keys := idx.ListKeys()
	report.TotalFiles = len(keys)

	// Process files concurrently for faster verification
	for _, key := range keys {
		wg.Add(1)
		go func(key string) {
			defer wg.Done()

			file, ok := idx.Lookup(key)
			if !ok {
				mu.Lock()
				report.IndexMismatches = append(report.IndexMismatches, key)
				mu.Unlock()
				return
			}

			// Read and validate JSON
			filePath := filepath.Join(dir, file.FileName+".json")
			data, err := os.ReadFile(filePath)
			if err != nil {
				mu.Lock()
				report.InvalidFiles = append(report.InvalidFiles, fmt.Sprintf("%s (read error: %v)", key, err))
				mu.Unlock()
				return
			}

			var parsed interface{}
			if err := json.Unmarshal(data, &parsed); err != nil {
				if repair {
					// Try to fix common JSON issues
					if fixed, err := repairJSON(data); err == nil {
						if err := os.WriteFile(filePath, fixed, 0644); err == nil {
							mu.Lock()
							report.Repairs = append(report.Repairs, key)
							report.ValidFiles++
							mu.Unlock()
							return
						}
					}
				}
				mu.Lock()
				report.InvalidFiles = append(report.InvalidFiles, fmt.Sprintf("%s (invalid JSON: %v)", key, err))
				mu.Unlock()
				return
			}

			mu.Lock()
			report.ValidFiles++
			mu.Unlock()
		}(key)
	}

	wg.Wait()
	return report, nil
}

// repairJSON attempts to fix common JSON formatting issues
func repairJSON(data []byte) ([]byte, error) {
	var parsed interface{}

	// First try strict unmarshal
	if err := json.Unmarshal(data, &parsed); err == nil {
		// Already valid JSON
		return json.Marshal(parsed)
	}

	// TODO: Add common repair strategies:
	// - Remove trailing commas
	// - Fix unescaped quotes
	// - Handle missing closing braces/brackets

	return nil, fmt.Errorf("unable to repair JSON")
}
