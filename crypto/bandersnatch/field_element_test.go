package bandersnatch

import (
	"testing"
)

func TestSanity(t *testing.T) {
	t.Logf("%x %x %x %x", uint64(m_64_0), uint64(m_64_1), uint64(m_64_2), uint64(m_64_3))
	t.Logf("%x", BaseFieldSize)
	if BaseFieldSize.ProbablyPrime(10) == false {
		t.Fatal("Modulus is not prime")
	}
}
