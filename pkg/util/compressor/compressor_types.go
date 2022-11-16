package compressor

type Type string

const (
	None   Type = "none"
	Snappy      = "snappy"
	Gzip        = "gzip"
	Lzo         = "lzo"
	Zstd        = "zstd"
)

var (
	compressors = []Type{None, Snappy, Gzip, Lzo, Zstd}
)

func AvailableCompressors() []Type {
	return compressors
}

func IsSupportedCompressor(compressor string) bool {
	return containsCompressor(Type(compressor), AvailableCompressors())
}

func containsCompressor(compressor Type, compressors []Type) bool {
	for _, current := range compressors {
		if compressor == current {
			return true
		}
	}
	return false
}

var GetExtensionFileByCompressor = map[Type]string{
	None:   "",
	Snappy: ".snappy",
	Gzip:   ".gz",
	Lzo:    ".lzo",
	Zstd:   ".zst",
}
