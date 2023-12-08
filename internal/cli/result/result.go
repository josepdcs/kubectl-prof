package result

import (
	"github.com/josepdcs/kubectl-prof/api"
	"time"
)

type File struct {
	FileName        string
	FileSizeInBytes int64
	Checksum        string
	Chunks          []api.ChunkData
	Timestamp       time.Time
}
