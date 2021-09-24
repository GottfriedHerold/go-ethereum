package bandersnatch

import (
	"fmt"
	"math/big"
	"math/rand"
)

// naive implementation. Field elements are represented as (unsigned) big-endian byte slices representing an integer in [0, BaseFieldSize)
// (because big.Int is convertible to/from it, although big.Int's internal representation is little-endian word slices + sign bits)
// Note that we decide not to embed a big.Int, to avoid pointer indirection: This makes assigment, equality and zero-initialization actually work
// at the expense of speed, but this is only really used to test other implementations against for correctness anyway.
type bsFieldElement_8 struct {
	v [32]byte
}

var bsFieldElement_8_zero bsFieldElement_8

var bsFieldElement_8_one bsFieldElement_8 = bsFieldElement_8{v: [32]byte{31: 1}}

func (z *bsFieldElement_8) add(x, y *bsFieldElement_8) {
	var xInt *big.Int = big.NewInt(0).SetBytes(x.v[:])
	var yInt *big.Int = big.NewInt(0).SetBytes(y.v[:])
	xInt.Add(xInt, yInt)
	xInt.Mod(xInt, BaseFieldSize)
	xInt.FillBytes(z.v[:])
}

func (z *bsFieldElement_8) sub(x, y *bsFieldElement_8) {
	var xInt *big.Int = big.NewInt(0).SetBytes(x.v[:])
	var yInt *big.Int = big.NewInt(0).SetBytes(y.v[:])
	xInt.Sub(xInt, yInt)
	xInt.Mod(xInt, BaseFieldSize) // Note that Int.Mod returns elements in [0, BaseFieldSize), even if xInt is negative.
	xInt.FillBytes(z.v[:])
}

func (z *bsFieldElement_8) isZero() bool {
	return z.v == bsFieldElement_8_zero.v
}

func (z *bsFieldElement_8) isOne() bool {
	return z.v == bsFieldElement_8_one.v
}

func (z *bsFieldElement_8) mul(x, y *bsFieldElement_8) {
	var xInt *big.Int = big.NewInt(0).SetBytes(x.v[:])
	var yInt *big.Int = big.NewInt(0).SetBytes(y.v[:])
	xInt.Mul(xInt, yInt)
	xInt.Mod(xInt, BaseFieldSize)
	xInt.FillBytes(z.v[:])
}

func (z *bsFieldElement_8) toInt() *big.Int {
	var xInt *big.Int = big.NewInt(0).SetBytes(z.v[:])
	return xInt
}

func (z *bsFieldElement_8) setInt(v *big.Int) {
	var xInt *big.Int = big.NewInt(0)
	xInt.Mod(v, BaseFieldSize)
	xInt.FillBytes(z.v[:])
}

// generates a random field element. Non crypto-grade randomness. Used for testing only.
func (z *bsFieldElement_8) setRandomUnsafe(rnd *rand.Rand) {
	var xInt *big.Int = big.NewInt(0).Rand(rnd, BaseFieldSize)
	xInt.FillBytes(z.v[:])
}

// useful for debugging
func (z *bsFieldElement_8) Format(s fmt.State, ch rune) {
	var xInt *big.Int = big.NewInt(0).SetBytes(z.v[:])
	xInt.Format(s, ch)
}
