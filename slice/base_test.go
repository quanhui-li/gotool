package slice

import (
	"fmt"
	"testing"
)

func TestDelete(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	res, val, err := Delete(arr, 4)
	if err != nil {
		fmt.Println(err)
		return
	}
	t.Log(res)
	t.Log(val)
}

func BenchmarkDelete(b *testing.B) {
	for i := 0; i < b.N; i++ {
		arr := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
		res, val, err := Delete(arr, 4)
		if err != nil {
			fmt.Println(err)
			return
		}
		b.Log(res, val)
	}
}

func BenchmarkAppendDelete(b *testing.B) {
	for i := 0; i < b.N; i++ {
		index := 4
		arr := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
		val := arr[index]
		arr = append(arr[:index], arr[index+1:]...)
		b.Log(arr, val)
	}
}
