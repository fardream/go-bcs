package bcs

import (
	"bytes"
	"container/list"
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
)

// maxChunkSize is the maximum size to allocate at once, limiting DoS attacks
const maxChunkSize = 1024 * 1024 // 1MB chunks

// Unmarshal unmarshals the bcs serialized data into v.
//
// Refer to notes in [Marshal] for details how data serialized/deserialized.
//
// During the unmarshalling process
//  1. if [Unmarshaler], use "UnmarshalBCS" method.
//  2. if not [Unmarshaler] but [Enum], use the specialization for [Enum].
//  3. otherwise standard process.
func Unmarshal(data []byte, v any) (int, error) {
	return NewDecoder(bytes.NewReader(data)).Decode(v)
}

// UnmarshalAll is like [Unmarshal], but will additionally error if the input
// bytes are not completely consumed by the call to [Unmarshal].
//
// This is useful for ensuring that a particular set of bytes *completely* represents
// the data it is decoded to (i.e. no bytes are left over after decoding).
func UnmarshalAll(data []byte, v any) error {
	if n, err := NewDecoder(bytes.NewReader(data)).Decode(v); err != nil {
		return err
	} else if n != len(data) {
		return fmt.Errorf("did not unmarshal all bytes, got %d expected %d", n, len(data))
	}

	return nil
}

// Decoder takes an [io.Reader] and decodes value from it.
type Decoder struct {
	reader     io.Reader
	byteBuffer [1]byte
}

// NewDecoder creates a new [Decoder] from an [io.Reader]
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		reader: r,
	}
}

// DecodeWithSize decodes a value from the decoder, and returns the number of bytes it consumed from the decoder.
//
//   - If the value is [Unmarshaler], the corresponding UnmarshalBCS will be called.
//   - If the value is [Enum], it will be special handled for [Enum]
func (d *Decoder) Decode(v any) (int, error) {
	reflectValue := reflect.ValueOf(v)
	if reflectValue.Kind() != reflect.Pointer || reflectValue.IsNil() {
		return 0, fmt.Errorf("not a pointer or nil pointer")
	}

	return d.decode(reflectValue)
}

// decode is the main lifter, it first checks if a value can be [reflect.Value.CanInterface],
// then checks if the value implements [Unmarshaler] or [Enum], and then switch on the kind of the value:
// - pointer, create a new one and decode into its element.
// - interface, decode into element.
// - function, channel, unsafe pointers, ignore
// - otherwise call [decodeVanilla].
func (d *Decoder) decode(v reflect.Value) (int, error) {
	// if v cannot interface, ignore
	if !v.CanInterface() {
		return 0, nil
	}

	vAddr := v
	if v.Kind() != reflect.Pointer && v.CanAddr() {
		vAddr = v.Addr()
	}

	// Unmarshaler
	if i, isUnmarshaler := vAddr.Interface().(Unmarshaler); isUnmarshaler {
		return i.UnmarshalBCS(d.reader)
	}

	// Enum
	if _, isEnum := v.Interface().(Enum); isEnum {
		switch v.Kind() {
		case reflect.Pointer:
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			return d.decodeEnum(v.Elem())
		case reflect.Interface:
			if v.IsNil() {
				return 0, fmt.Errorf("trying to decode into nil pointer/interface")
			}
			return d.decodeEnum(v.Elem())
		default:
			return d.decodeEnum(v)
		}
	}

	// switch kind
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return d.decode(v.Elem())

	case reflect.Interface:
		if v.IsNil() {
			return 0, fmt.Errorf("cannot decode into nil interface")
		}
		return d.decode(v.Elem())

	case reflect.Chan, reflect.Func, reflect.Uintptr, reflect.UnsafePointer:
		// silently ignore
		return 0, nil
	default:
		return d.decodeVanilla(v)
	}
}

