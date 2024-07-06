package bcs_test

import (
	"slices"
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
	{input: []uint16{1, 2}, expected: []byte{2, 1, 0, 2, 0}},
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
	{input: utf8Str, expected: utf8Encoded},
}

func TestMarshal_basicTypes(t *testing.T) {
	for _, aCase := range basicMarshalTests {
		r, err := bcs.Marshal(aCase.input)
		if err != nil {
			t.Errorf("failed to marshal %v: %v", aCase.input, err)
		}
		if !slices.Equal(r, aCase.expected) {
			t.Errorf("want: %v\ngot:  %v\n", aCase.expected, r)
		}
	}
}

type MyStruct struct {
	Boolean bool
	Bytes   []byte
	Label   string
}

type Wrapper struct {
	Inner  MyStruct
	String string
}

type WrapperWithWrongOptional struct {
	Inner MyStruct
	Outer string `bcs:"optional"`
}

type WrapperWithOptional struct {
	Inner MyStruct
	Outer *string `bcs:"optional"`
}

// struct from [bcs repo]
//
// [bcs repo]: https://github.com/diem/bcs
func TestMarshal_struct(t *testing.T) {
	s := &MyStruct{
		Boolean: true,
		Bytes:   []byte{0xC0, 0xDE},
		Label:   "a",
	}

	sBytes, err := bcs.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	sBytesExpected := []byte{1, 2, 0xC0, 0xDE, 1, 97}
	if !slices.Equal(sBytes, sBytesExpected) {
		t.Fatalf("want: %v\ngot:  %v\n", sBytesExpected, sBytes)
	}

	w := Wrapper{
		Inner:  *s,
		String: "b",
	}

	wBytes, err := bcs.Marshal(w)
	if err != nil {
		t.Fatal(err)
	}
	wBytesExpected := append(sBytesExpected, 1, 98)
	if !slices.Equal(wBytes, wBytesExpected) {
		t.Fatalf("want: %v\ngot:  %v\n", wBytesExpected, wBytes)
	}
}

func TestMarshal_optional(t *testing.T) {
	if _, err := bcs.Marshal(WrapperWithWrongOptional{}); err == nil {
		t.Fatalf("optional should be pointer or interface")
	} else {
		t.Log(err.Error())
	}
	optionalUnset := WrapperWithOptional{
		Inner: MyStruct{
			Boolean: true,
			Bytes:   []byte{0xC0, 0xDE},
			Label:   "a",
		},
	}
	optionalUnsetBytes, err := bcs.Marshal(optionalUnset)
	if err != nil {
		t.Error(err)
	}
	optionalUnsetExpected := []byte{1, 2, 0xC0, 0xDE, 1, 97, 0}
	if !slices.Equal(optionalUnsetBytes, optionalUnsetExpected) {
		t.Errorf("want: %v\ngot:  %v\n", optionalUnsetExpected, optionalUnsetBytes)
	}

	optionalSet := optionalUnset
	s := "123"
	optionalSet.Outer = &s
	optionalSetBytes, err := bcs.Marshal(optionalSet)
	if err != nil {
		t.Error(err)
	}
	optionalSetExpected := []byte{1, 2, 0xC0, 0xDE, 1, 97, 1, 3, 49, 50, 51}
	if !slices.Equal(optionalSetBytes, optionalSetExpected) {
		t.Errorf("want: %v\ngot:  %v\n", optionalSetExpected, optionalSetBytes)
	}
}

func TestMarshal_option(t *testing.T) {
	t.Run("some", func(t *testing.T) {
		var p0, p1 bcs.Option[[]byte]
		input := []byte{0xC0, 0xDE}
		inputExpected := []byte{1, 2, 0xC0, 0xDE}
		p0.Some = input
		b, err := bcs.Marshal(&p0)
		if err != nil {
			t.Error(err)
		}
		if !slices.Equal(b, inputExpected) {
			t.Errorf("want: %v\ngot:  %v\n", inputExpected, b)
		}
		_, err = bcs.Unmarshal(b, &p1)
		if err != nil {
			t.Error(err)
		}
		if !slices.Equal(input, p1.Some) {
			t.Errorf("want: %v\ngot:  %v\n", input, p1.Some)
		}
	})
	t.Run("none", func(t *testing.T) {
		var p0, p1 bcs.Option[[]byte]
		inputExpected := []byte{0}
		p0.None = true
		b, err := bcs.Marshal(&p0)
		if err != nil {
			t.Error(err)
		}
		if !slices.Equal(b, inputExpected) {
			t.Errorf("want: %v\ngot:  %v\n", inputExpected, b)
		}
		_, err = bcs.Unmarshal(b, &p1)
		if err != nil {
			t.Error(err)
		}
		if !p1.None {
			t.Errorf("None field should be true")
		}
	})
}
