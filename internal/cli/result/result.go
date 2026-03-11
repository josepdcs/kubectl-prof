package result

import (
	"time"

	"github.com/josepdcs/kubectl-prof/api"
)

type File struct {
	FileName        string
	FileSizeInBytes int64
	Checksum        string
	Chunks          []api.ChunkData
	Timestamp       time.Time
	// Content holds the base64-encoded compressed file bytes when the agent
	// embedded them directly in the result event (no exec download needed).
	Content string
}