// decodeVanilla decodes bool, ints, slice, struct, array, and string.
func (d *Decoder) decodeVanilla(v reflect.Value) (int, error) {
	kind := v.Kind()
	if !v.CanSet() {
		return 0, fmt.Errorf("cannot change value of kind %s", kind.String())
	}

	switch kind {
	case reflect.Bool:
		t, n, err := d.readByte()
		if err != nil {
			return n, err
		}

		if t == 0 {
			v.SetBool(false)
		} else {
			v.SetBool(true)
		}

		return n, nil

	case reflect.Int8, reflect.Uint8:
		return 1, binary.Read(d.reader, binary.LittleEndian, v.Addr().Interface())
	case reflect.Int16, reflect.Uint16:
		return 2, binary.Read(d.reader, binary.LittleEndian, v.Addr().Interface())
	case reflect.Int32, reflect.Uint32:
		return 4, binary.Read(d.reader, binary.LittleEndian, v.Addr().Interface())
	case reflect.Int64, reflect.Uint64:
		return 8, binary.Read(d.reader, binary.LittleEndian, v.Addr().Interface())

	case reflect.Struct:
		return d.decodeStruct(v)

	case reflect.Slice:
		sliceType := v.Type().Elem()
		if sliceType.Kind() == reflect.Uint8 {
			return d.decodeByteSlice(v)
		}

		return d.decodeSlice(v)

	case reflect.Array:
		return d.decodeArray(v)

	case reflect.String:
		return d.decodeString(v)

	default:
		return 0, fmt.Errorf("unsupported vanilla decoding type: %s", kind.String())
	}
}

// decodeString
func (d *Decoder) decodeString(v reflect.Value) (int, error) {
	size, n, err := ULEB128Decode[int](d.reader)
	if err != nil {
		return n, err
	}

	if size == 0 {
		v.SetString("")
		return n, nil
	}

	// Allocate in chunks. `tmp` is guaranteed to allocate at most 2^32 / (1024*1024) = 4096 elements.
	// This also avoids copying the data many times
	tmp := make([][]byte, 0, (size+maxChunkSize-1)/maxChunkSize)
	totalRead := 0

	for totalRead < size {
		// Determine how much to read in this iteration
		remaining := size - totalRead
		chunkSize := min(remaining, maxChunkSize)

		tmp = append(tmp, make([]byte, chunkSize))

		// Read the chunk
		read, err := d.reader.Read(tmp[len(tmp)-1])
		totalRead += read

		if err != nil {
			return n + totalRead, err
		}
	}
	// If there's only one chunk, no need to copy the data
	if len(tmp) == 1 {
		v.SetString(string(tmp[0]))
	} else {
		v.SetString(string(bytes.Join(tmp, nil)))
	}

	return n + totalRead, nil
}

// readByte reads one byte from the input, error if no byte is read.
func (d *Decoder) readByte() (byte, int, error) {
	b := d.byteBuffer[:]
	n, err := d.reader.Read(b)
	if err != nil {
		return 0, n, err
	}
	if n == 0 {
		return 0, n, io.ErrUnexpectedEOF
	}

	return b[0], n, nil
}

func (d *Decoder) decodeStruct(v reflect.Value) (int, error) {
	t := v.Type()

	var n int

fieldLoop:
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if !field.CanInterface() {
			continue fieldLoop
		}
		tag, err := parseTagValue(t.Field(i).Tag.Get(tagName))
		if err != nil {
			return n, err
		}

		switch {
		case tag.isIgnored(): // ignored
			continue fieldLoop
		case tag.isOptional(): // optional
			isOptional, k, err := d.readByte()
			n += k
			if err != nil {
				return n, err
			}
			if isOptional == 0 {
				field.Set(reflect.Zero(field.Type()))
			} else {
				field.Set(reflect.New(field.Type().Elem()))
				k, err := d.decode(field.Elem())
				n += k
				if err != nil {
					return n, err
				}
			}
		default:
			k, err := d.decode(field)
			n += k
			if err != nil {
				return n, err
			}
		}
	}

	return n, nil
}

