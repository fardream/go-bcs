package bcs

import (
	"encoding/binary"
	"fmt"
	"io"
)

const MaxUleb128 = uint64(1<<32 - 1)

// MaxUleb128Length is the max possible number of bytes for an ULEB128 encoded integer.
// All integers must fit in a u32, so the length is 10.
const MaxUleb128Length = 5

// ULEB128SupportedTypes is a contraint interface that limits the input to
// [ULEB128Encode] and [ULEB128Decode] to signed and unsigned integers.
type ULEB128SupportedTypes interface {
	~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uint | ~int8 | ~int16 | ~int32 | ~int64 | ~int
}

// ULEB128Encode converts an integer into []byte (see [wikipedia] and [bcs])
//
// This reuses [binary.PutUvarint] in standard library.
//
// [wikipedia]: https://en.wikipedia.org/wiki/LEB128
// [bcs]: https://github.com/diem/bcs#uleb128-encoded-integers
func ULEB128Encode[T ULEB128SupportedTypes](input T) ([]byte, error) {
	v := uint64(input)

	if v > MaxUleb128 {
		return nil, fmt.Errorf("input %d was larger than the max allowed ULEB128", v)
	}

	result := make([]byte, MaxUleb128Length)
	i := binary.PutUvarint(result, v)
	return result[:i], nil
}

// ULEB128Decode decodes [io.Reader] into an integer, returns the resulted value, the number of byte read, and a possible error.
//
// [binary.ReadUvarint] is not used here because
//   - it doesn't support returning the number of bytes read.
//   - it accepts only [io.ByteReader], but the recommended way of creating one from [bufio.NewReader] will read more than 1 byte at the
//     to fill the buffer.
func ULEB128Decode[T ULEB128SupportedTypes](r io.Reader) (T, int, error) {
	buf := make([]byte, 1)
	var v, shift T
	var n int
	for n < MaxUleb128Length {
		i, err := r.Read(buf)
		if i == 0 {
			return 0, n, fmt.Errorf("zero read in. possible EOF")
		}
		if err != nil {
			return 0, n, err
		}
		n += i

		d := T(buf[0])
		ld := d & 127

		if (ld<<shift)>>shift != ld {
			return 0, n, fmt.Errorf("overflow at index %d: %v", n-1, ld)
		}

		ld <<= shift
		v = ld + v
		if v < ld {
			return 0, n, fmt.Errorf("overflow after adding index %d: %v %v", n-1, ld, v)
		}

		if uint64(v) > MaxUleb128 {
			return 0, n, fmt.Errorf("overflow at index %d: %v, value does not fit in u32", n-1, ld)
		}

		if d <= 127 {
			if shift > 0 && d == 0 {
				return 0, n, fmt.Errorf("ULEB128 encoding was not minimal in size")
			}

			if uint64(v) > MaxUleb128 {
				return 0, n, fmt.Errorf("overflow at index %d: %v, value does not fit in u32", n-1, ld)
			}

			return v, n, nil
		}

		shift += 7
	}

	return 0, n, fmt.Errorf("failed to find most significant bytes after reading %d bytes", n)
}
