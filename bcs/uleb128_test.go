package bcs_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/fardream/go-bcs/bcs"
)

type ULEB128Test struct {
	Input    uint32
	Expected []byte
}

var tests = []ULEB128Test{
	{0, []byte{0}},
	{1, []byte{1}},
	{128, []byte{0x80, 1}},
	{16384, []byte{0x80, 0x80, 1}},
	{2097152, []byte{0x80, 0x80, 0x80, 1}},
	{268435456, []byte{0x80, 0x80, 0x80, 0x80, 1}},
	{9487, []byte{0x8f, 0x4a}},
}

func TestULEB128Encode(t *testing.T) {
	for _, aCase := range tests {
		r := bcs.ULEB128Encode(aCase.Input)
		if !cmp.Equal(r, aCase.Expected) {
			t.Errorf("encoding %d to %v, expecting: %v", aCase.Input, r, aCase.Expected)
		}
	}
}

func TestULEB128Decode(t *testing.T) {
	for _, aCase := range tests {
		r, n, e := bcs.ULEB128Decode[uint32](aCase.Expected)
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

	for _, aCase := range tests[3:] {
		r, n, e := bcs.ULEB128Decode[uint8](aCase.Expected)
		if e == nil {
			t.Fatalf("should overflow: %d %d", r, n)
		} else {
			t.Logf("succeeded overflowing: %v", e)
		}
	}
}
