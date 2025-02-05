package bcs_test

import (
	"errors"
	"io"
	"slices"
	"testing"

	"github.com/fardream/go-bcs/bcs"
	"github.com/stretchr/testify/require"
)

type NestedEnum struct {
	V0 *EnumExample
	V1 *uint8
}

func (e NestedEnum) IsBcsEnum() {}

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

func TestEnumInvalid_Unmarshal(t *testing.T) {
	cases := [][]byte{
		{5, 42},
	}

	for _, v := range cases {
		e := &EnumExample{}

		require.NotPanics(t, func() {
			n, err := bcs.Unmarshal(v, e)
			require.Error(t, err)
			require.Equal(t, 1, n)
		})
	}
}

func TestNestedEnum_Unmarshal(t *testing.T) {
	cases := [][]byte{
		{0, 0, 42},
		{1, 0},
		{1, 42},
		{0, 4, 3, 97, 98, 99},
	}

	for _, v := range cases {
		e := &NestedEnum{}

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

type Foo struct {
	A CustomUnmarshal
}

type CustomUnmarshal struct {
	IsInt int
}

func (u *CustomUnmarshal) UnmarshalBCS(r io.Reader) (int, error) {
	b := make([]byte, 1)
	if n, err := io.ReadFull(r, b); err != nil {
		return n, err
	}

	if b[0] == 0 {
		u.IsInt = 0
	} else if b[0] == 1 {
		u.IsInt = 1
	} else {
		return len(b), errors.New("invalid")
	}

	return len(b), nil
}

func (u CustomUnmarshal) MarshalBCS() ([]byte, error) {
	if u.IsInt == 0 {
		return []byte{0}, nil
	} else if u.IsInt == 1 {
		return []byte{1}, nil
	} else {
		return nil, errors.New("invalid")
	}
}

func TestNestedCustom_Unmarshal(t *testing.T) {
	cases := [][]byte{
		{0},
		{1},
	}

	for _, v := range cases {
		e := &Foo{}

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
