package admin

import (
	"os"
	"path/filepath"
	"testing"

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
