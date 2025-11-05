// Package admin provides maintenance and operational tools for smolDB
package admin

// CompactionStats tracks statistics during compaction
type CompactionStats struct {
	FilesProcessed    int
	BytesBefore       int64
	BytesAfter        int64
	WalEntriesTrimmed int
}

// IntegrityReport contains details about database integrity scan
type IntegrityReport struct {
	TotalFiles      int
	ValidFiles      int
	InvalidFiles    []string
	Repairs         []string
	IndexMismatches []string
}
