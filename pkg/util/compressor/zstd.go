package compressor

import "github.com/klauspost/compress/zstd"

type SnappyCompressor struct {
}

func NewZstdCompressor() *ZstdCompressor {
	return &ZstdCompressor{}
}

func (c *ZstdCompressor) Encode(src []byte) ([]byte, error) {
	encoder, _ := zstd.NewWriter(nil)
	return encoder.EncodeAll(src, make([]byte, 0, len(src))), nil
}

func (c *ZstdCompressor) Decode(src []byte) ([]byte, error) {
	decoder, _ := zstd.NewReader(nil)
	return decoder.DecodeAll(src, nil)
}
