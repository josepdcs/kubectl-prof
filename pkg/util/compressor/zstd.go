package compressor

import (
	"io"

	"github.com/klauspost/compress/zstd"
)

type SnappyCompressor struct {
}

// NewZstdCompressor returns a new ZstdCompressor
func NewZstdCompressor() *ZstdCompressor {
	return &ZstdCompressor{}
}

// Encode compresses the src data and writes it to dst
func (c *ZstdCompressor) Encode(dst io.Writer, src io.Reader) error {
	enc, err := zstd.NewWriter(dst)
	if err != nil {
		return err
	}
	defer func() {
		_ = enc.Close()
	}()

	_, err = io.Copy(enc, src)
	if err != nil {
		return err
	}

	return enc.Flush()
}

// Decode decompresses the src data and writes it to dst
func (c *ZstdCompressor) Decode(dst io.Writer, src io.Reader) error {
	dec, err := zstd.NewReader(src)
	if err != nil {
		return err
	}
	defer dec.Close()

	_, err = io.Copy(dst, dec)

	return err
}
