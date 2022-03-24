package api

type Compressor string

const (
	None   Compressor = "none"
	Snappy Compressor = "snappy"
	Gzip   Compressor = "gzip"
	Lzo    Compressor = "lzo"
	Zstd   Compressor = "zstd"
)

var (
	compressors = []Compressor{None, Snappy, Gzip, Lzo, Zstd}
)

func AvailableCompressors() []Compressor {
	return compressors
}

func IsSupportedCompressor(event string) bool {
	return containsCompressor(Compressor(event), AvailableCompressors())
}

func containsCompressor(e Compressor, events []Compressor) bool {
	for _, current := range events {
		if e == current {
			return true
		}
	}
	return false
}
