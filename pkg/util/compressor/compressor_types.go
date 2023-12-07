package compressor

import "github.com/samber/lo"

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
	return lo.Contains(AvailableCompressors(), Type(compressor))
}

var GetExtensionFileByCompressor = map[Type]string{
	None:   "",
	Snappy: ".snappy",
	Gzip:   ".gz",
	Lzo:    ".lzo",
	Zstd:   ".zst",
}
