package compressor

import (
	"io"

	"github.com/pkg/errors"
)

// Compressor is the interface that wraps the basic Encode and Decode methods
type Compressor interface {
	Encode(dst io.Writer, src io.Reader) error
	Decode(dst io.Writer, src io.Reader) error
}

// Get returns a new compressor based on the given type
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
