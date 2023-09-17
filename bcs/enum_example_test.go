package bcs_test

import (
	"fmt"

	"github.com/fardream/go-bcs/bcs"
)

type AnotherStruct struct {
	S string
}

type EnumExample struct {
	V0 *uint8
	V1 *uint16 `bcs:"-"`
	V2 *uint32
	v3 uint8
	V4 *AnotherStruct
}

// IsBcsEnum tells this is an enum
func (e EnumExample) IsBcsEnum() {}

func ExampleEnum() {
	// an enum
	v0 := &EnumExample{
		V0: new(uint8),
	}
	*v0.V0 = 42

	v0m, err := bcs.Marshal(v0)
	if err != nil {
		panic(err)
	}
	fmt.Println("v0:", v0m)

	// first value will be picked up
	v1 := &EnumExample{
		V0: new(uint8),
		V2: new(uint32),
	}
	*v1.V0 = 0
	*v1.V2 = 10
	v1m, err := bcs.Marshal(v1)
	if err != nil {
		panic(err)
	}
	fmt.Println("v1:", v1m)

	// enum must be set
	v2 := &EnumExample{}

	_, err = bcs.Marshal(v2)
	if err == nil {
		panic(fmt.Errorf("unset enum should error out"))
	}

	// setting V1, which is ignored, and v3, which is unexported, should be ignored
	v3 := &EnumExample{
		V1: new(uint16),
		v3: 90,
		V4: &AnotherStruct{
			S: "abc",
		},
	}

	v3m, err := bcs.Marshal(v3)
	if err != nil {
		panic(err)
	}
	// print [4 3 97 98 99], which is 4 (enum int), 3 (size of the bytes of the string), 97=a, 98=b, 99=c
	fmt.Println("v3:", v3m)
	// Output: v0: [0 42]
	// v1: [0 0]
	// v3: [4 3 97 98 99]
}
