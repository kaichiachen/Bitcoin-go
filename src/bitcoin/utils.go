package bitcoin

func FitBytes(bs []byte, l int) []byte {
	if len(bs) < l {
		for len(bs) < l {
			bs = append([]byte{0}, bs...)
		}
	}
	return bs[:l]
}

func ArrayOfBytes(len int, unit byte) []byte {
	bs := []byte{}

	for len != 0 {
		bs = append(bs, unit)
		len--
	}

	return bs
}
