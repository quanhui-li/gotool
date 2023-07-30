package slice

import (
	"fmt"
	"testing"
)

func TestDelete(t *testing.T) {
	arr := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	res, err := Delete(arr, 4)
	if err != nil {
		fmt.Println(err)
		return
	}
	t.Log(res)
}

func BenchmarkDelete(b *testing.B) {
	for i := 0; i < b.N; i++ {
		arr := make([]int, 0, 3000)
		for i := 0; i < 3000; i++ {
			arr = append(arr, i)
		}
		res, val, err := DeleteWithVal(arr, 2000)
		if err != nil {
			fmt.Println(err)
			return
		}
		b.Log(res, val)
	}
}

func BenchmarkAppendDelete(b *testing.B) {
	for i := 0; i < b.N; i++ {
		arr := make([]int, 0, 3000)
		for i := 0; i < 3000; i++ {
			arr = append(arr, i)
		}
		index := 2000
		val := arr[index]
		arr = append(arr[:index], arr[index+1:]...)
		b.Log(arr, val)
	}
}

func BenchmarkDeleteV1(b *testing.B) {
	arr := make([]int, 0, 10000)
	for i := 0; i < 10000; i++ {
		arr = append(arr, i)
	}
	for i := 0; i < b.N; i++ {
		index := 2000
		val := arr[index]
		arr = append(arr[:index], arr[index+1:]...)
		b.Log(arr, val)
	}
}

func TestDeleteVal(t *testing.T) {
	arr := []string{"a", "b", "c", "d", "a", "a", "e"}
	arr, err := DeleteVal(arr, "a")
	if err != nil {
		t.Log(err.Error())
	} else {
		fmt.Println(arr)
	}
}
