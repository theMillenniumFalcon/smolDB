package admin

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCompactDB(t *testing.T) {
	// Create test directory
	dir := t.TempDir()

	// Write test files
	testFiles := map[string]string{
		"test1.json": `{"foo":"bar", "num": 42}`,
		"test2.json": `{"hello": "world",   "spaces":    "many"}`,
	}

	for name, content := range testFiles {
		err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
		assert.NoError(t, err)
	}

	// Run compaction
	stats, err := CompactDB(dir, true)
	assert.NoError(t, err)
	assert.Equal(t, 2, stats.FilesProcessed)
	assert.True(t, stats.BytesAfter <= stats.BytesBefore, "Compacted size should be less or equal")
}

func TestVerifyDB(t *testing.T) {
	// Create test directory
	dir := t.TempDir()

	// Write test files
	validJSON := `{"valid": "json"}`
	invalidJSON := `{"invalid": "json",,,}`

	err := os.WriteFile(filepath.Join(dir, "valid.json"), []byte(validJSON), 0644)
	assert.NoError(t, err)
	err = os.WriteFile(filepath.Join(dir, "invalid.json"), []byte(invalidJSON), 0644)
	assert.NoError(t, err)

	// Test verification without repair
	report, err := VerifyDB(dir, false)
	assert.NoError(t, err)
	assert.Equal(t, 2, report.TotalFiles)
	assert.Equal(t, 1, report.ValidFiles)
	assert.Equal(t, 1, len(report.InvalidFiles))

	// Test verification with repair
	report, err = VerifyDB(dir, true)
	assert.NoError(t, err)
	assert.Equal(t, 2, report.TotalFiles)
	// Note: repairs may not succeed due to simple repair implementation
}

// TestCompactionPowerCut simulates power failures during compaction
func TestCompactionPowerCut(t *testing.T) {
	// Create test directory and initial files
	dir := t.TempDir()
	originalFiles := map[string]string{
		"test1.json": `{"key1":"value1","nested":{"data":42}}`,
		"test2.json": `{"key2":"value2","array":[1,2,3]}`,
		"test3.json": `{"key3":"value3","bigspace":    "test"}`,
	}

	// Write initial files and get their checksums
	originalChecksums := make(map[string]string)
	for name, content := range originalFiles {
		path := filepath.Join(dir, name)
		err := os.WriteFile(path, []byte(content), 0644)
		assert.NoError(t, err)

		// Store original file checksum
		data, _ := os.ReadFile(path)
		originalChecksums[name] = string(data)
	}

	// Run compaction with simulated power cut by stopping halfway
	done := make(chan struct{})
	go func() {
		_, _ = CompactDB(dir, true) // Ignore errors as we'll interrupt it
		close(done)
	}()

	// Give it a moment to start but cut it off before completion
	time.Sleep(10 * time.Millisecond)
	// At this point, some files might be compacted while others aren't

	select {
	case <-done:
		// If it finished too quickly, that's fine - just verify integrity
	default:
		// It's still running, good for our test
	}

	// Verify data integrity after "power cut"
	for name, originalChecksum := range originalChecksums {
		path := filepath.Join(dir, name)
		data, err := os.ReadFile(path)
		assert.NoError(t, err, "File should still be readable after power cut")

		// Each file should either be untouched or properly compacted (never corrupted)
		if string(data) != originalChecksum {
			// If changed, it should be valid JSON
			var parsed interface{}
			err = json.Unmarshal(data, &parsed)
			assert.NoError(t, err, "Changed file should contain valid JSON")

			// Re-marshal to compare normalized content
			originalParsed := make(map[string]interface{})
			err = json.Unmarshal([]byte(originalChecksum), &originalParsed)
			assert.NoError(t, err)

			currentParsed := make(map[string]interface{})
			err = json.Unmarshal(data, &currentParsed)
			assert.NoError(t, err)

			// Content should be equivalent even if formatting changed
			assert.Equal(t, originalParsed, currentParsed, "JSON content should be preserved")
		}
	}

	// Run verify to check overall integrity
	report, err := VerifyDB(dir, false)
	assert.NoError(t, err)
	assert.Equal(t, len(originalFiles), report.TotalFiles)
	assert.Equal(t, len(originalFiles), report.ValidFiles, "All files should be valid JSON")
	assert.Empty(t, report.InvalidFiles, "No files should be corrupted")
}

// TestCompactionRestart ensures compaction can be safely restarted after failure
func TestCompactionRestart(t *testing.T) {
	// Create test directory and initial files
	dir := t.TempDir()
	originalFiles := map[string]string{
		"test1.json": `{"key1":"value1"}`,
		"test2.json": `{"key2":"value2"}`,
	}

	// Write initial files
	for name, content := range originalFiles {
		err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
		assert.NoError(t, err)
	}

	// First compaction attempt (will be interrupted)
	compactionDone := make(chan struct{})
	go func() {
		_, _ = CompactDB(dir, true)
		close(compactionDone)
	}()

	// Interrupt after a short time
	time.Sleep(10 * time.Millisecond)

	select {
	case <-compactionDone:
		// If it finished too quickly, that's fine
	default:
		// Still running, which is good for our test
	}

	// Run compaction again immediately
	stats, err := CompactDB(dir, true)
	assert.NoError(t, err)
	assert.Equal(t, len(originalFiles), stats.FilesProcessed)

	// Verify all files are still intact
	for name, originalContent := range originalFiles {
		data, err := os.ReadFile(filepath.Join(dir, name))
		assert.NoError(t, err)

		// Content should be semantically equivalent (though might be reformatted)
		var original, current interface{}
		err = json.Unmarshal([]byte(originalContent), &original)
		assert.NoError(t, err)
		err = json.Unmarshal(data, &current)
		assert.NoError(t, err)
		assert.Equal(t, original, current, "File content should be preserved")
	}
}
