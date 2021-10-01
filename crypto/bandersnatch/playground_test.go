package bandersnatch

import (
	"math/big"
	"testing"
)

func TestPlayground(t *testing.T) {
	x := big.NewInt(1)
	x.Lsh(x, 256)
	x.Mod(x, BaseFieldSize)
	t.Logf("%x", x)
}
