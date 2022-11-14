package bcs

import "fmt"

// Unmarshal unmarshal the bcs serialized data into v.
// It returns the number of bytes consumed and a possible error.
// If error is not nil, the consumed bytes will be 0.
func Unmarshal(data []byte, v any) (int, error) {
	return 0, fmt.Errorf("unimplemented")
}
