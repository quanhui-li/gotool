package slice

// Comparable 泛型约束，可以比较的数据类型
type Comparable interface {
	~int | ~int32 | ~int64 | ~uint | ~uint8 | ~uint32 |
		~uint64 | ~float32 | ~float64 | ~string | ~bool
}
