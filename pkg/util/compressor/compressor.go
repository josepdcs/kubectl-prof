package compressor

import (
	"github.com/pkg/errors"
)

type Compressor interface {
	Encode(src []byte) ([]byte, error)
	Decode(str []byte) ([]byte, error)
}

func Get(c Type) (Compressor, error) {
	switch c {
	case None:
		return NewNoneCompressor(), nil
	case Snappy:
		return NewSnappyCompressor(), nil
	case Gzip:
		return NewGzipCompressor(), nil
	case Lzo:
		return NewLzoCompressor(), nil
	case Zstd:
		return NewZstdCompressor(), nil
	default:
		return nil, errors.Errorf("could not find compressor for %s", c)
	}
}
