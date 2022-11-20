package bcs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

// Encoder takes an [io.Writer] and encodes value into it.
type Encoder struct {
	w io.Writer
}

// NewEncoder creates a new [Encoder] from an [io.Writer]
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: w,
	}
}

// Encode a value v into the encoder.
//
//   - If the value is [Marshaler], then the corresponding
//     MarshalBCS implementation will be called.
//   - If the value is [Enum], then it will be special handled for enum.
func (e *Encoder) Encode(v any) error {
	return e.encode(reflect.ValueOf(v))
}

// encode a value
func (e *Encoder) encode(v reflect.Value) error {
	// if v not CanInterface,
	// this value is an unexported value, skip it.
	if !v.CanInterface() {
		return nil
	}

	// test for the two enums we defined.
	// 1. Marshaler
	// 2. Enum.
	i := v.Interface()
	if m, ismarshaler := i.(Marshaler); ismarshaler {
		bytes, err := m.MarshalBCS()
		if err != nil {
			return err
		}

		_, err = e.w.Write(bytes)

		return err
	}
	if _, isenum := i.(Enum); isenum {
		return e.encodeEnum(reflect.Indirect(v))
	}

	kind := v.Kind()

	switch kind {
	case reflect.Bool, // boolean
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, // all the ints
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64: // all the uints
		// use little endian to encode those.
		return binary.Write(e.w, binary.LittleEndian, v.Interface())
	case reflect.Ptr:
		// if v is nil pointer, use the zero value for v.
		// we don't check for optional flag here.
		// that should be checked when the container struct is encoded,
		// if this pointer is contained in a struct.
		if v.IsNil() {
			return e.encode(reflect.Indirect(reflect.New(v.Type())))
		} else {
			return e.encode(reflect.Indirect(v))
		}
	case reflect.Slice:
		return e.encodeSlice(v)
	case reflect.String:
		str := []byte(v.String())
		return e.encodeByteSlice(str)
	case reflect.Struct:
		return e.encodeStruct(v)
	default:
		return fmt.Errorf("unsupported kind: %s", kind.String())
	}
}

// encodeEnum encodes an [Enum]
func (e *Encoder) encodeEnum(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		// ignore fields that are not exported
		if !field.CanInterface() {
			continue
		}

		fieldType := t.Field(i)
		// check the tag
		tag, err := parseTagValue(fieldType.Tag.Get(tagName))
		if err != nil {
			return err
		}
		if tag&tagValue_Ignore > 0 {
			continue
		}
		fieldKind := field.Kind()
		if fieldKind != reflect.Pointer && fieldKind != reflect.Interface {
			return fmt.Errorf("enum only supports fields that are either pointers or interfaces, unless they are ignored")
		}
		if !field.IsNil() {
			if _, err := e.w.Write(ULEB128Encode(i)); err != nil {
				return err
			}
			if fieldKind == reflect.Pointer {
				return e.encode(reflect.Indirect(field))
			} else {
				return e.encode(v)
			}
		}
	}

	return fmt.Errorf("no field is set in the enum")
}

// encodeByteSlice is specialized since bytes those can be simply put into the output.
func (e *Encoder) encodeByteSlice(b []byte) error {
	l := len(b)
	if _, err := e.w.Write(ULEB128Encode(l)); err != nil {
		return err
	}

	if _, err := e.w.Write(b); err != nil {
		return err
	}

	return nil
}

func (e *Encoder) encodeSlice(v reflect.Value) error {
	length := v.Len()
	if _, err := e.w.Write(ULEB128Encode(length)); err != nil {
		return err
	}

	for i := 0; i < length; i++ {
		if err := e.encode(v.Index(i)); err != nil {
			return err
		}
	}

	return nil
}

func (e *Encoder) encodeStruct(v reflect.Value) error {
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanInterface() {
			continue
		}
		tag, err := parseTagValue(t.Field(i).Tag.Get(tagName))
		if err != nil {
			return err
		}
		if tag&tagValue_Ignore != 0 {
			continue
		}

		if tag&tagValue_Optional != 0 {
			if v.Kind() != reflect.Pointer && v.Kind() != reflect.Interface {
				return fmt.Errorf("optional field can only be pointer or interface")
			}
			if field.IsNil() {
				_, err := e.w.Write([]byte{0})
				if err != nil {
					return err
				}
			} else {
				if _, err := e.w.Write([]byte{1}); err != nil {
					return err
				}
				if err := e.encode(reflect.Indirect(field)); err != nil {
					return err
				}
			}
		} else if err := e.encode(field); err != nil {
			return err
		}
	}

	return nil
}

// Marshal a value into bcs bytes.
//
// Many constructs supported by bcs don't exist in golang or move-lang.
//
//   - [Enum] is used to simulate the effects of rust enum.
//   - Use tag `optional` to indicate an optional value in rust.
//     the field must be pointer or interface.
//   - Use tag `-` to ignore fields.
//   - Unexported fields are ignored.
//
// Note that bcs doesn't have schema, and field names are irrelavant. The fields
// of struct are serialized in the order that they are defined.
//
// Pointers are serialized as the type they point to. Nil pointers will be serialized
// as zero value of the type they point to unless it's marked as `optional`.
//
// During marshalling process, how v is marshalled depends on if v implemented [Marshaler] or [Enum]
//  1. if [Marshaler], use "MarshalBCS" method of the class.
//  2. if not [Marshaler] but [Enum], use specialization for [Enum]
//  3. otherwise standard process.
func Marshal(v any) ([]byte, error) {
	var b bytes.Buffer
	e := NewEncoder(&b)

	if err := e.Encode(v); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

// MustMarshal [Marshal] v, and panics if error.
func MustMarshal(v any) []byte {
	result, err := Marshal(v)
	if err != nil {
		panic(err)
	}

	return result
}
