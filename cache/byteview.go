package cache

type ByteView struct {
	bytes []byte
}

func (bv ByteView) Len() int {
	return len(bv.bytes)
}

func (bv ByteView) Bytes() []byte {
	return bv.cloneBytes()
}

func (bv ByteView) String() string {
	return string(bv.bytes)
}

func (bv ByteView) cloneBytes() []byte {
	if bv.bytes == nil {
		return nil
	}
	cloned := make([]byte, len(bv.bytes))
	copy(cloned, bv.bytes)
	return cloned
}