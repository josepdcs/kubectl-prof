package compressor

import (
	"bytes"
	"github.com/rasky/go-lzo"
)

type LzoCompressor struct {
}

func NewLzoCompressor() *LzoCompressor {
	return &LzoCompressor{}
}

func (c *LzoCompressor) Encode(src []byte) ([]byte, error) {
	return lzo.Compress1X999(src), nil
}

func (c *LzoCompressor) Decode(src []byte) ([]byte, error) {
	var b bytes.Buffer
	(&b).Write(src)
	return lzo.Decompress1X(&b, 0, 0)
}
