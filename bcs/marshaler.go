package bcs

// Marshaler customizes the marshalling behavior for a type
type Marshaler interface {
	MarshalBCS() ([]byte, error)
}

// Unmarshaler customizes the unmarshalling behavior for a type.
//
// This is different from many other unmarshalers in golang that it
// returns the bytes consumed and the error. When the error is not nil,
// the byte consumed is guaranteed to be 0.
type Unmarshaler interface {
	UnmarshalBCS([]byte) (int, error)
}
