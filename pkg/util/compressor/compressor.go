package compressor

import (
	"fmt"
	"github.com/josepdcs/kubectl-prof/api"
)

type Compressor interface {
	Encode(src []byte) ([]byte, error)
	Decode(str []byte) ([]byte, error)
}

func Get(c api.Compressor) (Compressor, error) {
	switch c {
	case api.None:
		return NewNoneCompressor(), nil
	case api.Snappy:
		return NewSnappyCompressor(), nil
	case api.Gzip:
		return NewGzipCompressor(), nil
	case api.Lzo:
		return NewLzoCompressor(), nil
	case api.Zstd:
		return NewZstdCompressor(), nil
	default:
		return nil, fmt.Errorf("could not find compressor for %s", c)
	}
}
