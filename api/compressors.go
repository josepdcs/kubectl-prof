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

func IsSupportedCompressor(compressor string) bool {
	return containsCompressor(Compressor(compressor), AvailableCompressors())
}

func containsCompressor(compressor Compressor, compressors []Compressor) bool {
	for _, current := range compressors {
		if compressor == current {
			return true
		}
	}
	return false
}

var GetExtensionFileByCompressor = map[Compressor]string{
	None:   "",
	Snappy: ".snappy",
	Gzip:   ".gz",
	Lzo:    ".lzo",
	Zstd:   ".zst",
}
