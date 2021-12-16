package ordmap

type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64 |
		~string
}

func CompareOrdered[T Ordered](v1, v2 T) bool {
	return v1 < v2
}
