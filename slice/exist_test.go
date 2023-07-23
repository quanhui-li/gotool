package slice

import (
	"fmt"
	"testing"
)

func TestIsExist(t *testing.T) {
	sli := []string{"1", "2", "3", "4"}
	if IsExist(sli, "5") {
		fmt.Println("true")
	} else {
		fmt.Println("false")
	}
}
