package slice

import "github.com/liquanhui-99/gotool/errors"

// delete 常规的删除操作
// @param: slice切片，index需要删除元素的下标
// @return []T 删除后的切片数据
// @return T删除的值
// @return error错误返回值
func delete[T any](sli []T, index int) ([]T, T, error) {
	var zero T
	if len(sli) == 0 {
		return nil, zero, errors.LenZeroErr()
	}
	if index > len(sli)-1 || index < 0 {
		return nil, zero, errors.OutOfRangeErr(index, len(sli))
	}

	val := sli[index]
	i := 0
	for k, v := range sli {
		if k != index {
			sli[i] = v
			i++
		}
	}
	sli = sli[:i]
	return sli, val, nil
}

// DeleteVal 删除切片中指定的值
func DeleteVal[T Comparable](sli []T, val T) ([]T, error) {
	if len(sli) <= 0 {
		return nil, errors.LenZeroErr()
	}
	i := 0
	for _, v := range sli {
		if v != val {
			sli[i] = v
			i++
		}
	}
	sli = sli[:i]
	return sli, nil
}
