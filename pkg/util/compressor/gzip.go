package compressor

import (
	"compress/gzip"
	"io"
)

type GzipCompressor struct {
}

// NewGzipCompressor returns a new GzipCompressor
func NewGzipCompressor() *GzipCompressor {
	return &GzipCompressor{}
}

// Encode compresses the src data and writes it to dst
func (c *GzipCompressor) Encode(dst io.Writer, src io.Reader) error {
	w := gzip.NewWriter(dst)
	defer func(w *gzip.Writer) {
		_ = w.Close()
	}(w)

	_, err := io.Copy(w, src)
	if err != nil {
		return err
	}

	return w.Flush()
}

// Decode decompresses the src data and writes it to dst
func (c *GzipCompressor) Decode(dst io.Writer, src io.Reader) error {
	r, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	defer func(r *gzip.Reader) {
		_ = r.Close()
	}(r)

	_, err = io.Copy(dst, r)

	return err
}
