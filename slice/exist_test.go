package slice

import (
	"fmt"
	"testing"
)

func TestStringIsExist(t *testing.T) {
	sli := []string{"1", "2", "3", "4"}
	if IsExist(sli, "3") {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}
}

func TestIntIsExist(t *testing.T) {
	sli := []int{1, 2, 3, 4, 5, 6, 7, 8, 9}
	if IsExist(sli, 1) {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}
}

func TestFloatIsExist(t *testing.T) {
	sli := []float64{1.1, 2.2, 3.3, 4.4, 5.5}
	if IsExist(sli, 3.1) {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}
}
