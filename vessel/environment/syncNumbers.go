package environment

type syncNumber uint8

func (s syncNumber) next() syncNumber {
	if s == 255 {
		return 0
	} else {
		return s + 1
	}
}

func (s syncNumber) newerThan(t syncNumber) bool {
	if s < 128 {
		return s < t && t < s+128
	} else {
		return !(s-128 < t && t < s)
	}
}
