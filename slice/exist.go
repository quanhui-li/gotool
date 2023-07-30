package slice

// IsExist 判断元素是否在切片中
func IsExist[T Comparable](sli []T, param T) bool {
	for _, item := range sli {
		if item == param {
			return true
		}
	}
	return false
}
