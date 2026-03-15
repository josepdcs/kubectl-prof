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
}