func (d *Decoder) decodeEnum(v reflect.Value) (int, error) {
	if v.Kind() != reflect.Struct {
		return 0, fmt.Errorf("only support struct for Enum, got %s", v.Kind().String())
	}
	enumId, n, err := ULEB128Decode[int](d.reader)
	if err != nil {
		return n, err
	}

	if enumId >= v.NumField() {
		return n, fmt.Errorf("enum field %d is out of range", enumId)
	}

	field := v.Field(enumId)

	// Initialize nil pointer fields before decoding
	if field.Kind() == reflect.Ptr && field.IsNil() {
		field.Set(reflect.New(field.Type().Elem()))
	}

	k, err := d.decode(field)
	n += k

	return n, err
}

func (d *Decoder) decodeByteSlice(v reflect.Value) (int, error) {
	size, n, err := ULEB128Decode[int](d.reader)
	if err != nil {
		return n, err
	}

	if size == 0 {
		return n, nil
	}

	// Allocate in chunks. `tmp` is guaranteed to allocate at most 2^32 / (1024*1024) = 4096 elements.
	// This also avoids copying the data many times
	tmp := make([][]byte, 0, (size+maxChunkSize-1)/maxChunkSize)
	totalRead := 0

	for totalRead < size {
		// Determine how much to read in this iteration
		remaining := size - totalRead
		chunkSize := min(remaining, maxChunkSize)

		tmp = append(tmp, make([]byte, chunkSize))

		// Read the chunk
		read, err := d.reader.Read(tmp[len(tmp)-1])
		totalRead += read

		if err != nil {
			return n + totalRead, err
		}
	}

	// If there's only one chunk, no need to do an extra copy inside Join
	if len(tmp) == 1 {
		v.SetBytes(tmp[0])
	} else {
		v.SetBytes(bytes.Join(tmp, nil))
	}

	return n + totalRead, nil
}

func (d *Decoder) decodeArray(v reflect.Value) (int, error) {
	size := v.Len()
	t := v.Type()
	elementType := t.Elem()

	var n int
	if elementType.Kind() == reflect.Pointer {
		for i := 0; i < size; i++ {
			idx := reflect.New(elementType.Elem())
			k, err := d.decode(idx.Elem())
			n += k
			if err != nil {
				return n, err
			}
			v.Index(i).Set(idx)
		}
	} else {
		for i := 0; i < size; i++ {
			idx := reflect.New(elementType)
			k, err := d.decode(idx.Elem())
			n += k
			if err != nil {
				return n, err
			}
			v.Index(i).Set(idx.Elem())
		}
	}

	return n, nil
}

func (d *Decoder) decodeSlice(v reflect.Value) (int, error) {
	// get the length of the slice.
	size, n, err := ULEB128Decode[int](d.reader)
	if err != nil {
		return n, err
	}

	// element type of the slice
	elementType := v.Type().Elem()

	// Use a linked list to hold elements
	// This avoids pre-allocating very large buffer when elements are large
	// or the number of elements is large
	l := list.New()

	if elementType.Kind() == reflect.Pointer {
		for i := 0; i < size; i++ {
			ind := reflect.New(elementType.Elem())
			k, err := d.decode(ind.Elem())
			n += k
			if err != nil {
				return n, err
			}
			l.PushBack(ind)
		}
	} else {
		for i := 0; i < size; i++ {
			ind := reflect.New(elementType)
			k, err := d.decode(ind.Elem())
			n += k
			if err != nil {
				return n, err
			}
			l.PushBack(ind)
		}
	}

	// Now it is okay to allocate a slice of full size
	// since we would have returned early with EoF if the
	// input was smaller than the declare size
	v.Set(reflect.MakeSlice(v.Type(), size, size))
	for i, e := 0, l.Front(); e != nil; i, e = i+1, e.Next() {
		if elementType.Kind() == reflect.Pointer {
			v.Index(i).Set(e.Value.(reflect.Value))
		} else {
			v.Index(i).Set(e.Value.(reflect.Value).Elem())
		}
	}

	return n, nil
}
