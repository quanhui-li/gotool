package slice

// Delete 删除指定下标的元素，不返回删除的元素值
func Delete[T any](sli []T, index int) ([]T, error) {
	res, _, err := delete(sli, index)
	return res, err
}

// DeleteWithVal 删除指定下标的元素并获取到删除的元素值
func DeleteWithVal[T any](sli []T, index int) ([]T, T, error) {
	return delete(sli, index)
}
