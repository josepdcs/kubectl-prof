package compressor

import (
	"io"

	"github.com/rasky/go-lzo"
)

type LzoCompressor struct {
}

// NewLzoCompressor returns a new LzoCompressor
func NewLzoCompressor() *LzoCompressor {
	return &LzoCompressor{}
}

// Encode compresses the src data and writes it to dst
func (c *LzoCompressor) Encode(dst io.Writer, src io.Reader) error {
	b, err := io.ReadAll(src)
	if err != nil {
		return err
	}

	d := lzo.Compress1X999(b)
	_, err = dst.Write(d)

	return err
}

// Decode decompresses the src data and writes it to dst
func (c *LzoCompressor) Decode(dst io.Writer, src io.Reader) error {
	b, err := lzo.Decompress1X(src, 0, 0)
	if err != nil {
		return err
	}

	_, err = dst.Write(b)
	if err != nil {
		return err
	}

	return nil
}
