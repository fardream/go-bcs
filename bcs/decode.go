package bcs

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

// Unmarshal unmarshal the bcs serialized data into v.
//
// Refer to notes in [Marshal] for details how data serialized/deserialized.
//
// During the unmarshalling process
//  1. if [Unmarshaler], use "UnmarshalBCS" method.
//  2. if not [Unmarshaler] but [Enum], use the specialization for [Enum].
//  3. otherwise standard process.
func Unmarshal(data []byte, v any) error {
	return NewDecoder(bytes.NewReader(data)).Decode(v)
}

// Decoder takes an [io.Reader] and decodes value from it.
type Decoder struct {
	r io.Reader
}

// NewDecoder creates a new [Decoder] from an [io.Reader]
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: r,
	}
}

// Decode a value from the decoder.
//
//   - If the value is [Unmarshaler], the corresponding UnmarshalBCS will be called.
//   - If the value is [Enum], it will be special handled for [Enum]
func (d *Decoder) Decode(v any) error {
	reflectValue := reflect.ValueOf(v)
	if reflectValue.Kind() != reflect.Pointer || reflectValue.IsNil() {
		return fmt.Errorf("not a pointer or nil pointer")
	}

	return d.decode(reflectValue)
}

func (d *Decoder) decode(v reflect.Value) error {
	// if v cannot interface, ignore
	if !v.CanInterface() {
		return nil
	}

	if i, isUnmarshaler := v.Interface().(Unmarshaler); isUnmarshaler {
		_, err := i.UnmarshalBCS(d.r)
		return err
	}

	if _, isEnum := v.Interface().(Enum); isEnum {
		switch v.Kind() {
		case reflect.Pointer, reflect.Interface:
			if v.IsNil() {
				return fmt.Errorf("trying to decode into nil pointer/interface")
			}
			return d.decodeEnum(v.Elem())
		default:
			return d.decodeEnum(v)
		}
	}

	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return d.decodeVanilla(v.Elem())

	case reflect.Interface:
		if v.IsNil() {
			return fmt.Errorf("cannot decode into nil interface")
		}
		return d.decode(v.Elem())

	case reflect.Chan, reflect.Func, reflect.Uintptr, reflect.UnsafePointer:
		// silently ignore
		return nil
	default:
		return d.decodeVanilla(v)
	}
}

func (d *Decoder) decodeVanilla(v reflect.Value) error {
	kind := v.Kind()

	if !v.CanSet() {
		return fmt.Errorf("cannot change value of kind %s", kind.String())
	}

	switch v.Kind() {
	case reflect.Bool:
		t, err := d.readByte()
		if err != nil {
			return nil
		}

		if t == 0 {
			v.SetBool(false)
		} else {
			v.SetBool(true)
		}

		return nil

	case reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, // ints
		reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64: // uints
		return binary.Read(d.r, binary.LittleEndian, v.Addr)

	case reflect.Struct:
		return d.decodeStruct(v)

	case reflect.Slice:
		return d.decodeSlice(v)

	case reflect.Array:
		return d.decodeArray(v)

	case reflect.String:
		return d.decodeString(v)

	default:
		return fmt.Errorf("unsupported vanilla decoding type: %s", kind.String())
	}
}

func (d *Decoder) decodeString(v reflect.Value) error {
	size, _, err := ULEB128Decode[int](d.r)
	if err != nil {
		return err
	}

	tmp := make([]byte, size, size)

	read, err := d.r.Read(tmp)
	if err != nil {
		return err
	}

	if size != read {
		return fmt.Errorf("wrong number of bytes read for []byte, want: %d, got %d", size, read)
	}

	v.Set(reflect.ValueOf(string(tmp)))

	return nil
}

func (d *Decoder) readByte() (byte, error) {
	b := make([]byte, 1, 1)
	n, err := d.r.Read(b)
	if err != nil {
		return 0, err
	}
	if n == 0 {
		return 0, fmt.Errorf("EOF")
	}

	return b[0], nil
}

func (d *Decoder) decodeStruct(v reflect.Value) error {
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
		// ignored
		if tag&tagValue_Ignore != 0 {
			continue
		}
		// optional
		if tag&tagValue_Optional != 0 {
			isOptional, err := d.readByte()
			if err != nil {
				return err
			}
			if isOptional == 0 {
				field.Set(reflect.Zero(v.Type()))
			} else {
				field.Set(reflect.New(field.Type().Elem()))
				err := d.decode(field.Elem())
				if err != nil {
					return err
				}
			}
		}

		if err := d.decode(field); err != nil {
			return err
		}
	}

	return nil
}

func (d *Decoder) decodeEnum(v reflect.Value) error {
	if v.Kind() != reflect.Struct {
		return fmt.Errorf("only support struct for Enum, got %s", v.Kind().String())
	}
	enumId, _, err := ULEB128Decode[int](d.r)
	if err != nil {
		return err
	}

	field := v.Field(enumId)

	return d.decode(field)
}

func (d *Decoder) decodeByteSlice(v reflect.Value) error {
	size, _, err := ULEB128Decode[int](d.r)
	if err != nil {
		return err
	}

	tmp := make([]byte, size, size)

	read, err := d.r.Read(tmp)
	if err != nil {
		return err
	}

	if size != read {
		return fmt.Errorf("wrong number of bytes read for []byte, want: %d, got %d", size, read)
	}

	v.Set(reflect.ValueOf(tmp))

	return nil
}

func (d *Decoder) decodeArray(v reflect.Value) error {
	size := v.Len()
	t := v.Type()

	for i := 0; i < size; i++ {
		v.Index(i).Set(reflect.New(t.Elem()))
		if err := d.decode(v.Index(i)); err != nil {
			return err
		}
	}

	return nil
}

func (d *Decoder) decodeSlice(v reflect.Value) error {
	size, _, err := ULEB128Decode[int](d.r)
	if err != nil {
		return err
	}

	t := v.Type()
	tmp := reflect.MakeSlice(t, 0, size)
	for i := 0; i < size; i++ {
		ind := reflect.New(t.Elem())
		if err := d.decode(ind); err != nil {
			return err
		}
		tmp = reflect.Append(tmp, ind)
	}

	v.Set(tmp)

	return nil
}
