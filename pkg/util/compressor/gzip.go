package compressor

import (
	"bytes"
	"io"
	"runtime"

	gzip "github.com/klauspost/pgzip"
)

type GzipCompressor struct {
}

func NewGzipCompressor() *GzipCompressor {
	return &GzipCompressor{}
}

func (c *GzipCompressor) Encode(src []byte) ([]byte, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_ = w.SetConcurrency(1<<24, runtime.GOMAXPROCS(0))
	_, err := w.Write(src)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func (c *GzipCompressor) Decode(src []byte) ([]byte, error) {
	var b bytes.Buffer
	gr, err := gzip.NewReader(bytes.NewBuffer(src))
	if err != nil {
		return nil, err
	}
	defer func(gr *gzip.Reader) {
		err := gr.Close()
		if err != nil {
			return
		}
	}(gr)
	data, err := io.ReadAll(gr)
	if err != nil {
		return nil, err
	}
	b.Write(data)
	return b.Bytes(), nil
}
