package compressor

import "io"

// NoneCompressor do nothing
type NoneCompressor struct {
}

func NewNoneCompressor() *NoneCompressor {
	return &NoneCompressor{}
}

// Encode do nothing: output is equal to input
func (c *NoneCompressor) Encode(_ io.Writer, _ io.Reader) error {
	return nil
}

// Decode do nothing: output is equal to input
func (c *NoneCompressor) Decode(_ io.Writer, _ io.Reader) error {
	return nil
}
