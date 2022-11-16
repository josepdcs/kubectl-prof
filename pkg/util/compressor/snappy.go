package compressor

import "github.com/golang/snappy"

type ZstdCompressor struct {
}

func NewSnappyCompressor() *SnappyCompressor {
	return &SnappyCompressor{}
}

func (c *SnappyCompressor) Encode(src []byte) ([]byte, error) {
	return snappy.Encode(nil, src), nil
}

func (c *SnappyCompressor) Decode(src []byte) ([]byte, error) {
	return snappy.Decode(nil, src)
}
