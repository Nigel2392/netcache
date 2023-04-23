package cache

import "unsafe"

type stringType interface {
	~string
}

type _string struct {
	str unsafe.Pointer
	len int
}

func strHash[T stringType](s T) uint64 {
	var (
		str = (*_string)(unsafe.Pointer(&s))
		h   = uint64(0)
	)
	for i := 0; i < str.len; i++ {
		h = 63*h + uint64(*(*byte)(unsafe.Pointer(uintptr(str.str) + uintptr(i))))
	}
	return h
}
