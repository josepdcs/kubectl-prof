package compressor

import (
	"io"

	"github.com/golang/snappy"
)

type ZstdCompressor struct {
}

// NewSnappyCompressor returns a new SnappyCompressor
func NewSnappyCompressor() *SnappyCompressor {
	return &SnappyCompressor{}
}

// Encode compresses the src data and writes it to dst
func (c *SnappyCompressor) Encode(dst io.Writer, src io.Reader) error {
	b, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	d := snappy.Encode(nil, b)
	_, err = dst.Write(d)
	if err != nil {
		return err
	}

	return nil
}

// Decode decompresses the src data and writes it to dst
func (c *SnappyCompressor) Decode(dst io.Writer, src io.Reader) error {
	b, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	e, err := snappy.Decode(nil, b)
	if err != nil {
		return err
	}

	_, err = dst.Write(e)
	if err != nil {
		return err
	}

	return nil
}
