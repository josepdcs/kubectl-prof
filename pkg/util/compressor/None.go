package compressor

//NoneCompressor do nothing
type NoneCompressor struct {
}

func NewNoneCompressor() *NoneCompressor {
	return &NoneCompressor{}
}

//Encode do nothing: output is equal to input
func (c *NoneCompressor) Encode(src []byte) ([]byte, error) {
	return src, nil
}

//Decode do nothing: output is equal to input
func (c *NoneCompressor) Decode(src []byte) ([]byte, error) {
	return src, nil
}
