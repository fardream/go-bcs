package bcs

// Enum emulates the [rust enum], contains only one method, IsBcsEnum, to
// indicate this is an enum in bcs.
//
// All the fields of the enum type must be pointers or interfaces except those
// ignored by "-". The index of the
// field in the field list is the integer value of the enum.
//
//	type AEnum struct {
//	  V0 *uint8 // variant 0
//	  V1 *uint16 `bcs:"-"` // ignored, so variant 1 is invalid
//	  V2 int `bcs:"-"` // cannot be set to nil, so variant 2 is invalid
//	  V3 *uint8 // variant 3
//	  v4 uint32 // ignored
//	}
//
// If there are mulitple non-nil fields when marshalling, the first one encountered will be serialized.
//
// The method IsBcsEnum doesn't actually do anything besides acting as an indicator.
//
// [rust enum]: https://doc.rust-lang.org/book/ch06-00-enums.html
type Enum interface {
	// IsBcsEnum doesn't do anything. Its function is to indicate this is an enum for bcs de/serialization.
	IsBcsEnum()
}
