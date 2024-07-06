package bcs_test

import (
	"slices"
	"testing"

	"github.com/fardream/go-bcs/bcs"
)

func TestEnum_Unmarshal(t *testing.T) {
	cases := [][]byte{
		{0, 42},
		{0, 0},
		{4, 3, 97, 98, 99},
	}

	for _, v := range cases {
		e := &EnumExample{}

		n, err := bcs.Unmarshal(v, e)
		if err != nil {
			t.Error(err)
		}

		if n != len(v) {
			t.Errorf("want parsed length: %d, got: %d", len(v), n)
		}

		nb, err := bcs.Marshal(e)
		if err != nil {
			t.Error(err)
		}

		if !slices.Equal(nb, v) {
			t.Errorf("want %v, got %v", v, nb)
		}
	}
}
