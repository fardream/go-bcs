package bcs

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
)

// Uint128 is like `u128` in move.
type Uint128 struct {
	lo uint64
	hi uint64
}

var (
	_ json.Marshaler   = (*Uint128)(nil)
	_ json.Unmarshaler = (*Uint128)(nil)
	_ Marshaler        = (*Uint128)(nil)
)

func (i Uint128) Big() *big.Int {
	loBig := NewBigIntFromUint64(i.lo)
	hiBig := NewBigIntFromUint64(i.hi)
	hiBig = hiBig.Lsh(hiBig, 64)

	return hiBig.Add(hiBig, loBig)
}

func (i Uint128) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.Big().String())
}

var maxU128 = (&big.Int{}).Lsh(big.NewInt(1), 128)

func checkUint128(bigI *big.Int) error {
	if bigI.Sign() < 0 {
		return fmt.Errorf("%s is negative", bigI.String())
	}

	if bigI.Cmp(maxU128) >= 0 {
		return fmt.Errorf("%s is greater than Max Uint 128", bigI.String())
	}

	return nil
}

func (i *Uint128) SetBigInt(bigI *big.Int) error {
	if err := checkUint128(bigI); err != nil {
		return err
	}

	r := make([]byte, 0, 16)
	bs := bigI.Bytes()
	for i := 0; i+len(bs) < 16; i++ {
		r = append(r, 0)
	}
	r = append(r, bs...)

	hi := binary.BigEndian.Uint64(r[0:8])
	lo := binary.BigEndian.Uint64(r[8:16])

	i.hi = hi
	i.lo = lo

	return nil
}

func (i *Uint128) UnmarshalText(data []byte) error {
	bigI := &big.Int{}
	_, ok := bigI.SetString(string(data), 10)
	if !ok {
		return fmt.Errorf("failed to parse %s as an integer", string(data))
	}

	return i.SetBigInt(bigI)
}

func (i *Uint128) UnmarshalJSON(data []byte) error {
	var dataStr string
	if err := json.Unmarshal(data, &dataStr); err != nil {
		return err
	}

	bigI := &big.Int{}
	_, ok := bigI.SetString(dataStr, 10)
	if !ok {
		return fmt.Errorf("failed to parse %s as an integer", dataStr)
	}

	return i.SetBigInt(bigI)
}

func NewUint128FromBigInt(bigI *big.Int) (*Uint128, error) {
	i := &Uint128{}

	if err := i.SetBigInt(bigI); err != nil {
		return nil, err
	}

	return i, nil
}

func NewUint128(s string) (*Uint128, error) {
	r := &big.Int{}
	r, ok := r.SetString(s, 10)
	if !ok {
		return nil, fmt.Errorf("failed to parse %s as an integer", s)
	}

	return NewUint128FromBigInt(r)
}

func (i Uint128) MarshalBCS() ([]byte, error) {
	r := make([]byte, 16)

	binary.LittleEndian.PutUint64(r, i.lo)
	binary.LittleEndian.PutUint64(r[8:], i.hi)

	return r, nil
}

func (i *Uint128) Cmp(j *Uint128) int {
	switch {
	case i.hi > j.hi || (i.hi == j.hi && i.lo > j.lo):
		return 1
	case i.hi == j.hi && i.lo == j.lo:
		return 0
	default:
		return -1
	}
}

// 63 ones
const ones63 uint64 = (1 << 63) - 1

// 1 << 63
var oneLsh63 = big.NewInt(0).Lsh(big.NewInt(1), 63)

func NewBigIntFromUint64(i uint64) *big.Int {
	r := big.NewInt(int64(i & ones63))
	if i > ones63 {
		r = r.Add(r, oneLsh63)
	}
	return r
}

func NewUint128FromUint64(lo, hi uint64) *Uint128 {
	return &Uint128{
		lo: lo,
		hi: hi,
	}
}

func (u Uint128) String() string {
	return u.Big().String()
}
