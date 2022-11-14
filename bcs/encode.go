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
func (e *Encoder) Encode(v any) error {
	return e.encode(reflect.ValueOf(v))
}

// encode a value with the tagValue.
// tagValue can be optional or ignore, and tagValue should not be passed onto
// sub encodings.
func (e *Encoder) encode(v reflect.Value) error {
	kind := v.Kind()

	if !v.CanInterface() {
		return nil
	}

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

	switch kind {
	case reflect.Bool, // boolean
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, // all the ints
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64: // all the uints
		return binary.Write(e.w, binary.LittleEndian, v.Interface())
	case reflect.Ptr:
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
func Marshal(v any) ([]byte, error) {
	var b bytes.Buffer
	e := NewEncoder(&b)

	if err := e.Encode(v); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}
