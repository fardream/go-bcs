package bcs_test

import (
	"fmt"
	"testing"

	"github.com/fardream/go-bcs/bcs"
)

func runVanillaCaseTest[T bool | uint8 | int8 | int16 | uint16 | int32 | uint32 | int64 | uint64 | string](v any, exp []byte) error {
	x, ok := v.(T)
	if !ok {
		return runVanillaSliceCaseTest[T](v, exp)
	}

	nv := new(T)
	n, err := bcs.Unmarshal(exp, nv)
	if err != nil {
		return err
	}
	if *nv != x {
		return fmt.Errorf("want value: %v, got value: %v", x, *nv)
	}
	if n != len(exp) {
		return fmt.Errorf("want length: %d, got length: %d", len(exp), n)
	}

	return nil
}

func runVanillaSliceCaseTest[T bool | uint8 | int8 | int16 | uint16 | int32 | uint32 | int64 | uint64 | string](v any, exp []byte) error {
	x, ok := v.([]T)
	if !ok {
		return nil
	}

	nv := make([]T, 0)
	n, err := bcs.Unmarshal(exp, &nv)
	if err != nil {
		return err
	}

	if len(nv) != len(x) {
		return fmt.Errorf("want length: %d, got %d", len(x), len(nv))
	}

	if n != len(exp) {
		return fmt.Errorf("want parsed length: %d, got %d", len(exp), n)
	}

	for i := 0; i < len(x); i++ {
		if nv[i] != x[i] {
			return fmt.Errorf("diff at %d %v %v", i, nv[i], x[i])
		}
	}

	return nil
}

func TestUnmarshal_BasicTypes(t *testing.T) {
	for _, aCase := range basicMarshalTests {
		if err := runVanillaCaseTest[bool](aCase.input, aCase.expected); err != nil {
			t.Fatal(err)
		}
		if err := runVanillaCaseTest[uint8](aCase.input, aCase.expected); err != nil {
			t.Fatal(err)
		}
		if err := runVanillaCaseTest[int8](aCase.input, aCase.expected); err != nil {
			t.Fatal(err)
		}
		if err := runVanillaCaseTest[uint16](aCase.input, aCase.expected); err != nil {
			t.Fatal(err)
		}
		if err := runVanillaCaseTest[int16](aCase.input, aCase.expected); err != nil {
			t.Fatal(err)
		}
		if err := runVanillaCaseTest[int32](aCase.input, aCase.expected); err != nil {
			t.Fatal(err)
		}
		if err := runVanillaCaseTest[uint32](aCase.input, aCase.expected); err != nil {
			t.Fatal(err)
		}
		if err := runVanillaCaseTest[int64](aCase.input, aCase.expected); err != nil {
			t.Fatal(err)
		}
		if err := runVanillaCaseTest[uint64](aCase.input, aCase.expected); err != nil {
			t.Fatal(err)
		}
		if err := runVanillaCaseTest[string](aCase.input, aCase.expected); err != nil {
			t.Fatal(err)
		}
	}
}

type UnmarshalStruct struct {
	WrapperWithOptional
	StructArray [2]*MyStruct
}

type UnmarshalCase struct {
	v         *UnmarshalStruct
	expected  []byte
	errNotNil bool
}

var unmarshalCases = []*UnmarshalCase{
	{
		v: &UnmarshalStruct{
			WrapperWithOptional: WrapperWithOptional{
				Inner: MyStruct{Bytes: []byte{9, 2}},
				Outer: new(string),
			},
			StructArray: [2]*MyStruct{
				{
					Boolean: true,
					Bytes:   []byte{1, 2, 3},
					Label:   "what",
				},
				{
					Boolean: false,
				},
			},
		},
		errNotNil: false,
		expected:  []byte{0, 2, 9, 2, 0, 1, 0, 1, 3, 1, 2, 3, 4, 119, 104, 97, 116, 0, 0, 0},
	},
}

func TestUnmarshal(t *testing.T) {
	for _, v := range unmarshalCases {
		m, err := bcs.Marshal(v.v)
		if err != nil {
			t.Error(err)
		}
		if !sliceEqual(m, v.expected) {
			t.Errorf("want: %v, got %v", v.expected, m)
		}
		nv := new(UnmarshalStruct)
		n, err := bcs.Unmarshal(v.expected, nv)
		if err != nil {
			t.Error(err)
		}
		if n != len(v.expected) {
			t.Errorf("want parsed length: %d, got: %d", len(v.expected), n)
		}

		nb, err := bcs.Marshal(nv)
		if err != nil {
			t.Fatal(err)
		}

		if !sliceEqual(nb, v.expected) {
			t.Fatalf("want: %v, got: %v", v.expected, nb)
		}
	}
}

func TestEmptyByteSlice(t *testing.T) {
	encoded, err := bcs.Marshal([]byte{})
	if err != nil {
		t.Error(err)
	}
	decoded := new([]byte)
	n, err := bcs.Unmarshal(encoded, decoded)
	if err != nil {
		t.Error(err)
	}
	if n != 1 {
		t.Errorf("want parsed length 1")
	}
}
