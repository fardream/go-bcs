package bcs_test

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/fardream/go-bcs/bcs"
)

type ULEB128Test struct {
	Input    uint32
	Expected []byte
}

var uleb128Tests = []ULEB128Test{
	{0, []byte{0}},
	{1, []byte{1}},
	{128, []byte{0x80, 1}},
	{16384, []byte{0x80, 0x80, 1}},
	{2097152, []byte{0x80, 0x80, 0x80, 1}},
	{268435456, []byte{0x80, 0x80, 0x80, 0x80, 1}},
	{9487, []byte{0x8f, 0x4a}},
	{uint32(bcs.MaxUleb128), []byte{0xff, 0xff, 0xff, 0xff, 0x0f}},
}

func TestULEB128Encode(t *testing.T) {
	for _, aCase := range uleb128Tests {
		r, err := bcs.ULEB128Encode(aCase.Input)
		if err != nil {
			t.Fatalf("failed to encode: %v", err)
		}
		if !slices.Equal(r, aCase.Expected) {
			t.Errorf("encoding %d to %v, expecting: %v", aCase.Input, r, aCase.Expected)
		}
	}
}

func TestULEB128Decode(t *testing.T) {
	for _, aCase := range uleb128Tests {
		r, n, e := bcs.ULEB128Decode[uint32](bytes.NewReader(aCase.Expected))
		if e != nil {
			t.Fatalf("failed to decode: %v", e)
		}
		if n != len(aCase.Expected) {
			t.Fatalf("didn't consume whole stream: %d", n)
		}
		if r != aCase.Input {
			t.Fatalf("decoded into wrong value: want %d, got: %d", aCase.Input, r)
		}
	}

	for _, aCase := range uleb128Tests[3:] {
		r, n, e := bcs.ULEB128Decode[uint8](bytes.NewReader(aCase.Expected))
		if e == nil {
			t.Fatalf("should overflow: %d %d", r, n)
		} else {
			t.Logf("succeeded overflowing: %v", e)
		}
	}
}

func TestULEB128DecodeNonCanonical(t *testing.T) {
	cases := [][]byte{
		{0x80, 0},
		{0x80, 0x80, 0},
	}

	for _, c := range cases {
		_, _, err := bcs.ULEB128Decode[int](bytes.NewReader(c))
		if !strings.Contains(err.Error(), "ULEB128 encoding was not minimal in size") {
			t.Fatalf("expected failure due to non minimal input, got %v", err)
		}
	}
}

func TestULEB128DecodeTooLarge(t *testing.T) {
	_, _, err := bcs.ULEB128Decode[uint64](bytes.NewReader([]byte{0x80, 0x80, 0x80, 0x80, 0x80, 0x01}))
	if !strings.Contains(err.Error(), "failed to find most significant bytes") {
		t.Fatalf("expected failure due to size, got %v", err)
	}

	_, _, err = bcs.ULEB128Decode[uint64](bytes.NewReader([]byte{0x80, 0x80, 0x80, 0x80, 0x10}))
	if !strings.Contains(err.Error(), "value does not fit in u32") {
		t.Fatalf("expected failure due to size, got %v", err)
	}
}

func TestULEB128EncodeTooLarge(t *testing.T) {
	cases := []uint64{
		2 << 33,
		2 << 36,
		bcs.MaxUleb128 + 1,
	}

	for _, c := range cases {
		_, err := bcs.ULEB128Encode(c)
		if !strings.Contains(err.Error(), "larger than the max allowed ULEB128") {
			t.Fatalf("expected failure due to size, got %v", err)
		}
	}
}
