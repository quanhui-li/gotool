package slice

import "testing"

func TestDeleteSpecifyVal(t *testing.T) {
	sli := []int{1, 1, 1, 2, 1, 2, 3, 1, 4, 5, 6, 7, 8, 9, 10}
	sli, err := DeleteSpecifyVal(sli, 1)
	if err != nil {
		t.Log(err.Error())
	} else {
		t.Log(sli)
	}

	sli, err = DeleteSpecifyVal(sli, 2)
	if err != nil {
		t.Log(err)
	} else {
		t.Log(sli)
	}
}
