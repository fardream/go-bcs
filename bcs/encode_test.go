package bcs_test

import (
	"testing"

	"github.com/fardream/go-bcs/bcs"
)

type BasicTypeTest struct {
	input    any
	expected []byte
}

const utf8Str = "çå∞≠¢õß∂ƒ∫"

var utf8Encoded = []byte{
	24, 0xc3, 0xa7, 0xc3, 0xa5, 0xe2, 0x88, 0x9e, 0xe2, 0x89, 0xa0, 0xc2,
	0xa2, 0xc3, 0xb5, 0xc3, 0x9f, 0xe2, 0x88, 0x82, 0xc6, 0x92, 0xe2, 0x88, 0xab,
}

// basicMarshalTests from [bcs repo]
//
// [bcs repo]: https://github.com/diem/bcs
var basicMarshalTests = []BasicTypeTest{
	{input: false, expected: []byte{0}},
	{input: true, expected: []byte{1}},
	{input: uint8(1), expected: []byte{1}},
	{input: int8(-1), expected: []byte{0xff}},
	{input: int16(-4660), expected: []byte{0xcc, 0xed}},
	{input: uint16(4660), expected: []byte{0x34, 0x12}},
	{input: int32(-305419896), expected: []byte{0x88, 0xa9, 0xcb, 0xed}},
	{input: uint32(305419896), expected: []byte{0x78, 0x56, 0x34, 0x12}},
	{input: int64(-1311768467750121216), expected: []byte{0x00, 0x11, 0x32, 0x54, 0x87, 0xa9, 0xcb, 0xed}},
	{input: uint64(1311768467750121216), expected: []byte{0x00, 0xef, 0xcd, 0xab, 0x78, 0x56, 0x34, 0x12}},
	{input: []uint16{1, 2}, expected: []byte{2, 1, 0, 2, 0}},
	{input: utf8Str, expected: utf8Encoded},
}

func TestMarshal_basicTypes(t *testing.T) {
	for _, aCase := range basicMarshalTests {
		r, err := bcs.Marshal(aCase.input)
		if err != nil {
			t.Errorf("failed to marshal %v: %v", aCase.input, err)
		}
		if !sliceEqual(r, aCase.expected) {
			t.Errorf("want: %v\ngot:  %v\n", aCase.expected, r)
		}
	}
}
