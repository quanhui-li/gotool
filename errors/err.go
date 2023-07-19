package errors

import "fmt"

func OutOfRangeErr(index, length int) error {
	return fmt.Errorf("index %d out of range, length %d", index, length)
}
