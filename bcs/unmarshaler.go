package bcs

import "io"

// Unmarshaler customizes the unmarshalling behavior for a type.
//
// Compared with other Unmarshalers in golang, the Unmarshaler here takes
// a [io.Reader] instead of []byte, since it is difficult to delimit the byte streams without unmarshalling.
// Method [UnmarshalBCS] returns the number of bytes read, and potentially an error.
type Unmarshaler interface {
	UnmarshalBCS(io.Reader) (int, error)
}
