package slice

import "reflect"

// IsExist 判断元素是否在切片中
func IsExist[T any](sli []T, param T) bool {
	for _, item := range sli {
		a := reflect.ValueOf(item)
		b := reflect.ValueOf(param)
		if a == b {
			return true
		}
	}
	return false
}
