package index

import (
	"encoding/json"
	"fmt"

	"github.com/cespare/xxhash/v2"
	af "github.com/spf13/afero"
)

// MetaData represents the metadata stored alongside each JSON file
type MetaData struct {
	Checksum string `json:"checksum"` // xxHash checksum of the JSON content
	Created  string `json:"created"`  // ISO timestamp when file was created
	Modified string `json:"modified"` // ISO timestamp of last modification
}

// calculateChecksum computes the xxHash checksum of the given bytes
func calculateChecksum(data []byte) string {
	hash := xxhash.Sum64(data)
	return fmt.Sprintf("%016x", hash)
}

// writeMetadata stores the metadata for a file
func (f *File) writeMetadata(meta *MetaData) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	bytes, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %v", err)
	}

	metaPath := f.resolveMetaPath()
	err = af.WriteFile(I.FileSystem, metaPath, bytes, 0644)
	if err != nil {
		return fmt.Errorf("failed to write metadata file: %v", err)
	}

	return nil
}

// readMetadata reads the metadata for a file
func (f *File) readMetadata() (*MetaData, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	metaPath := f.resolveMetaPath()
	bytes, err := af.ReadFile(I.FileSystem, metaPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata file: %v", err)
	}

	var meta MetaData
	err = json.Unmarshal(bytes, &meta)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata: %v", err)
	}

	return &meta, nil
}

// resolveMetaPath returns the path to the metadata file
func (f *File) resolveMetaPath() string {
	return f.ResolvePath() + ".meta"
}

// ValidateChecksum verifies the integrity of the file content against its stored checksum
func (f *File) ValidateChecksum() error {
	bytes, err := f.GetByteArray()
	if err != nil {
		return fmt.Errorf("failed to read file content: %v", err)
	}

	meta, err := f.readMetadata()
	if err != nil {
		return fmt.Errorf("failed to read metadata: %v", err)
	}

	currentChecksum := calculateChecksum(bytes)
	if currentChecksum != meta.Checksum {
		return fmt.Errorf("checksum mismatch: stored=%s, calculated=%s", meta.Checksum, currentChecksum)
	}

	return nil
}

// RepairChecksum updates the stored checksum to match the current file content
func (f *File) RepairChecksum() error {
	bytes, err := f.GetByteArray()
	if err != nil {
		return fmt.Errorf("failed to read file content: %v", err)
	}

	meta, err := f.readMetadata()
	if err != nil {
		// If metadata doesn't exist, create a new one
		meta = &MetaData{}
	}

	meta.Checksum = calculateChecksum(bytes)
	return f.writeMetadata(meta)
}
